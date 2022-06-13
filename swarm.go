package swarm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hetiansu5/urlquery"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const (
	defaultBaseURL = "https://myswarm.url/"
	apiVersionPath = "api/v9/"

	headerRateLimit = "RateLimit-Limit"
	headerRateReset = "RateLimit-Reset"
)

// AuthType represents an authentication type within Swarm.
//
// Swarm API docs: https://www.perforce.com/manuals/swarm/Content/Swarm/swarm-apidoc_endpoints.html
type AuthType int

// List of available authentication types.
//
// Swarm API docs: https://www.perforce.com/manuals/swarm/Content/Swarm/swarm-apidoc_endpoints.html
const (
	BasicAuth AuthType = iota
	JobToken
	OAuthToken
	PrivateToken
)

// A Client manages communication with the Swarm API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *retryablehttp.Client

	// disableRetries is used to disable the default retry logic.
	disableRetries bool

	// configureLimiterOnce is used to make sure the limiter is configured exactly
	// once and block all other calls until the initial (one) call is done.
	configureLimiterOnce sync.Once

	// Limiter is used to limit API calls and prevent 429 responses.
	limiter RateLimiter

	// Token type used to make authenticated API calls.
	authType AuthType

	// Base URL for API requests. Defaults to the public Swarm API, but can be
	// set to a domain endpoint to use with a self hosted Swarm server. baseURL
	// should always be specified with a trailing slash.
	baseURL *url.URL

	// Username and password used for basic authentication.
	username, password string

	// Token used to make authenticated API calls.
	token string

	// Protects the token field from concurrent read/write accesses.
	tokenLock sync.RWMutex

	// Services used for talking to different parts of the Swarm API.
	Workflows *WorkflowService
	Projects  *ProjectsService
}

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     int `json:"page" form:"page" binding:"required,gt=0"`         // 页码
	PageSize int `json:"pageSize" form:"pageSize" binding:"required,gt=0"` // 每页大小
}

// RateLimiter describes the interface that all (custom) rate limiters must implement.
type RateLimiter interface {
	Wait(context.Context) error
}

// NewBasicAuthClient returns a new Swarm API client. To use API methods which
// require authentication, provide a valid username and password.
func NewBasicAuthClient(username, password string, options ...ClientOptionFunc) (*Client, error) {
	client, err := newClient(options...)
	if err != nil {
		return nil, err
	}

	client.authType = BasicAuth
	client.username = username
	client.password = password

	return client, nil
}

func newClient(options ...ClientOptionFunc) (*Client, error) {
	c := &Client{}

	// Configure the HTTP client.
	c.client = &retryablehttp.Client{
		Backoff:      c.retryHTTPBackoff,
		CheckRetry:   c.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 400 * time.Millisecond,
		RetryMax:     5,
	}

	// Set the default base URL.
	c.setBaseURL(defaultBaseURL)

	// Apply any given client options.
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(c); err != nil {
			return nil, err
		}
	}

	// Create all the public services.
	c.Projects = &ProjectsService{client: c}
	c.Workflows = &WorkflowService{client: c}
	return c, nil
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry both rate limit (429) and server (>= 500) errors.
func (c *Client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	if !c.disableRetries && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
		return true, nil
	}
	return false, nil
}

// retryHTTPBackoff provides a generic callback for Client.Backoff which
// will pass through all calls based on the status code of the response.
func (c *Client) retryHTTPBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// Use the rate limit backoff function when we are rate limited.
	if resp != nil && resp.StatusCode == 429 {
		return rateLimitBackoff(min, max, attemptNum, resp)
	}

	// Set custom duration's when we experience a service interruption.
	min = 700 * time.Millisecond
	max = 900 * time.Millisecond

	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

// rateLimitBackoff provides a callback for Client.Backoff which will use the
// RateLimit-Reset header to determine the time to wait. We add some jitter
// to prevent a thundering herd.
//
// min and max are mainly used for bounding the jitter that will be added to
// the reset time retrieved from the headers. But if the final wait time is
// less then min, min will be used instead.
func rateLimitBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// rnd is used to generate pseudo-random numbers.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// First create some jitter bounded by the min and max durations.
	jitter := time.Duration(rnd.Float64() * float64(max-min))

	if resp != nil {
		if v := resp.Header.Get(headerRateReset); v != "" {
			if reset, _ := strconv.ParseInt(v, 10, 64); reset > 0 {
				// Only update min if the given time to wait is longer.
				if wait := time.Until(time.Unix(reset, 0)); wait > min {
					min = wait
				}
			}
		}
	}

	return min + jitter
}

// configureLimiter configures the rate limiter.
func (c *Client) configureLimiter(ctx context.Context) error {
	// Set default values for when rate limiting is disabled.
	limit := rate.Inf
	burst := 0

	defer func() {
		// Create a new limiter using the calculated values.
		c.limiter = rate.NewLimiter(limit, burst)
	}()

	// Create a new request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.String(), nil)
	if err != nil {
		return err
	}

	// Make a single request to retrieve the rate limit headers.
	resp, err := c.client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if v := resp.Header.Get(headerRateLimit); v != "" {
		if rateLimit, _ := strconv.ParseFloat(v, 64); rateLimit > 0 {
			// The rate limit is based on requests per minute, so for our limiter to
			// work correctly we divide the limit by 60 to get the limit per second.
			rateLimit /= 60
			// Configure the limit and burst using a split of 2/3 for the limit and
			// 1/3 for the burst. This enables clients to burst 1/3 of the allowed
			// calls before the limiter kicks in. The remaining calls will then be
			// spread out evenly using intervals of time.Second / limit which should
			// prevent hitting the rate limit.
			limit = rate.Limit(rateLimit * 0.66)
			burst = int(rateLimit * 0.33)
		}
	}

	return nil
}

// BaseURL return a copy of the baseURL.
func (c *Client) BaseURL() *url.URL {
	u := *c.baseURL
	return &u
}

// setBaseURL sets the base URL for API requests to a custom endpoint.
func (c *Client) setBaseURL(urlStr string) error {
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(baseURL.Path, apiVersionPath) {
		baseURL.Path += apiVersionPath
	}

	// Update the base URL of the client.
	c.baseURL = baseURL

	return nil
}

// NewRequest creates a new API request. The method expects a relative URL
// path that will be resolved relative to the base URL of the Client.
// Relative URL paths should always be specified without a preceding slash.
// If specified, the value pointed to by body is JSON encoded and included
// as the request body.
func (c *Client) NewRequest(method, path string, opt interface{}, options []RequestOptionFunc) (*retryablehttp.Request, error) {
	u := *c.baseURL
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	// Set the encoded path data
	u.RawPath = c.baseURL.Path + path
	u.Path = c.baseURL.Path + unescaped

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Accept", "application/json")

	var body interface{}
	switch {
	case method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch:
		reqHeaders.Set("Content-Type", "application/x-www-form-urlencoded")
		if opt != nil {
			//encoder := urlquery.NewEncoder(urlquery.WithNeedEmptyValue(true))
			encoder := urlquery.NewEncoder()
			if body, err = encoder.Marshal(opt); err != nil {
				return nil, err
			}
			//fmt.Println(string(body.([]byte)))
			//if q, err := query.Values(opt); err != nil {
			//	return nil, err
			//} else {
			//	encodedStr := q.Encode()
			//	body = []byte(encodedStr)
			//}
		}

	case opt != nil:
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	req, err := retryablehttp.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return req, nil
}

// Response is a Swarm API response. This wraps the standard http.Response
// returned from Swarm and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response
}

// newResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *retryablehttp.Request, v interface{}) (*Response, error) {
	// If not yet configured, try to configure the rate limiter. Fail
	// silently as the limiter will be disabled in case of an error.
	c.configureLimiterOnce.Do(func() { c.configureLimiter(req.Context()) })

	// Wait will block until the limiter can obtain a new token.
	err := c.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	// Set the correct authentication header. If using basic auth, then check
	// if we already have a token and if not first authenticate and get one.
	var basicAuthToken string
	switch c.authType {
	case BasicAuth:
		c.tokenLock.RLock()
		basicAuthToken = c.token
		c.tokenLock.RUnlock()
		if basicAuthToken == "" {
			// If we don't have a token yet, we first need to request one.
			basicAuthToken, err = c.generateBasicToken(req.Context(), basicAuthToken)
			if err != nil {
				return nil, err
			}
		}
		req.Header.Set("Authorization", "Basic "+basicAuthToken)
	case JobToken:
		if values := req.Header.Values("JOB-TOKEN"); len(values) == 0 {
			req.Header.Set("JOB-TOKEN", c.token)
		}
	case OAuthToken:
		if values := req.Header.Values("Authorization"); len(values) == 0 {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	case PrivateToken:
		if values := req.Header.Values("PRIVATE-TOKEN"); len(values) == 0 {
			req.Header.Set("PRIVATE-TOKEN", c.token)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized && c.authType == BasicAuth {
		resp.Body.Close()
		// The token most likely expired, so we need to request a new one and try again.
		if _, err := c.generateBasicToken(req.Context(), basicAuthToken); err != nil {
			return nil, err
		}
		return c.Do(req, v)
	}
	defer resp.Body.Close()

	response := newResponse(resp)

	err = CheckResponse(resp)
	if err != nil {
		// Even though there was an error, we still return the response
		// in case the caller wants to inspect it further.
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return response, err
}

func (c *Client) generateBasicToken(ctx context.Context, token string) (string, error) {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	// Return early if the token was updated while waiting for the lock.
	if c.token != token {
		return c.token, nil
	}

	if len(c.username) == 0 || len(c.password) == 0 {
		return "", errors.New("username or password is empty!")
	}

	rawToken := fmt.Sprintf("%s:%s", c.username, c.password)
	c.token = base64.URLEncoding.EncodeToString([]byte(rawToken))

	return c.token, nil
}

// Helper function to accept and format both the project ID or name as project
// identifier for all API calls.
func parseID(id interface{}) (string, error) {
	switch v := id.(type) {
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("invalid ID type %#v, the ID must be an int or a string", id)
	}
}

// Helper function to escape a project identifier.
func PathEscape(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), ".", "%2E")
}

// An ErrorResponse reports one or more errors caused by an API request.
//
// Swarm API docs:
// https://docs.swarm.com/ce/api/README.html#data-validation-and-error-reporting
type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	path, _ := url.QueryUnescape(e.Response.Request.URL.Path)
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, path)
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Body = data

		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			errorResponse.Message = "failed to parse unknown error format"
		} else {
			errorResponse.Message = parseError(raw)
		}
	}

	return errorResponse
}

// Format:
// {
//     "message": {
//         "<property-name>": [
//             "<error-message>",
//             "<error-message>",
//             ...
//         ],
//         "<embed-entity>": {
//             "<property-name>": [
//                 "<error-message>",
//                 "<error-message>",
//                 ...
//             ],
//         }
//     },
//     "error": "<error-message>"
// }
func parseError(raw interface{}) string {
	switch raw := raw.(type) {
	case string:
		return raw

	case []interface{}:
		var errs []string
		for _, v := range raw {
			errs = append(errs, parseError(v))
		}
		return fmt.Sprintf("[%s]", strings.Join(errs, ", "))

	case map[string]interface{}:
		var errs []string
		for k, v := range raw {
			errs = append(errs, fmt.Sprintf("{%s: %s}", k, parseError(v)))
		}
		sort.Strings(errs)
		return strings.Join(errs, ", ")

	default:
		return fmt.Sprintf("failed to parse unexpected error type: %T", raw)
	}
}
