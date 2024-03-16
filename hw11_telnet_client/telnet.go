package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func (tc *telnetClient) Connect() error {
	_, err := net.ResolveTCPAddr("tcp4", tc.address)
	if err != nil {
		log.Fatal(err)
	}

	dialer := &net.Dialer{}
	if ctxMain == nil {
		ctxMain = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctxMain, tc.timeout)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "tcp4", tc.address)
	if err != nil {
		return fmt.Errorf("cannot connect to %v: %w", tc.address, err)
	}
	fmt.Fprintln(os.Stderr, "...Connected to", tc.address)
	tc.conn = conn
	return nil
}

func (tc *telnetClient) Send() error {
	if tc.conn == nil {
		return ErrNoConnectionEstablished
	}

	// tc.conn.SetWriteDeadline(time.Now().Add(time.Second * 1))
	_, err := io.Copy(tc.conn, tc.in)
	if err != nil {
		return ErrTerminatedByUser
	}

	return nil
}

func (tc *telnetClient) Receive() error {
	if tc.conn == nil {
		return ErrNoConnectionEstablished
	}

	_, err := io.Copy(tc.out, tc.conn)
	if err != nil {
		return ErrTerminatedByServer
	}

	return nil
}

func (tc *telnetClient) Close() error {
	if tc.conn == nil {
		return ErrNoConnectionEstablished
	}
	return tc.conn.Close()
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

// Place your code here.
// P.S. Author's solution takes no more than 50 lines.
