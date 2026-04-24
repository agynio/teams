package server

import (
	"context"
	"fmt"
	"log"

	notificationsv1 "github.com/agynio/agents/.gen/go/agynio/api/notifications/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

const agentUpdatedEvent = "agent.updated"

func (s *Server) publishAgentUpdated(ctx context.Context, agentID uuid.UUID, organizationID uuid.UUID) {
	if s.notifications == nil {
		return
	}
	payload, err := structpb.NewStruct(map[string]any{
		"agent_id":        agentID.String(),
		"organization_id": organizationID.String(),
	})
	if err != nil {
		log.Printf("agents: build agent.updated payload: %v", err)
		return
	}
	_, err = s.notifications.Publish(ctx, &notificationsv1.PublishRequest{
		Event:   agentUpdatedEvent,
		Rooms:   []string{fmt.Sprintf("agent:%s", agentID)},
		Payload: payload,
		Source:  "agents",
	})
	if err != nil {
		log.Printf("agents: publish agent.updated: %v", err)
	}
}

func (s *Server) publishAgentUpdatedByID(ctx context.Context, agentID uuid.UUID) {
	if s.notifications == nil {
		return
	}
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		log.Printf("agents: fetch agent for notification: %v", err)
		return
	}
	s.publishAgentUpdated(ctx, agent.Meta.ID, agent.OrganizationID)
}

func (s *Server) resolveAgentID(ctx context.Context, agentID *uuid.UUID, mcpID *uuid.UUID, hookID *uuid.UUID) (uuid.UUID, error) {
	if agentID != nil {
		return *agentID, nil
	}
	if mcpID != nil {
		mcp, err := s.store.GetMcp(ctx, *mcpID)
		if err != nil {
			return uuid.UUID{}, err
		}
		return mcp.AgentID, nil
	}
	if hookID != nil {
		hook, err := s.store.GetHook(ctx, *hookID)
		if err != nil {
			return uuid.UUID{}, err
		}
		return hook.AgentID, nil
	}
	return uuid.UUID{}, fmt.Errorf("missing target identifier")
}

func (s *Server) publishAgentUpdatedForVolume(ctx context.Context, volumeID uuid.UUID) {
	if s.notifications == nil {
		return
	}
	agentIDs, err := s.store.ListAgentIDsForVolume(ctx, volumeID)
	if err != nil {
		log.Printf("agents: list volume agents: %v", err)
		return
	}
	for _, agentID := range agentIDs {
		s.publishAgentUpdatedByID(ctx, agentID)
	}
}

func (s *Server) publishAgentUpdatedForTarget(ctx context.Context, agentID *uuid.UUID, mcpID *uuid.UUID, hookID *uuid.UUID) {
	if s.notifications == nil {
		return
	}
	resolvedID, err := s.resolveAgentID(ctx, agentID, mcpID, hookID)
	if err != nil {
		log.Printf("agents: resolve agent for notification: %v", err)
		return
	}
	s.publishAgentUpdatedByID(ctx, resolvedID)
}
