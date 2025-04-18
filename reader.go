package suckhttp

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

type httpReader struct {
	conn net.Conn
	data []byte
	time time.Duration
}

func requestReader(ctx context.Context, conn net.Conn, timeout time.Duration) (string, []string, []byte, time.Duration, error) {
	return read(ctx, conn, timeout)
}

func read(ctx context.Context, conn net.Conn, timeout time.Duration) (string, []string, []byte, time.Duration, error) {
	reader := &httpReader{conn: conn}
	head := ""
	headers := make([]string, 0, 50)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		conn.SetReadDeadline(time.Now().Add(timeout))
	} else {
		conn.SetReadDeadline(time.Time{})
	}
	var body *bytes.Buffer
	contentLength := 0
	var trailerHeaders []string
	firstLine := true
	readBody := false
	var err error

loop:
	for {
		select {
		case <-ctx.Done():
			return "", nil, nil, reader.time, errors.New("Canceled")
		default:
			line := reader.readLine()
			if line == nil {
				err = reader.loadNext(1024)
				if err != nil {
					return "", nil, nil, reader.time, err
				}
				continue
			}
			if !readBody && len(line) == 0 {
				readBody = true
			} else if readBody && len(line) == 0 {
				break loop
			}

			if !readBody {
				s := string(line)
				if firstLine {
					firstLine = false
					head = s
					continue
				}
				index := strings.Index(s, ":")
				if index != -1 {
					name := strings.ToLower(strings.TrimSpace(s[:index]))
					value := strings.TrimSpace(s[index+1:])
					if name == Content_Length {
						contentLength, _ = strconv.Atoi(value)
					}
					if name == Transfer_Encoding && strings.ToLower(value) == "chunked" {
						contentLength = -1
						body = new(bytes.Buffer)
					} else if name == Trailer {
						s := strings.Split(strings.ToLower(value), ",")
						trailerHeaders = make([]string, len(s))
						for i := 0; i < len(s); i++ {
							trailerHeaders[i] = strings.TrimSpace(s[i])
						}
					}
					headers = append(headers, name)
					headers = append(headers, value)
					continue
				}
			} else {
				if contentLength > 0 {
					body = new(bytes.Buffer)
					data, err := reader.read(contentLength)
					// if err == io.EOF && contentLength == len(data) {
					// 	log.Println("EOF:", "Readed done", contentLength)
					// }
					if err != nil {
						return "", nil, nil, reader.time, err
					}
					body.Write(data)
				} else if contentLength == -1 {
					//log.Println("chunked")
					if len(line) == 0 {
						continue
					}
					count := make([]byte, 8)
					if len(line)%2 > 0 {
						line = append(line, 0)
						for j := len(line) - 2; j >= 0; j-- {
							line[j+1] = line[j]
						}
						line[0] = '0'
					}
					l, _ := hex.Decode(count, line)
					a := 0
					for j := l - 1; j >= 0; j-- {
						a |= int(count[j]) << (8 * (l - 1 - j))
					}
					// log.Println(a)

					if a == 0 && len(trailerHeaders) == 0 {
						reader.readLine()
						// headersHandler("content-length", strconv.Itoa(body.Len()))
						break loop
					} else if a == 0 && len(trailerHeaders) > 0 {
						// fmt.Println(trailerHeaders)
						// fmt.Println(headers)
						reader.readLine() //TODO: Hack for ozon
						// headersHandler("content-length", strconv.Itoa(body.Len()))
						break loop
						return "", nil, nil, reader.time, errors.New("Not implemented trailer headers")
					}

					data, err := reader.read(a + 2)
					if err != nil {
						return "", nil, nil, reader.time, err
					}
					body.Write(data[:a])
					continue
				}
				break loop
			}
		}
	}

	if body != nil {
		return head, headers, body.Bytes(), reader.time, nil
	}
	return head, headers, nil, reader.time, nil
}

func (reader *httpReader) readBuf(buf []byte) (int, error) {
	now := time.Now()
	n, err := reader.conn.Read(buf)
	reader.time += time.Now().Sub(now)
	return n, err
}

func (reader *httpReader) loadNext(count int) error {
	buf := make([]byte, count)
	if reader.data == nil {
		reader.data = make([]byte, 0)
	}
	n, err := reader.readBuf(buf)
	if err != nil {
		return err
	}
	// reader.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	reader.data = append(reader.data, buf[:n]...)
	return nil
}

func (reader *httpReader) loadWhile(count int) error {
	buf := make([]byte, count)
	shift := 0
	if reader.data == nil {
		reader.data = make([]byte, 0)
	}
	ost := count
	for {
		n, err := reader.readBuf(buf[shift:])
		if err != nil && n != ost {
			return err
		}
		//reader.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		reader.data = append(reader.data, buf[shift:shift+n]...)
		if n == ost {
			break
		}
		shift += n
		ost -= n
	}
	return nil
}

var clf []byte = []byte("\r\n")

func (reader *httpReader) readLine() []byte {
	if len(reader.data) == 0 {
		return nil
	}
	index := bytes.Index(reader.data, clf)
	if index == -1 {
		return nil
	}
	result := reader.data[:index]
	reader.data = reader.data[index+2:]
	// log.Println("Readed line", string(result))
	return result
}

func (reader *httpReader) read(count int) ([]byte, error) {
	if len(reader.data) < count {
		err := reader.loadWhile(count - len(reader.data))
		if err != nil {
			return nil, err
		}
	}

	result := reader.data[:count]
	reader.data = reader.data[count:]
	return result, nil
}
