package suckhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	method     HttpMethod
	Uri        url.URL
	headers    []string
	Body       []byte
	timeout    time.Duration
	remoteAddr net.Addr
}

func (request *Request) Send(ctx context.Context, conn net.Conn) (response *Response, err error) {
	message, err := CreateRequestMessage(request.method, request.Uri.RequestURI(), request.headers, request.Body)
	if err != nil {
		return
	}
	// fmt.Println(string(message))
	conn.SetDeadline(time.Now().Add(request.timeout))
	_, err = conn.Write(message)
	if err != nil {
		return
	}

	response, err = ReadResponse(ctx, conn, request.timeout)
	conn.SetDeadline(time.Time{})
	return
}

func ReadRequest(ctx context.Context, conn net.Conn, timeout time.Duration) (result *Request, err error) {
	result = &Request{remoteAddr: conn.RemoteAddr()}
	head, headers, body, time, err := requestReader(ctx, conn, timeout)
	if err != nil {
		return
	}
	s := strings.Split(head, " ")
	if len(s) < 3 {
		return nil, fmt.Errorf("Error in request HEAD: %s", head)
	}
	result.method = HttpMethod(s[0])
	result.headers = headers
	u, err := url.Parse("https://" + result.GetHeader("host") + s[1])
	if err == nil {
		result.Uri = *u
	}
	result.Body = body
	result.Time = time
	return
}

func NewRequest(method HttpMethod, uri string) (*Request, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	headers := make([]string, 0, 2)
	if u.Hostname() != "" {
		headers = append(headers, []string{"host", u.Hostname()}...)
	}
	return &Request{method: method, Uri: *u, timeout: time.Minute, headers: headers}, nil
}

func (request *Request) AddHeader(key, value string) *Request {
	if request.headers == nil {
		request.headers = make([]string, 0)
	}
	request.headers = append(request.headers, []string{key, value}...)
	return request
}

func (request *Request) SetHeader(key, value string) *Request {
	if request.headers == nil {
		request.headers = make([]string, 0)
	}
	for i := 0; i < len(request.headers); i += 2 {
		if request.headers[i] == key {
			request.headers[i+1] = value
			return request
		}
	}
	request.headers = append(request.headers, []string{key, value}...)
	return request
}

func (request *Request) DeleteHeader(key string) *Request {
	if request.headers == nil {
		return request
	}
	headers := make([]string, 0)
	for i := 0; i < len(request.headers); i += 2 {
		if request.headers[i] != key {
			headers = append(headers, []string{request.headers[i], request.headers[i+1]}...)
		}
	}
	request.headers = headers
	return request
}

func (request *Request) GetHeader(name string) string {
	for i := 0; i < len(request.headers); i += 2 {
		if request.headers[i] == name {
			return request.headers[i+1]
		}
	}
	return ""
}

func (request *Request) GetRemoteAddr() string {
	if a := request.GetHeader("x-real-ip"); a != "" {
		return a
	}
	return request.remoteAddr.String()
}

func (request *Request) String() string {
	message, err := CreateRequestMessage(request.method, request.Uri.RequestURI(), request.headers, request.Body)
	if err != nil {
		return ""
	}
	return string(message)
}

func (request *Request) Clone(newuri string, timeout time.Duration) (*Request, error) {
	if request == nil {
		return nil, errors.New("Request is nil")
	}
	uri, err := url.Parse(newuri)
	if err != nil {
		return nil, err
	}
	result := &Request{
		method:     request.method,
		headers:    request.headers,
		timeout:    timeout,
		remoteAddr: request.remoteAddr,
		Uri:        *uri,
	}
	return result, nil
}
