package suckhttp

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	go func() {
		if err := listenWithoutClose(); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(time.Second)

	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := NewRequest(GET, "/hi")
	r.AddHeader(Referer, "127.0.0.1")
	fmt.Println(string(r.String()))
	resp, err := r.Send(context.Background(), conn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(resp.String()))

	time.Sleep(time.Second)
	r, _ = NewRequest(GET, "/hi")
	r.AddHeader(Referer, "127.0.0.1")
	fmt.Println(string(r.String()))
	resp, err = r.Send(context.Background(), conn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(resp.String()))
}

func listenWithClose() error {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		request, err := ReadRequest(context.Background(), conn, time.Second)
		if err != nil {
			return err
		}
		response := NewResponse(200, "OK")
		response.SetBody([]byte(request.url.RequestURI()))
		err = response.Write(conn, time.Second)
		if err != nil {
			return err
		}

		err = conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func listenWithoutClose() error {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			for {
				request, err := ReadRequest(context.Background(), c, time.Second*10)
				if err != nil {
					fmt.Println("read", err)
					break
				}
				response := NewResponse(200, "OK")
				response.SetBody([]byte(request.url.RequestURI()))
				err = response.Write(c, time.Second)
				if err != nil {
					fmt.Println("write", err)
					break
				}
			}

			conn.Close()
		}(conn)
	}
	// if !close {
	// 	err = conn.Close()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
