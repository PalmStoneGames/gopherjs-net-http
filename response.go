// Copyright 2016 Palm Stone Games, Inc. All rights reserved.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "io"

// Response represents the response from an HTTP request.
//
type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200

					  // Header maps header keys to values.  If the response had multiple
					  // headers with the same key, they may be concatenated, with comma
					  // delimiters.  (Section 4.2 of RFC 2616 requires that multiple headers
					  // be semantically equivalent to a comma-delimited sequence.) Values
					  // duplicated by other fields in this struct (e.g., ContentLength) are
					  // omitted from Header.
					  //
					  // Keys in the map are canonicalized (see CanonicalHeaderKey).
	Header Header

					  // Body represents the response body.
					  //
					  // The http Client and Transport guarantee that Body is always
					  // non-nil, even on responses without a body or responses with
					  // a zero-length body. It is the caller's responsibility to
					  // close Body. The default HTTP client's Transport does not
					  // attempt to reuse HTTP/1.0 or HTTP/1.1 TCP connections
					  // ("keep-alive") unless the Body is read to completion and is
					  // closed.
					  //
					  // The Body is automatically dechunked if the server replied
					  // with a "chunked" Transfer-Encoding.
	Body io.ReadCloser

					  // ContentLength records the length of the associated content.  The
					  // value -1 indicates that the length is unknown.  Unless Request.Method
					  // is "HEAD", values >= 0 indicate that the given number of bytes may
					  // be read from Body.
	ContentLength int64

					  // The Request that was sent to obtain this Response.
					  // Request's Body is nil (having already been consumed).
					  // This is only populated for Client requests.
	Request *Request
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (r *Response) Cookies() []*Cookie {
	return readSetCookies(r.Header)
}