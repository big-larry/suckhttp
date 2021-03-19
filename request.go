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
	method     HttpMethod
	url        url.URL
	headers    []string
	body       []byte
	timeout    time.Duration
	remoteAddr net.Addr
}

func (request *Request) Send(ctx context.Context, conn net.Conn) (response *Response, err error) {
	message, err := CreateRequestMessage(request.method, request.url.RequestURI(), request.headers, request.body)
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
	head, headers, body, err := requestReader(ctx, conn, timeout)
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
		result.url = *u
	}
	result.body = body
	return
}

func NewRequest(method HttpMethod, uri string) (*Request, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return &Request{method: method, url: *u, timeout: time.Minute, headers: []string{"host", u.Hostname()}}, nil
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

func (request *Request) String() string {
	message, err := CreateRequestMessage(request.method, request.url.RequestURI(), request.headers, request.body)
	if err != nil {
		return ""
	}
	return string(message)
}
