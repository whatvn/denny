package grpc

import (
	"context"
	"github.com/whatvn/denny/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := log.New(&log.JSONFormatter{})
	start := time.Now()
	panicking := true

	logger.WithFields(map[string]interface{}{
		"start": start,
		"Uri":   info.FullMethod,
	})

	defer func() {
		// We don't want to recover anything, but we want to log Internal error
		// in case of a panic. We pray here reportServerRPCMetrics is very
		// lightweight and it doesn't panic itself.
		code := codes.OK
		switch {
		case err != nil:
			code = status.Code(err)
		case panicking:
			code = codes.Internal
		}
		logger.WithField("code", code)
		logger.Infof("Finish in %v", time.Now().Sub(start))

	}()

	ctx = context.WithValue(ctx, log.LogKey, logger)
	resp, err = handler(ctx, req)
	panicking = false // normal exit, no panic happened, disarms defer
	return
}
