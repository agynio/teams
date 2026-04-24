package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type agentIDQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func touchAgentUpdatedAt(ctx context.Context, tx pgx.Tx, agentID uuid.UUID) error {
	result, err := tx.Exec(ctx, "UPDATE agents SET updated_at = NOW() WHERE id = $1", agentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return NotFound("agent")
	}
	return nil
}

func touchAgentsUpdatedAt(ctx context.Context, tx pgx.Tx, agentIDs []uuid.UUID) error {
	for _, agentID := range agentIDs {
		if err := touchAgentUpdatedAt(ctx, tx, agentID); err != nil {
			return err
		}
	}
	return nil
}

func resolveAgentID(ctx context.Context, tx pgx.Tx, agentID *uuid.UUID, mcpID *uuid.UUID, hookID *uuid.UUID) (uuid.UUID, error) {
	if agentID != nil {
		return *agentID, nil
	}
	if mcpID != nil {
		return agentIDForMcp(ctx, tx, *mcpID)
	}
	if hookID != nil {
		return agentIDForHook(ctx, tx, *hookID)
	}
	return uuid.UUID{}, fmt.Errorf("missing target identifier")
}

func agentIDForMcp(ctx context.Context, tx pgx.Tx, mcpID uuid.UUID) (uuid.UUID, error) {
	var agentID uuid.UUID
	row := tx.QueryRow(ctx, "SELECT agent_id FROM mcps WHERE id = $1", mcpID)
	if err := row.Scan(&agentID); err != nil {
		if err == pgx.ErrNoRows {
			return uuid.UUID{}, NotFound("mcp")
		}
		return uuid.UUID{}, err
	}
	return agentID, nil
}

func agentIDForHook(ctx context.Context, tx pgx.Tx, hookID uuid.UUID) (uuid.UUID, error) {
	var agentID uuid.UUID
	row := tx.QueryRow(ctx, "SELECT agent_id FROM hooks WHERE id = $1", hookID)
	if err := row.Scan(&agentID); err != nil {
		if err == pgx.ErrNoRows {
			return uuid.UUID{}, NotFound("hook")
		}
		return uuid.UUID{}, err
	}
	return agentID, nil
}

func agentIDsForVolume(ctx context.Context, queryer agentIDQueryer, volumeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := queryer.Query(ctx,
		`SELECT DISTINCT COALESCE(va.agent_id, mcps.agent_id, hooks.agent_id)
FROM volume_attachments va
LEFT JOIN mcps ON va.mcp_id = mcps.id
LEFT JOIN hooks ON va.hook_id = hooks.id
WHERE va.volume_id = $1`,
		volumeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := make([]uuid.UUID, 0)
	seen := map[uuid.UUID]struct{}{}
	for rows.Next() {
		var agentID uuid.UUID
		if err := rows.Scan(&agentID); err != nil {
			return nil, err
		}
		if _, ok := seen[agentID]; ok {
			continue
		}
		seen[agentID] = struct{}{}
		agents = append(agents, agentID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return agents, nil
}
