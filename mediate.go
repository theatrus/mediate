// Package mediate provides retryable, failure
// tolerant and rate limited HTTP Transport / RoundTripper interfaces
// for all net.Http client users.
package mediate

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct. Bodies
// should be deep copied here due to closing.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	return r2
}

// cloneResponse makes a new shallow clone of an http.Response
func cloneResponse(r *http.Response) *http.Response {
	// shallow copy of the struct
	r2 := new(http.Response)
	*r2 = *r
	return r2
}

type canceler interface {
	CancelRequest(*http.Request)
}

// FixedRetry transport - on any failure, the request will be retried
// at most count times.
type fixedRetries struct {
	transport      http.RoundTripper
	retriesAllowed int
}

func FixedRetries(count int, transport http.RoundTripper) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &fixedRetries{transport: transport, retriesAllowed: count}
}

func (t *fixedRetries) CancelRequest(req *http.Request) {
	tr, ok := t.transport.(canceler)
	if ok {
		tr.CancelRequest(req)
	}
}

func (t *fixedRetries) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastError error
	for retry := 0; retry < t.retriesAllowed; retry++ {
		nreq := cloneRequest(req)
		resp, lastError := t.transport.RoundTrip(nreq)
		if lastError == nil {
			return resp, nil
		}
	}
	return nil, lastError
}

/////////////////////////

type reliableBody struct {
	transport http.RoundTripper
}

// ReliableBody builds a RoundTripper which will consume all
// of the response Body into a new memory buffer, and returns
// the response with this alternate Body.
//
// This is less memory efficient compared to streaming the response
// from the socket directly, but allows API to work with complete
// operations making retries and other actions trivial.
func ReliableBody(transport http.RoundTripper) http.RoundTripper {
	return &reliableBody{transport}
}

func (t *reliableBody) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewReader(body)
	resp.Body = ioutil.NopCloser(buf)
	return resp, nil
}
