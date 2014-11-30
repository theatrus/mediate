package mediate

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type mockRoundTripper struct {
	RespondWith      *http.Response
	RespondWithError error
}

func (t *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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
		return 0, errors.New("Generic error")
	} else {
		return 0, nil
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
