package grpc

import (
	"context"

	"github.com/whatvn/denny/middleware"
	"google.golang.org/grpc"
)

func ValidatorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if v, ok := req.(middleware.IValidator); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	resp, err = handler(ctx, req)
	return
}
