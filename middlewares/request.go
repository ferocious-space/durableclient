package middlewares

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type BodyFN func() (io.Reader, error)

type Request struct {
	body BodyFN
	*http.Request
}

type LenReader interface {
	Len() int
}

func (r *Request) Clone(ctx context.Context) *Request {
	return &Request{
		body:    r.body,
		Request: r.Request.Clone(ctx),
	}
}

func (r *Request) SetBody(body interface{}) error {
	reader, length, err := getReader(body)
	if err != nil {
		return err
	}
	r.body = reader
	r.ContentLength = length
	return nil
}

func getReader(rawBody interface{}) (BodyFN, int64, error) {
	var reader BodyFN
	var length int64
	switch body := rawBody.(type) {
	case BodyFN:
		reader = body
		tmp, err := body()
		if err != nil {
			return nil, 0, err
		}
		if lr, ok := tmp.(LenReader); ok {
			length = int64(lr.Len())
		}
		if c, ok := tmp.(io.Closer); ok {
			c.Close()
		}
	case func() (io.Reader, error):
		reader = body
		tmp, err := body()
		if err != nil {
			return nil, 0, err
		}
		if lr, ok := tmp.(LenReader); ok {
			length = int64(lr.Len())
		}
		if c, ok := tmp.(io.Closer); ok {
			c.Close()
		}
	case []byte:
		buf := body
		reader = func() (io.Reader, error) {
			return bytes.NewReader(buf), nil
		}
		length = int64(len(buf))
	case *bytes.Buffer:
		buf := body
		reader = func() (io.Reader, error) {
			return bytes.NewReader(buf.Bytes()), nil
		}
		length = int64(buf.Len())

	case *bytes.Reader:
		bodyBuffer := new(bytes.Buffer)
		n, err := io.Copy(bodyBuffer, body)
		if err != nil {
			return nil, 0, err
		}
		reader = func() (io.Reader, error) {
			return bodyBuffer, nil
		}
		length = n
	case io.Reader:
		bodyBuffer := new(bytes.Buffer)
		n, err := io.Copy(bodyBuffer, body)
		if err != nil {
			return nil, 0, err
		}
		reader = func() (io.Reader, error) {
			return bodyBuffer, nil
		}
		length = n
	case nil:

	default:
		return nil, 0, errors.Errorf("cannot handle %T", rawBody)
	}
	switch body := rawBody.(type) {
	case io.Closer:
		body.Close()
	default:

	}
	return reader, length, nil
}

func FromRequest(req *http.Request) (*Request, error) {
	r := &Request{Request: req}
	return r, r.SetBody(req.Body)
}
