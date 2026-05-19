package server

import (
	"context"
	"strings"
	"testing"
	"time"

	agentsv1 "github.com/agynio/agents/.gen/go/agynio/api/agents/v1"
	authorizationv1 "github.com/agynio/agents/.gen/go/agynio/api/authorization/v1"
	identityv1 "github.com/agynio/agents/.gen/go/agynio/api/identity/v1"
	notificationsv1 "github.com/agynio/agents/.gen/go/agynio/api/notifications/v1"
	"github.com/agynio/agents/internal/store"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToProtoVolumeIncludesTTL(t *testing.T) {
	ttl := "24h"
	volume := store.Volume{
		Meta: store.EntityMeta{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Persistent:  true,
		MountPath:   "/data",
		Size:        "1Gi",
		Description: "volume",
		TTL:         &ttl,
	}

	protoVolume := toProtoVolume(volume)
	if protoVolume.Ttl == nil {
		t.Fatalf("expected ttl to be set")
	}
	if protoVolume.GetTtl() != ttl {
		t.Fatalf("expected ttl %q, got %q", ttl, protoVolume.GetTtl())
	}
}

func TestToProtoVolumeOmitsTTLWhenNil(t *testing.T) {
	volume := store.Volume{
		Meta: store.EntityMeta{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Persistent:  true,
		MountPath:   "/data",
		Size:        "1Gi",
		Description: "volume",
	}

	protoVolume := toProtoVolume(volume)
	if protoVolume.Ttl != nil {
		t.Fatalf("expected ttl to be nil")
	}
}

func TestCreateAgentValidatesAvailabilityBeforeIdentity(t *testing.T) {
	server := New(&store.Store{}, noopAuthorizationWriter{}, noopIdentityWriter{}, noopNotificationsClient{})

	_, err := server.CreateAgent(context.Background(), &agentsv1.CreateAgentRequest{
		OrganizationId: uuid.NewString(),
		Model:          uuid.NewString(),
		InitImage:      "ghcr.io/agynio/agent-init-codex:latest",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", err)
	}
	if !strings.Contains(err.Error(), "availability: must be internal or private") {
		t.Fatalf("expected availability error, got %v", err)
	}
}

type noopAuthorizationWriter struct{}

func (noopAuthorizationWriter) Check(context.Context, *authorizationv1.CheckRequest, ...grpc.CallOption) (*authorizationv1.CheckResponse, error) {
	return &authorizationv1.CheckResponse{Allowed: true}, nil
}

func (noopAuthorizationWriter) Write(context.Context, *authorizationv1.WriteRequest, ...grpc.CallOption) (*authorizationv1.WriteResponse, error) {
	return &authorizationv1.WriteResponse{}, nil
}

type noopIdentityWriter struct{}

func (noopIdentityWriter) RegisterIdentity(context.Context, *identityv1.RegisterIdentityRequest, ...grpc.CallOption) (*identityv1.RegisterIdentityResponse, error) {
	return &identityv1.RegisterIdentityResponse{}, nil
}

func (noopIdentityWriter) SetNickname(context.Context, *identityv1.SetNicknameRequest, ...grpc.CallOption) (*identityv1.SetNicknameResponse, error) {
	return &identityv1.SetNicknameResponse{}, nil
}

func (noopIdentityWriter) RemoveNickname(context.Context, *identityv1.RemoveNicknameRequest, ...grpc.CallOption) (*identityv1.RemoveNicknameResponse, error) {
	return &identityv1.RemoveNicknameResponse{}, nil
}

type noopNotificationsClient struct{}

func (noopNotificationsClient) Publish(context.Context, *notificationsv1.PublishRequest, ...grpc.CallOption) (*notificationsv1.PublishResponse, error) {
	return &notificationsv1.PublishResponse{}, nil
}

func (noopNotificationsClient) Subscribe(context.Context, *notificationsv1.SubscribeRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[notificationsv1.SubscribeResponse], error) {
	return nil, status.Error(codes.Unimplemented, "subscribe")
}
