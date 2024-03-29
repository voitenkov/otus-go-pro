package main

import (
	"fmt"
	"io"
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

// Place your code here.
// P.S. Author's solution takes no more than 50 lines.
type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func (tc *telnetClient) Connect() error {
	tc.conn, err = net.DialTimeout("tcp", tc.address, tc.timeout)
	if err != nil {
		return fmt.Errorf("cannot connect to %v: %w", tc.address, err)
	}
	fmt.Fprintln(os.Stderr, "...Connected to", tc.address)
	return nil
}

func (tc *telnetClient) Send() error {
	if tc.conn == nil {
		return ErrNoConnectionEstablished
	}

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
