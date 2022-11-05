package httpreq

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	url1 "net/url"
)

var (
	defaultClient = Client{Client: http.DefaultClient}
)

type Client struct {
	*http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{Client: client}
}

func (c *Client) NewRequest(ctx context.Context, method, url string, body interface{}) *Request {
	return newRequest(c.Client, ctx, method, url, body)
}

type Request struct {
	err        error
	client     *http.Client
	req        *http.Request
	query      url1.Values
	body       interface{}
	statusCode int
	respBody   []byte
}

func NewRequest(ctx context.Context, method, url string, body interface{}) *Request {
	return defaultClient.NewRequest(ctx, method, url, body)
}

func newRequest(client *http.Client, ctx context.Context, method, url string, body interface{}) *Request {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	r := &Request{
		err:    err,
		client: client,
		req:    req,
		query:  map[string][]string{},
		body:   body,
	}
	r.ContentType(matchContentType(body))
	return r
}

func matchContentType(body interface{}) string {
	switch body.(type) {
	case []byte, string:
		return ContentTypeText
	case url1.Values, map[string][]string:
		return ContentTypeFormURLEncoded
	default:
		return ContentTypeJSON
	}
}

func (r *Request) Header(key string, value string, values ...string) *Request {
	if r.err == nil {
		r.req.Header.Set(key, value)
		for i := range values {
			r.req.Header.Add(key, values[i])
		}
	}
	return r
}

func (r *Request) ContentType(value string) *Request {
	if r.err == nil {
		r.req.Header.Set(HeaderContentType, value)
	}
	return r
}

func (r *Request) Query(key, value string, values ...string) *Request {
	if r.err == nil {
		r.query.Set(key, value)
		for i := range values {
			r.query.Add(key, values[i])
		}
	}
	return r
}

func (r *Request) Call() *Request {
	if r.err != nil {
		return r
	}

	if r.body != nil {
		var bodyBytes []byte
		marshaller, ok := marshallers[r.req.Header[HeaderContentType][0]]
		if !ok {
			r.err = fmt.Errorf("invalid Content-Type: %v", r.req.Header[HeaderContentType][0])
			return r
		}
		bodyBytes, r.err = marshaller.Marshal(r.body)
		if r.err != nil {
			return r
		}
		r.req.Body = bytesReaderCloser{Reader: bytes.NewReader(bodyBytes)}
	}

	r.req.URL.RawQuery = r.query.Encode()
	resp, err := r.client.Do(r.req)
	if err != nil {
		r.err = err
		return r
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.err = err
		return r
	}
	r.statusCode = resp.StatusCode
	r.respBody = body
	return r
}

func (r *Request) Expect200() *Request {
	return r.ExpectStatusCodes(http.StatusOK)
}

func (r *Request) ExpectStatusCodes(codes ...int) *Request {
	if r.err != nil {
		return r
	}

	for i := range codes {
		if codes[i] == r.statusCode {
			return r
		}
	}
	r.err = UnexpectedStatusCodeError{StatusCode: r.statusCode}
	return r
}

func (r *Request) Response() ([]byte, error) {
	return r.respBody, r.err
}

type bytesReaderCloser struct {
	*bytes.Reader
}

func (b bytesReaderCloser) Close() error {
	return nil
}
