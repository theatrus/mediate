package mediate

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

type mockRoundTripper struct {
	RespondWith      *http.Response
	RespondWithError error
	Calls            int
}

func (t *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Calls++
	if t.RespondWithError == nil {
		return t.RespondWith, nil
	} else {
		return nil, t.RespondWithError
	}
}

// An io.ReaderCloser which can return errors or zero length
// responses based on the ReturnError
type mockReaderCloser struct {
	ReturnError bool
}

func (m *mockReaderCloser) Read(p []byte) (n int, err error) {
	if m.ReturnError {
		return 0, io.ErrClosedPipe
	} else {
		return 0, io.EOF
	}
}

func (m *mockReaderCloser) Close() error {
	return nil
}

func newMock() (rc *mockReaderCloser, rt *mockRoundTripper) {
	rc = &mockReaderCloser{}
	rt = &mockRoundTripper{RespondWith: &http.Response{Body: rc}}
	return
}

func TestMock(t *testing.T) {
	_, rt := newMock()
	req := &http.Request{}

	check, error := rt.RoundTrip(req)
	assert.Nil(t, error)
	assert.Equal(t, check, rt.RespondWith, "Wrong response received")
}

func TestReliableBody(t *testing.T) {
	rc, rt := newMock()
	reliable := ReliableBody(rt)

	req := &http.Request{}
	// Check that calling the roundtrip now returns an error
	rc.ReturnError = true
	_, err := reliable.RoundTrip(req)
	assert.NotNil(t, err)

	req = &http.Request{}
	// Check that reliable works with no errors
	rc.ReturnError = false
	_, err = reliable.RoundTrip(req)
	assert.Nil(t, err)
}

func TestFixedRetries(t *testing.T) {
	_, rt := newMock()
	// Three fixed retries
	reliable := FixedRetries(3, rt)

	// Check that basic requests work
	req := &http.Request{}
	_, err := reliable.RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, 1, rt.Calls, "One call")

	// Now we fail the response
	req = &http.Request{}
	rt.RespondWithError = errors.New("generic error")
	fail, err := reliable.RoundTrip(req)
	assert.Nil(t, fail)
	assert.NotNil(t, err)
	assert.Equal(t, rt.RespondWithError, err, "Errors didn't match")
	assert.Equal(t, 4, rt.Calls, "Three retries")
}
