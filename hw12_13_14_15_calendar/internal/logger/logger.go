package logger

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Logger struct{}

func New(level string) *Logger {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalf("failed to parse the level: %v", err)
	}

	logrus.SetLevel(logLevel)
	return &Logger{}
}

func (l Logger) Info(msg ...interface{}) {
	logrus.Info(msg...)
}

func (l Logger) Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func (l Logger) Error(msg ...interface{}) {
	logrus.Error(msg...)
}

func (l Logger) Warn(msg ...interface{}) {
	logrus.Warn(msg...)
}

func (l Logger) Debug(msg ...interface{}) {
	logrus.Debug(msg...)
}

func (l Logger) LogHTTPRequest(request *http.Request, duration time.Duration, statusCode int) {
	ip, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		logrus.Errorf("error splitting host and port: %q", request.RemoteAddr)
	}

	formatString := "%s [%s] %s %s %s %d %d %s"
	t := time.Now().Format("02/Jan/2006:15:04:05 -0700")
	logrus.Infof(formatString, ip, t, request.Method, request.URL.String(), request.Proto, statusCode,
		duration.Microseconds(), request.UserAgent())
}

func (l Logger) LogGRPCRequest(ctx context.Context, info *grpc.UnaryServerInfo, duration time.Duration,
	statusCode string,
) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		logrus.Errorf("error receiving peer information: %s", peer.Addr.String())
	}

	userAgent := ""
	metadata, ok := metadata.FromIncomingContext(ctx)
	if ok {
		userAgent = strings.Join(metadata["user-agent"], " ")
	}

	ip, _, err := net.SplitHostPort(peer.Addr.String())
	if err != nil {
		logrus.Errorf("error splitting host and port: %q", peer.Addr.String())
	}

	formatString := "%s [%s] %s %s %s %s %d %s"
	t := time.Now().Format("02/Jan/2006:15:04:05 -0700")
	logrus.Infof(formatString, ip, t, "rpc", info.FullMethod, "HTTP/2", statusCode, duration.Microseconds(), userAgent)
}
