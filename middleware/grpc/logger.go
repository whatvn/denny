package grpc

import (
	"context"
	"github.com/whatvn/denny/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"time"
)

func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var (
		logger    = log.New(&log.JSONFormatter{})
		start     = time.Now()
		panicking = true
	)

	p, ok := peer.FromContext(ctx)
	if ok {
		logger.WithField("request_ip", p.Addr.String())
	}
	logger.WithFields(map[string]interface{}{
		"start":   start,
		"uri":     info.FullMethod,
		"request": logger.ToJsonString(req),
	})

	defer func() {
		code := codes.OK
		switch {
		case err != nil:
			code = status.Code(err)
		case panicking:
			code = codes.Internal
		}
		logger.WithField("code", code.String())
		logger.Infof("latency: ", time.Now().Sub(start))

	}()

	ctx = context.WithValue(ctx, log.LogKey, logger)
	resp, err = handler(ctx, req)
	panicking = false // normal exit, no panic happened, disarms defer
	return
}
