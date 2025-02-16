package probe

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"github.com/glendsoza/sprobe/status"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func unsetEnv(t testing.TB, key string) {
	if originalValue, ok := os.LookupEnv(key); ok {
		t.Cleanup(func() { os.Setenv(key, originalValue) })
		os.Unsetenv(key)
	}
}

func TestHTTPProbeProxy(t *testing.T) {

	res := "welcome to http probe proxy"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, res)
	}))
	defer server.Close()

	localProxy := server.URL

	t.Setenv("http_proxy", localProxy)
	t.Setenv("HTTP_PROXY", localProxy)
	unsetEnv(t, "no_proxy")
	unsetEnv(t, "NO_PROXY")

	prober := NewHttpProbe(true)

	// take some time to wait server boot
	time.Sleep(2 * time.Second)
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		assert.NoError(t, err)
	}

	_, response, _ := prober.Probe(req, time.Second*3)
	assert.NotEqual(t, res, response)
}

func TestHTTPProbeChecker(t *testing.T) {
	handleReq := func(s int, body string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(s)
			w.Write([]byte(body))
		}
	}

	// Echo handler that returns the contents of request headers in the body
	headerEchoHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		output := ""
		for k, arr := range r.Header {
			for _, v := range arr {
				output += fmt.Sprintf("%s: %s\n", k, v)
			}
		}
		w.Write([]byte(output))
	}

	redirectHandler := func(s int, bad bool) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/new", s)
			} else if bad && r.URL.Path == "/new" {
				http.Error(w, "", http.StatusInternalServerError)
			}
		}
	}

	redirectHandlerWithBody := func(s int, body string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/new", s)
			} else if r.URL.Path == "/new" {
				w.WriteHeader(s)
				w.Write([]byte(body))
			}
		}
	}

	followNonLocalRedirects := true
	prober := NewHttpProbe(followNonLocalRedirects)
	testCases := []struct {
		handler    func(w http.ResponseWriter, r *http.Request)
		reqHeaders http.Header
		health     status.Status
		accBody    string
		notBody    string
	}{
		// The probe will be filled in below.  This is primarily testing that an HTTP GET happens.
		{
			handler: handleReq(http.StatusOK, "ok body"),
			health:  status.Success,
			accBody: "ok body",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"Accept-Encoding": {"gzip"},
			},
			health:  status.Success,
			accBody: "Accept-Encoding: gzip",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"Accept-Encoding": {"foo"},
			},
			health:  status.Success,
			accBody: "Accept-Encoding: foo",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"Accept-Encoding": {""},
			},
			health:  status.Success,
			accBody: "Accept-Encoding: \n",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"X-Muffins-Or-Cupcakes": {"muffins"},
			},
			health:  status.Success,
			accBody: "X-Muffins-Or-Cupcakes: muffins",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"User-Agent": {"foo/1.0"},
			},
			health:  status.Success,
			accBody: "User-Agent: foo/1.0",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"User-Agent": {""},
			},
			health:  status.Success,
			notBody: "User-Agent",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"User-Agent": {"foo/1.0"},
				"Accept":     {"text/html"},
			},
			health:  status.Success,
			accBody: "Accept: text/html",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"User-Agent": {"foo/1.0"},
				"Accept":     {"foo/*"},
			},
			health:  status.Success,
			accBody: "User-Agent: foo/1.0",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"X-Muffins-Or-Cupcakes": {"muffins"},
				"Accept":                {"foo/*"},
			},
			health:  status.Success,
			accBody: "X-Muffins-Or-Cupcakes: muffins",
		},
		{
			handler: headerEchoHandler,
			reqHeaders: http.Header{
				"Accept": {"foo/*"},
			},
			health:  status.Success,
			accBody: "Accept: foo/*",
		},
		{
			handler: handleReq(http.StatusInternalServerError, "fail body"),
			health:  status.Failure,
		},
		{
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(3 * time.Second)
			},
			health: status.Failure,
		},
		{
			handler: redirectHandler(http.StatusMovedPermanently, false), // 301
			health:  status.Success,
		},
		{
			handler: redirectHandler(http.StatusMovedPermanently, true), // 301
			health:  status.Failure,
		},
		{
			handler: redirectHandler(http.StatusFound, false), // 302
			health:  status.Success,
		},
		{
			handler: redirectHandler(http.StatusFound, true), // 302
			health:  status.Failure,
		},
		{
			handler: redirectHandler(http.StatusTemporaryRedirect, false), // 307
			health:  status.Success,
		},
		{
			handler: redirectHandler(http.StatusTemporaryRedirect, true), // 307
			health:  status.Failure,
		},
		{
			handler: redirectHandler(http.StatusPermanentRedirect, false), // 308
			health:  status.Success,
		},
		{
			handler: redirectHandler(http.StatusPermanentRedirect, true), // 308
			health:  status.Failure,
		},
		{
			handler: redirectHandlerWithBody(http.StatusPermanentRedirect, ""), // redirect with empty body
			health:  status.Warning,
			accBody: "Probe terminated redirects, Response body:",
		},
		{
			handler: redirectHandlerWithBody(http.StatusPermanentRedirect, "ok body"), // redirect with body
			health:  status.Warning,
			accBody: "Probe terminated redirects, Response body: ok body",
		},
	}
	for i, test := range testCases {
		t.Run(fmt.Sprintf("case-%2d", i), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.handler(w, r)
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			assert.NoError(t, err)
			_, port, err := net.SplitHostPort(u.Host)
			assert.NoError(t, err)
			_, err = strconv.Atoi(port)
			assert.NoError(t, err)
			req, err := http.NewRequest("GET", server.URL, nil)
			req.Header = test.reqHeaders
			assert.NoError(t, err)
			health, output, err := prober.Probe(req, 1*time.Second)
			if test.health == status.Unknown {
				assert.Error(t, err)
			}
			if test.health != status.Unknown {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.health, health)
			if health != status.Failure && test.health != status.Failure {
				assert.Contains(t, output, test.accBody)
				if test.notBody != "" {
					assert.NotContains(t, test.notBody, output)
				}
			}
		})
	}
}
