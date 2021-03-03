package suckhttp

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Request    *Request
	StatusCode int
	StatusText string
	Headers    []string
	Body       []byte
}

func NewResponse(statusCode int, statusText string) *Response {
	return &Response{StatusCode: statusCode, StatusText: statusText}
}

func (response *Response) SetStatusCode(statusCode int, statusText string) {
	response.StatusCode = statusCode
	response.StatusText = statusText
}

func (response *Response) Write(conn net.Conn, timeout time.Duration) error {
	message, err := CreateResponseMessage(response.StatusCode, response.StatusText, response.Headers, response.Body)
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(timeout))
	_, err = conn.Write(message)
	return err
}

func ReadResponse(ctx context.Context, conn net.Conn, timeout time.Duration) (response *Response, err error) {
	head, headers, body, err := requestReader(ctx, conn, timeout)
	if err != nil {
		return
	}
	response = &Response{}
	s := strings.Split(head, " ")
	response.StatusCode, err = strconv.Atoi(s[1])
	if len(s) > 2 {
		response.StatusText = s[2]
	}
	response.Headers = headers
	enc := response.GetHeader(Content_Encoding)
	if enc == "gzip" {
		body, err = ungzip(body)
	} else if enc == "br" {
		body, err = unbr(body)
	} else if enc == "zstd" {
		// TODO
	}
	response.DeleteHeader(Transfer_Encoding)
	response.Body = body
	return
}

func (response *Response) SetBody(body []byte) *Response {
	response.Body = body
	return response
}

func (response *Response) AddHeader(key, value string) *Response {
	if response.Headers == nil {
		response.Headers = make([]string, 0)
	}
	response.Headers = append(response.Headers, []string{key, value}...)
	return response
}

func (response *Response) SetHeader(key, value string) *Response {
	if response.Headers == nil {
		response.Headers = make([]string, 0)
	}
	for i := 0; i < len(response.Headers); i += 2 {
		if response.Headers[i] == key {
			response.Headers[i+1] = value
			return response
		}
	}
	response.Headers = append(response.Headers, []string{key, value}...)
	return response
}

func (response *Response) DeleteHeader(key string) *Response {
	if response.Headers == nil {
		return response
	}
	headers := make([]string, 0)
	for i := 0; i < len(response.Headers); i += 2 {
		if response.Headers[i] != key {
			headers = append(headers, []string{response.Headers[i], response.Headers[i+1]}...)
		}
	}
	response.Headers = headers
	return response
}

func (response *Response) GetHeader(name string) string {
	for i := 0; i < len(response.Headers); i += 2 {
		if response.Headers[i] == name {
			return response.Headers[i+1]
		}
	}
	return ""
}

func (response *Response) GetHeaderValues(name string) []string {
	result := make([]string, 0)
	for i := 0; i < len(response.Headers); i += 2 {
		if response.Headers[i] == name {
			result = append(result, response.Headers[i+1])
		}
	}
	return result
}
