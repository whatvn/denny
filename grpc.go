package denny

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

// ChainUnaryServerInterceptors chains multiple interceptors together.
//
// The first one becomes the outermost, and the last one becomes the
// innermost, i.e. `ChainUnaryServerInterceptors(a, b, c)(h) === a(b(c(h)))`.
//
// nil-valued interceptors are silently skipped.
func chainUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	switch {
	case len(interceptors) == 0:
		// Noop interceptor.
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	case interceptors[0] == nil:
		// Skip nils.
		return chainUnaryServerInterceptors(interceptors[1:]...)
	case len(interceptors) == 1:
		// No need to actually chain anything.
		return interceptors[0]
	default:
		return combinator(interceptors[0], chainUnaryServerInterceptors(interceptors[1:]...))
	}
}

// combinator is an interceptor that chains just two interceptors together.
func combinator(first, second grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return first(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
			return second(ctx, req, info, handler)
		})
	}
}

func NewGrpcServer(interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	//var (
	//	builtinInterceptors = []grpc.UnaryServerInterceptor{
	//		//grpc_middleware.LoggerInterceptor,
	//	}
	//)
	interceptor := grpc_opentracing.UnaryServerInterceptor()
	return grpc.NewServer(grpc.UnaryInterceptor(interceptor))
}
