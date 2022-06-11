package swarm

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var timeLayout = "2006-01-02T15:04:05Z07:00"

// setup sets up a test HTTP server along with a gitlab.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup(t *testing.T) (*http.ServeMux, *httptest.Server, *Client) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)

	// client is the Gitlab client being tested.
	client, err := NewBasicAuthClient("abc", "123", WithBaseURL(server.URL))
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	return mux, server, client
}

// teardown closes the test HTTP server.
func teardown(server *httptest.Server) {
	server.Close()
}

func testURL(t *testing.T, r *http.Request, want string) {
	if got := r.RequestURI; got != want {
		t.Errorf("Request url: %+v, want %s", got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %s, want %s", got, want)
	}
}

func testBody(t *testing.T, r *http.Request, want string) {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r.Body)

	if err != nil {
		t.Fatalf("Failed to Read Body: %v", err)
	}

	if got := buffer.String(); got != want {
		t.Errorf("Request body: %s, want %s", got, want)
	}
}

func testParams(t *testing.T, r *http.Request, want string) {
	if got := r.URL.RawQuery; got != want {
		t.Errorf("Request query: %s, want %s", got, want)
	}
}

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}

func errorOption(*retryablehttp.Request) error {
	return errors.New("RequestOptionFunc returns an error")
}

func TestNewBasicAuthClient(t *testing.T) {
	Convey("test NewBasicAuthClient", t, func() {
		var (
			c, err          = NewBasicAuthClient("", "")
			expectedBaseURL = defaultBaseURL + apiVersionPath
		)
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		So(c.BaseURL().String(), ShouldEqual, expectedBaseURL)
	})
}

func TestCheckResponse(t *testing.T) {
	Convey("test CheckResponse", t, func() {
		var (
			c, err = NewBasicAuthClient("", "")
		)
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)

		req, err := c.NewRequest(http.MethodGet, "test", nil, nil)
		So(err, ShouldBeNil)
		resp := &http.Response{
			Request:    req.Request,
			StatusCode: http.StatusBadRequest,
			Body: ioutil.NopCloser(strings.NewReader(`
		{
			"message": {
				"prop1": [
					"message 1",
					"message 2"
				],
				"prop2":[
					"message 3"
				],
				"embed1": {
					"prop3": [
						"msg 1",
						"msg2"
					]
				},
				"embed2": {
					"prop4": [
						"some msg"
					]
				}
			},
			"error": "message 1"
		}`)),
		}

		errResp := CheckResponse(resp)
		So(errResp, ShouldNotBeNil)

		want := "GET https://myswarm.url/api/v9/test: 400 {error: message 1}, {message: {embed1: {prop3: [msg 1, msg2]}}, {embed2: {prop4: [some msg]}}, {prop1: [message 1, message 2]}, {prop2: [message 3]}}"
		So(errResp.Error(), ShouldEqual, want)
	})
}

func TestRequestWithContext(t *testing.T) {
	Convey("test RequestWithContext", t, func() {
		c, err := NewBasicAuthClient("", "")
		So(err, ShouldBeNil)

		ctx, cancel := context.WithCancel(context.Background())
		req, err := c.NewRequest(http.MethodGet, "test", nil, []RequestOptionFunc{WithContext(ctx)})
		So(err, ShouldBeNil)
		defer cancel()

		So(req.Context(), ShouldEqual, ctx)
	})

}

func TestPathEscape(t *testing.T) {
	Convey("test RequestWithContext", t, func() {
		want := "diaspora%2Fdiaspora"
		got := PathEscape("diaspora/diaspora")
		So(got, ShouldEqual, want)
	})
}
