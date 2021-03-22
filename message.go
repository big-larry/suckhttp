package suckhttp

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/big-larry/suckutils"
)

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	DELETE HttpMethod = "DELETE"
	HEAD   HttpMethod = "HEAD"
)

func createMessage(head string, headers []string, body []byte) (result []byte, err error) {
	if len(head) == 0 {
		return nil, errors.New("Empty HEAD")
	}
	var message bytes.Buffer
	message.WriteString(head)
	message.WriteString("\r\n")
	for i := 0; i < len(headers); i += 2 {
		if headers[i] == Content_Length && len(body) > 0 {
			continue
		}
		message.WriteString(suckutils.ConcatFour(headers[i], ": ", headers[i+1], "\r\n"))
	}
	if len(body) > 0 {
		message.WriteString(suckutils.ConcatFour(Content_Length, ": ", strconv.Itoa(len(body)), "\r\n\r\n"))
		message.Write(body)
	} else {
		message.WriteString("\r\n")
	}
	result = message.Bytes()
	return
}

func CreateRequestMessage(method HttpMethod, location string, headers []string, body []byte) (result []byte, err error) {
	return createMessage(suckutils.ConcatFour(string(method), " ", location, " HTTP/1.1"), headers, body)
}

func CreateResponseMessage(statusCode int, statusText string, headers []string, body []byte) (result []byte, err error) {
	return createMessage(suckutils.ConcatFour("HTTP/1.1", strconv.Itoa(statusCode), " ", statusText), headers, body)
}
