// Copyright 2016 Palm Stone Games, Inc. All rights reserved.
// Copyright (c) 2013 Richard Musiol. All rights reserved.

package http

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

var DefaultTransport RoundTripper = &XHRTransport{}

var errRequestCanceled = errors.New("code.palmstonegames.com/gopherjs-net-http: request canceled")

type XHRTransport struct {
	inflight map[*Request]*js.Object
}

func (t *XHRTransport) RoundTrip(req *Request) (*Response, error) {
	xhrConstructor := js.Global.Get("XMLHttpRequest")
	if xhrConstructor == js.Undefined {
		return nil, errors.New("code.palmstonegames.com/gopherjs-net-http: XMLHttpRequest not available")
	}
	xhr := xhrConstructor.New()

	if t.inflight == nil {
		t.inflight = map[*Request]*js.Object{}
	}
	t.inflight[req] = xhr
	defer delete(t.inflight, req)

	respCh := make(chan *Response)
	errCh := make(chan error)

	xhr.Set("onload", func() {
		header, _ := textproto.NewReader(bufio.NewReader(bytes.NewReader([]byte(xhr.Call("getAllResponseHeaders").String() + "\n")))).ReadMIMEHeader()
		body := js.Global.Get("Uint8Array").New(xhr.Get("response")).Interface().([]byte)

		contentLength := int64(-1)
		switch req.Method {
		case "HEAD":
			if l, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64); err == nil {
				contentLength = l
			}
		default:
			contentLength = int64(len(body))
		}

		respCh <- &Response{
			Status:        xhr.Get("status").String() + " " + xhr.Get("statusText").String(),
			StatusCode:    xhr.Get("status").Int(),
			Header:        Header(header),
			ContentLength: contentLength,
			Body:          ioutil.NopCloser(bytes.NewReader(body)),
			Request:       req,
		}
	})

	xhr.Set("onerror", func(e *js.Object) {
		errCh <- fmt.Errorf("code.palmstonegames.com/gopherjs-net-http: XMLHttpRequest failed with error: %v", e.String())
	})

	xhr.Set("onabort", func(e *js.Object) {
		errCh <- errRequestCanceled
	})

	xhr.Call("open", req.Method, req.URL.String())
	xhr.Set("responseType", "arraybuffer") // has to be after "open" until https://bugzilla.mozilla.org/show_bug.cgi?id=1110761 is resolved
	for key, values := range req.Header {
		for _, value := range values {
			xhr.Call("setRequestHeader", key, value)
		}
	}
	var body []byte
	if req.Body != nil {
		var err error
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	xhr.Call("send", body)

	select {
	case resp := <-respCh:
		return resp, nil
	case err := <-errCh:
		return nil, err
	}
}

func (t *XHRTransport) CancelRequest(req *Request) {
	if xhr, ok := t.inflight[req]; ok {
		xhr.Call("abort")
	}
}
