package suckhttp

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	request    *Request
	statusCode int
	statusText string
	headers    []string
	body       []byte
	Time       time.Duration
}

func NewResponse(statusCode int, statusText string) *Response {
	return &Response{statusCode: statusCode, statusText: statusText}
}

func (response *Response) SetStatusCode(statusCode int, statusText string) {
	response.statusCode = statusCode
	response.statusText = statusText
}

func (response *Response) Write(conn net.Conn, timeout time.Duration) error {
	message, err := CreateResponseMessage(response.statusCode, response.statusText, response.headers, response.body)
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(timeout))
	_, err = conn.Write(message)
	return err
}

func ReadResponse(ctx context.Context, conn net.Conn, timeout time.Duration) (response *Response, err error) {
	head, headers, body, time, err := requestReader(ctx, conn, timeout)
	if err != nil {
		return
	}
	response = &Response{Time: time}
	pos1 := strings.Index(head, " ")
	space_pos := strings.Index(head[pos1+1:], " ")
	if space_pos == -1 {
		response.statusCode, err = strconv.Atoi(strings.TrimSpace(head[pos1+1:]))
	} else {
		response.statusCode, err = strconv.Atoi(head[pos1+1 : pos1+1+space_pos])
		response.statusText = strings.TrimSpace(head[pos1+space_pos+2:])
	}
	response.headers = headers
	// enc := response.GetHeader(Content_Encoding)
	// if enc == "gzip" {
	// 	body, err = ungzip(body)
	// } else if enc == "br" {
	// 	body, err = unbr(body)
	// } else if enc == "zstd" {
	// }
	response.DeleteHeader(Transfer_Encoding)
	response.body = body
	return
}

func (response *Response) SetBody(body []byte) *Response {
	response.body = body
	return response
}

func (response *Response) AddHeader(key, value string) *Response {
	if response.headers == nil {
		response.headers = make([]string, 0)
	}
	response.headers = append(response.headers, []string{key, value}...)
	return response
}

func (response *Response) SetHeader(key, value string) *Response {
	if response.headers == nil {
		response.headers = make([]string, 0)
	}
	for i := 0; i < len(response.headers); i += 2 {
		if response.headers[i] == key {
			response.headers[i+1] = value
			return response
		}
	}
	response.headers = append(response.headers, []string{key, value}...)
	return response
}

func (response *Response) DeleteHeader(key string) *Response {
	if response.headers == nil {
		return response
	}
	headers := make([]string, 0)
	for i := 0; i < len(response.headers); i += 2 {
		if response.headers[i] != key {
			headers = append(headers, []string{response.headers[i], response.headers[i+1]}...)
		}
	}
	response.headers = headers
	return response
}

func (response *Response) GetHeaderFirstValue(name string) string {
	for i := 0; i < len(response.headers); i += 2 {
		if response.headers[i] == name {
			return response.headers[i+1]
		}
	}
	return ""
}

func (response *Response) GetHeaderValues(name string) []string {
	result := make([]string, 0)
	for i := 0; i < len(response.headers); i += 2 {
		if response.headers[i] == name {
			result = append(result, response.headers[i+1])
		}
	}
	return result
}

func (response *Response) GetBody() []byte {
	return response.body
}
func (response *Response) Bytes() []byte {
	message, err := CreateResponseMessage(response.statusCode, response.statusText, response.headers, response.body)
	if err != nil {
		return nil
	}
	return message
}
func (response *Response) String() string {
	message, err := CreateResponseMessage(response.statusCode, response.statusText, response.headers, response.body)
	if err != nil {
		return ""
	}
	return string(message)
}

func (response *Response) GetStatus() (int, string) {
	return response.statusCode, response.statusText
}
