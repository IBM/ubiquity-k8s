package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/trace"
)

type testLogger struct {
	buf bytes.Buffer
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(&l.buf, format, v...)
}

func (l *testLogger) Print(v ...interface{}) {
	fmt.Fprint(&l.buf, v...)
}

func (l *testLogger) Println(v ...interface{}) {
	fmt.Fprintln(&l.buf, v...)
}

func (l *testLogger) Clear() {
	l.buf.Reset()
}

func (l *testLogger) Dump() []byte {
	return l.buf.Bytes()
}

type TransportTestSuite struct {
	suite.Suite

	client *http.Client
	logger *testLogger

	oldTraceLogger trace.Printer
}

func TestTransportTestSuite(t *testing.T) {
	suite.Run(t, new(TransportTestSuite))
}

func (suite *TransportTestSuite) SetupSuite() {
	suite.oldTraceLogger = trace.Logger
}

func (suite *TransportTestSuite) TearDownSuite() {
	trace.Logger = suite.oldTraceLogger
}

func (suite *TransportTestSuite) SetupTest() {
	suite.logger = new(testLogger)
	trace.Logger = suite.logger
	suite.client = &http.Client{Transport: NewTraceLoggingTransport(nil)}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Client")

}

func helloRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/hello", http.StatusMovedPermanently)
	} else {
		fmt.Fprintln(w, "Hello, Client")
	}
}

func (suite *TransportTestSuite) TestTraceSimple() {
	ts := httptest.NewServer(http.HandlerFunc(helloHandler))
	defer ts.Close()

	expect := []string{
		"REQUEST: ",
		"GET / HTTP/1.1",
		"Host: " + strings.TrimLeft(ts.URL, "http://"),

		"RESPONSE: ",
		"Elapsed: ",
		"HTTP/1.1 200 OK",
		"Content-Length: 14",
		"Content-Type: text/plain; charset=utf-8",

		"Hello, Client",
	}

	resp, err := suite.client.Get(ts.URL)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	for _, e := range expect {
		suite.Contains(string(suite.logger.Dump()), e)
	}
}

func (suite *TransportTestSuite) TestTraceRedirect() {
	ts := httptest.NewServer(http.HandlerFunc(helloRedirectHandler))
	defer ts.Close()

	host := strings.TrimLeft(ts.URL, "http://")
	expect := []string{
		"REQUEST: ",
		"GET / HTTP/1.1",
		"Host: " + host,

		"RESPONSE: ",
		"Elapsed: ",
		"HTTP/1.1 301 Moved Permanently",
		"Content-Length: 41",
		"Content-Type: text/html; charset=utf-8",
		"Location: /hello",
		"<a href=\"/hello\">Moved Permanently</a>.",

		// redirected request
		"GET /hello HTTP/0.0",
		"Host: " + host,
		"Referer: " + ts.URL,

		// redirected response
		"HTTP/1.1 200 OK",
		"Content-Length: 14",
		"Content-Type: text/plain; charset=utf-8",

		"Hello, Client",
	}

	resp, err := suite.client.Get(ts.URL)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	for _, e := range expect {
		suite.Contains(string(suite.logger.Dump()), e)
	}
}
