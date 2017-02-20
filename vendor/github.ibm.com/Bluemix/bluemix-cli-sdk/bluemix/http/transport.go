package http

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/terminal"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/trace"
)

// TraceLoggingTransport is a thin wrapper around Transport. It dumps HTTP
// request and response using trace logger, based on the "BLUEMIX_TRACE"
// environment variable. Sensitive user data will be replaced by text
// "[PRIVATE DATA HIDDEN]".
type TraceLoggingTransport struct {
	rt http.RoundTripper
}

// NewTraceLoggingTransport returns a TraceLoggingTransport wrapping around
// the passed RoundTripper. If the passed RoundTripper is nil, HTTP
// DefaultTransport is used.
func NewTraceLoggingTransport(rt http.RoundTripper) *TraceLoggingTransport {
	if rt == nil {
		return &TraceLoggingTransport{
			rt: http.DefaultTransport,
		}
	}
	return &TraceLoggingTransport{
		rt: rt,
	}
}

func (r *TraceLoggingTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	r.dumpRequest(req, start)
	resp, err = r.rt.RoundTrip(req)
	if err != nil {
		return
	}
	r.dumpResponse(resp, start)
	return
}

func (r *TraceLoggingTransport) dumpRequest(req *http.Request, start time.Time) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")

	dumpedRequest, err := httputil.DumpRequest(req, shouldDisplayBody)
	if err != nil {
		trace.Logger.Printf("An error occurred while dumping request:\n%v\n", err)
		return
	}

	trace.Logger.Printf("\n%s [%s]\n%s\n",
		terminal.HeaderColor("REQUEST:"),
		start.Format(time.RFC3339),
		trace.Sanitize(string(dumpedRequest)))

	if !shouldDisplayBody {
		trace.Logger.Println("[MULTIPART/FORM-DATA CONTENT HIDDEN]")
	}
}

func (r *TraceLoggingTransport) dumpResponse(res *http.Response, start time.Time) {
	end := time.Now()

	dumpedResponse, err := httputil.DumpResponse(res, true)
	if err != nil {
		trace.Logger.Printf("An error occurred while dumping response:\n%v\n", err)
		return
	}

	trace.Logger.Printf("\n%s [%s] %s %.0fms\n%s\n",
		terminal.HeaderColor("RESPONSE:"),
		end.Format(time.RFC3339),
		terminal.HeaderColor("Elapsed:"),
		end.Sub(start).Seconds()*1000,
		trace.Sanitize(string(dumpedResponse)))
}
