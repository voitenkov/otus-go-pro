package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

var (
	address                    string
	ctxMain                    context.Context
	stop                       context.CancelFunc
	timeout                    time.Duration
	err                        error
	ErrNoArguments             = errors.New("no arguments provided")
	ErrNoPortProvided          = errors.New("no port provided")
	ErrExtraArgumentsProvided  = errors.New("extra arguments provided")
	ErrPortNotNumeric          = errors.New("port is not numeric")
	ErrPortOutOfBounds         = errors.New("port should be from 1 to 65535")
	ErrTerminatedByUser        = errors.New("connection terminated by user")
	ErrTerminatedByServer      = errors.New("connection terminated by peer")
	ErrNoConnectionEstablished = errors.New("no connection established")
)

func init() {
	flag.DurationVar(&timeout, "timeout", time.Second*10, "connection timeout")
}

func main() {
	ctxMain, stop = signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctxMain.Done()
		fmt.Println(ctxMain.Err())
		os.Exit(1)
	}()

	flag.Parse()
	argsCount := len(flag.Args())

	switch {
	case argsCount == 0:
		err = ErrNoArguments
	case argsCount == 1:
		err = ErrNoPortProvided
	case argsCount == 2:
		port, converr := strconv.Atoi(flag.Arg(1))
		if converr != nil {
			err = ErrPortNotNumeric
		} else {
			if port < 1 || port > 65535 {
				err = ErrPortOutOfBounds
			}
		}
	case argsCount > 2:
		err = ErrExtraArgumentsProvided
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	address = net.JoinHostPort(flag.Arg(0), flag.Arg(1))
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)
	if client == nil {
		panic("telnet client failed")
	}

	err = client.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := client.Send()
		if err != nil {
			fmt.Fprintln(os.Stderr, "...EOF")
			fmt.Println(err)
			client.Close()
			os.Exit(1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := client.Receive()
		if err != nil {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
			fmt.Println(err)
			client.Close()
			os.Exit(1)
		}
	}()

	wg.Wait()
	client.Close()
}
