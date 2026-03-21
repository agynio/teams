package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		_ = info
		identity, err := ExtractIdentity(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		return handler(WithIdentity(ctx, identity), req)
	}
}
