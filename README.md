mediate
============

[![GoDoc](https://godoc.org/github.com/theatrus/mediate?status.svg)](https://godoc.org/github.com/theatrus/mediate) [![Build Status](https://travis-ci.org/theatrus/mediate.svg)](https://travis-ci.org/theatrus/mediate)

HTTP client transport for real-world and failure-tolerant
communication in Go.

The following Golang adapters are designed to mix-in to the chain of
`http.RoundTripper`s to add functionality such as:

 * `FixedRetries`: on recoverable errors, clone the request and attempt to
 re-issue it again a fixed number of times.
 * `ReliableBody`: consume the entire `http.Response.Body` into
 memory. If reading the body fails, the entire `http.RoundTripper`
 fails.

## Example

	httpClient := &http.Client{}
	httpClient.Transport = mediate.FixedRetries(3,
		mediate.ReliableBody(http.DefaultTransport),
	)
