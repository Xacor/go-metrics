package interceptors

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func InitVerifySignature(signKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if signKey == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "uanble to get request metadata")
		}

		data, err := json.Marshal(req)
		if err != nil {
			return nil, status.Error(codes.Internal, "unable to marshall request")
		}

		h := hmac.New(sha256.New, []byte(signKey))
		_, err = h.Write(data)
		if err != nil {
			return nil, status.Error(codes.Internal, "unable to calculate request hash")
		}

		csign := []byte(md.Get("HashSHA256")[0])
		ssign := h.Sum(nil)

		if !hmac.Equal(csign, ssign) {
			return nil, status.Error(codes.PermissionDenied, "signatures are't equal")
		}

		return handler(ctx, req)
	}
}
