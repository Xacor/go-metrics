package interceptors

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func InitCheckSubnet(allowedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if allowedSubnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "unable to get request metadata")
		}

		cIP := net.ParseIP(md.Get("X-Real-IP")[0])
		if allowedSubnet.Contains(cIP) {
			return nil, status.Error(codes.PermissionDenied, "subnet not allowed")
		}

		return handler(ctx, req)
	}
}
