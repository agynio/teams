package server

import (
	"context"

	identityv1 "github.com/agynio/agents/.gen/go/agynio/api/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const identityServiceDeregisterIdentityFullMethodName = "/agynio.api.identity.v1.IdentityService/DeregisterIdentity"

type IdentityWriter interface {
	RegisterIdentity(ctx context.Context, req *identityv1.RegisterIdentityRequest, opts ...grpc.CallOption) (*identityv1.RegisterIdentityResponse, error)
	DeregisterIdentity(ctx context.Context, identityID string, opts ...grpc.CallOption) error
}

type identityWriter struct {
	conn   grpc.ClientConnInterface
	client identityv1.IdentityServiceClient
}

func NewIdentityWriter(conn grpc.ClientConnInterface) IdentityWriter {
	if conn == nil {
		panic("identity connection is required")
	}
	return &identityWriter{conn: conn, client: identityv1.NewIdentityServiceClient(conn)}
}

func (w *identityWriter) RegisterIdentity(ctx context.Context, req *identityv1.RegisterIdentityRequest, opts ...grpc.CallOption) (*identityv1.RegisterIdentityResponse, error) {
	return w.client.RegisterIdentity(ctx, req, opts...)
}

func (w *identityWriter) DeregisterIdentity(ctx context.Context, identityID string, opts ...grpc.CallOption) error {
	req := &identityv1.GetIdentityTypeRequest{IdentityId: identityID}
	resp := new(emptypb.Empty)
	return w.conn.Invoke(ctx, identityServiceDeregisterIdentityFullMethodName, req, resp, opts...)
}
