package suckhttp

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Method     HttpMethod
	URL        url.URL
	Headers    []string
	Body       []byte
	Timeout    time.Duration
	RemoteAddr net.Addr
}

func (request *Request) Send(ctx context.Context, conn net.Conn) (response *Response, err error) {
	message, err := CreateRequestMessage(request.Method, request.URL.RequestURI(), request.Headers, request.Body)
	if err != nil {
		return
	}
	// fmt.Println(string(message))
	conn.SetDeadline(time.Now().Add(request.Timeout))
	_, err = conn.Write(message)
	if err != nil {
		return
	}

	response, err = ReadResponse(ctx, conn, request.Timeout)
	return
}

func ReadRequest(ctx context.Context, conn net.Conn, timeout time.Duration) (result *Request, err error) {
	result = &Request{RemoteAddr: conn.RemoteAddr()}
	head, headers, body, err := requestReader(ctx, conn, timeout)
	if err != nil {
		return
	}
	s := strings.Split(head, " ")
	if len(s) < 3 {
		return nil, fmt.Errorf("Error in request HEAD: %s", head)
	}
	result.Method = HttpMethod(s[0])
	result.Headers = headers
	u, err := url.Parse("https://" + result.GetHeader("host") + s[1])
	if err == nil {
		result.URL = *u
	}
	result.Body = body
	return
}

func NewRequest(method HttpMethod, uri string) (*Request, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return &Request{Method: method, URL: *u, Timeout: time.Minute, Headers: []string{"host", u.Hostname()}}, nil
}

func (request *Request) AddHeader(key, value string) *Request {
	if request.Headers == nil {
		request.Headers = make([]string, 0)
	}
	request.Headers = append(request.Headers, []string{key, value}...)
	return request
}

func (request *Request) SetHeader(key, value string) *Request {
	if request.Headers == nil {
		request.Headers = make([]string, 0)
	}
	for i := 0; i < len(request.Headers); i += 2 {
		if request.Headers[i] == key {
			request.Headers[i+1] = value
			return request
		}
	}
	request.Headers = append(request.Headers, []string{key, value}...)
	return request
}

func (request *Request) DeleteHeader(key string) *Request {
	if request.Headers == nil {
		return request
	}
	headers := make([]string, 0)
	for i := 0; i < len(request.Headers); i += 2 {
		if request.Headers[i] != key {
			headers = append(headers, []string{request.Headers[i], request.Headers[i+1]}...)
		}
	}
	request.Headers = headers
	return request
}

func (request *Request) GetHeader(name string) string {
	for i := 0; i < len(request.Headers); i += 2 {
		if request.Headers[i] == name {
			return request.Headers[i+1]
		}
	}
	return ""
}

func (request *Request) String() string {
	message, err := CreateRequestMessage(request.Method, request.URL.RequestURI(), request.Headers, request.Body)
	if err != nil {
		return ""
	}
	return string(message)
}
