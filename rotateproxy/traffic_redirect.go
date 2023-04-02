package rotateproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	largeBufferSize = 32 * 1024 // 32KB large buffer
	ErrNotSocks5Proxy = errors.New("this is not a socks proxy server")
)

type RedirectClient struct {
	ListenAddr string
}

func NewRedirectClient(addr string) *RedirectClient {
	c := &RedirectClient{}
	c.ListenAddr = addr
	return c
}

func (c *RedirectClient) Serve() error {
	l, err := net.Listen("tcp", c.ListenAddr)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", c.ListenAddr)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("[!] accept error: %v\n", err)
			continue
		}
		go c.HandleConn(conn)
	}
}

// getValidSocks5Connection 获取可用的socks5连接并完成握手阶段
func (c *RedirectClient) getValidSocks5Connection() (net.Conn, error) {
	var cc net.Conn
	var err error
	if ProxyURL != "" {
		cc, err = net.DialTimeout("tcp", ProxyURL, 5*time.Second)
		if err == nil {
			fmt.Printf("[*] use %v\n", ProxyURL)
			return cc, nil
		}
		closeConn(cc)
		fmt.Printf("[!] cannot connect to %v\n", ProxyURL)
	}

	cc, err = net.DialTimeout("tcp", DefaultProxy, 5*time.Second)
	if err == nil {
		fmt.Printf("[*] use %v\n", DefaultProxy)
		return cc, nil
	}
	return nil, err
}

func (c *RedirectClient) HandleConn(conn net.Conn) {
	defer closeConn(conn)
	cc, err := c.getValidSocks5Connection()
	if err != nil {
		fmt.Printf("[!] getValidSocks5Connection failed: %v\n", err)
		return
	}
	defer closeConn(cc)
	err = transport(conn, cc)
	if err != nil {
		fmt.Printf("[!] transport error: %v\n", err)
	}
}

func closeConn(conn net.Conn) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
			fmt.Printf("[*] close connection: %v\n", err)
		}
	}()
	err = conn.Close()
	return err
}

func transport(rw1, rw2 io.ReadWriter) error {
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error{
		return copyBuffer(rw1, rw2)
	})

	g.Go(func() error{
		return copyBuffer(rw2, rw1)
	})
	var err error
	if err = g.Wait(); err != nil && err == io.EOF {
		err = nil
	}
	return err
}

func copyBuffer(dst io.Writer, src io.Reader) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("[!] copyBuffer: %v", e)
		}
	}()
	buf := make([]byte, largeBufferSize)

	_, err = CopyBufferWithCloseErr(dst, src, buf)
	return err
}
