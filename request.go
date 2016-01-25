// Copyright 2016 Palm Stone Games, Inc. All rights reserved.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/url"
	"io"
	"fmt"
	"io/ioutil"
)

// A Request represents an HTTP request received by a server
// or to be sent by a client.
//
// The field semantics differ slightly between client and server
// usage. In addition to the notes on the fields below, see the
// documentation for Request.Write and RoundTripper.
type Request struct {
					  // Method specifies the HTTP method (GET, POST, PUT, etc.).
					  // For client requests an empty string means GET.
	Method string

					  // URL specifies either the URI being requested (for server
					  // requests) or the URL to access (for client requests).
					  //
					  // For server requests the URL is parsed from the URI
					  // supplied on the Request-Line as stored in RequestURI.  For
					  // most requests, fields other than Path and RawQuery will be
					  // empty. (See RFC 2616, Section 5.1.2)
					  //
					  // For client requests, the URL's Host specifies the server to
					  // connect to, while the Request's Host field optionally
					  // specifies the Host header value to send in the HTTP
					  // request.
	URL *url.URL

					  // A header maps request lines to their values.
					  // If the header says
					  //
					  //	accept-encoding: gzip, deflate
					  //	Accept-Language: en-us
					  //	Connection: keep-alive
					  //
					  // then
					  //
					  //	Header = map[string][]string{
					  //		"Accept-Encoding": {"gzip, deflate"},
					  //		"Accept-Language": {"en-us"},
					  //		"Connection": {"keep-alive"},
					  //	}
					  //
					  // HTTP defines that header names are case-insensitive.
					  // The request parser implements this by canonicalizing the
					  // name, making the first character and any characters
					  // following a hyphen uppercase and the rest lowercase.
					  //
					  // For client requests certain headers are automatically
					  // added and may override values in Header.
					  //
					  // See the documentation for the Request.Write method.
	Header Header

					  // Body is the request's body.
					  //
					  // For client requests a nil body means the request has no
					  // body, such as a GET request. The HTTP Client's Transport
					  // is responsible for calling the Close method.
					  //
					  // For server requests the Request Body is always non-nil
					  // but will return EOF immediately when no body is present.
					  // The Server will close the request body. The ServeHTTP
					  // Handler does not need to.
	Body io.ReadCloser
}

// NewRequest returns a new Request given a method, URL, and optional body.
//
// If the provided body is also an io.Closer, the returned
// Request.Body is set to body and will be closed by the Client
// methods Do, Post, and PostForm, and Transport.RoundTrip.
//
// NewRequest returns a Request suitable for use with Client.Do or
// Transport.RoundTrip.
// To create a request for use with testing a Server Handler use either
// ReadRequest or manually update the Request fields. See the Request
// type's documentation for the difference between inbound and outbound
// request fields.
func NewRequest(method, urlStr string, body io.Reader) (*Request, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	req := &Request{
		Method:     method,
		URL:        u,
		Header:     make(Header),
		Body:       rc,
	}

	return req, nil
}

// AddCookie adds a cookie to the request.  Per RFC 6265 section 5.4,
// AddCookie does not attach more than one Cookie header field.  That
// means all cookies, if any, are written into the same line,
// separated by semicolon.
func (r *Request) AddCookie(c *Cookie) {
	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
	if c := r.Header.Get("Cookie"); c != "" {
		r.Header.Set("Cookie", c+"; "+s)
	} else {
		r.Header.Set("Cookie", s)
	}
}

func (r *Request) closeBody() {
	if r.Body != nil {
		r.Body.Close()
	}
}