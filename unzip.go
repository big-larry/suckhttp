package suckhttp

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/dsnet/compress/brotli"
)

func ungzip(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)
	var buf bytes.Buffer
	var r io.Reader
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	_, err = buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func unbr(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	zr, err := brotli.NewReader(buf, nil)
	if err != nil {
		return nil, err
	}

	result := &bytes.Buffer{}
	if _, err := io.Copy(result, zr); err != nil {
		return nil, err
	}

	if err := zr.Close(); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}
