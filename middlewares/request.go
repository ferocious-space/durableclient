package middlewares

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
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

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Request = r.Request.WithContext(ctx)
	return r
}

func (r *Request) Clone(ctx context.Context) *Request {
	r.Request = r.Request.Clone(ctx)
	return r
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

func (r *Request) Bytes() ([]byte, error) {
	if r.body == nil {
		return nil, nil
	}
	b, err := r.body()
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
		buf, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, 0, err
		}
		reader = func() (io.Reader, error) {
			return bytes.NewBuffer(buf), nil
		}
		length = int64(len(buf))
	case io.Reader:
		buf, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, 0, err
		}
		reader = func() (io.Reader, error) {
			return bytes.NewBuffer(buf), nil
		}
		length = int64(len(buf))
	case nil:
	default:
		return nil, 0, errors.Errorf("cannot handle %T", rawBody)
	}
	return reader, length, nil
}

func FromRequest(req *http.Request) (*Request, error) {
	reader, _, err := getReader(req.Body)
	if err != nil {
		return nil, err
	}
	return &Request{
		body:    reader,
		Request: req,
	}, nil
}

func DrainReader(body io.ReadCloser) {
	defer body.Close()
	_, _ = ioutil.ReadAll(body)
}
