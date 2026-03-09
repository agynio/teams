package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func scanAgent(row pgx.Row) (Agent, error) {
	var agent Agent
	var config []byte
	if err := row.Scan(
		&agent.Meta.ID,
		&agent.Title,
		&agent.Description,
		&config,
		&agent.Meta.CreatedAt,
		&agent.Meta.UpdatedAt,
	); err != nil {
		return Agent{}, err
	}
	if config == nil {
		return Agent{}, fmt.Errorf("agent config is null")
	}
	agent.Config = JSONData(config)
	return agent, nil
}

func scanTool(row pgx.Row) (Tool, error) {
	var tool Tool
	var config []byte
	if err := row.Scan(
		&tool.Meta.ID,
		&tool.Type,
		&tool.Name,
		&tool.Description,
		&config,
		&tool.Meta.CreatedAt,
		&tool.Meta.UpdatedAt,
	); err != nil {
		return Tool{}, err
	}
	if config == nil {
		return Tool{}, fmt.Errorf("tool config is null")
	}
	tool.Config = JSONData(config)
	return tool, nil
}

func scanMcpServer(row pgx.Row) (McpServer, error) {
	var server McpServer
	var config []byte
	if err := row.Scan(
		&server.Meta.ID,
		&server.Title,
		&server.Description,
		&config,
		&server.Meta.CreatedAt,
		&server.Meta.UpdatedAt,
	); err != nil {
		return McpServer{}, err
	}
	if config == nil {
		return McpServer{}, fmt.Errorf("mcp server config is null")
	}
	server.Config = JSONData(config)
	return server, nil
}

func scanWorkspaceConfiguration(row pgx.Row) (WorkspaceConfiguration, error) {
	var workspace WorkspaceConfiguration
	var config []byte
	if err := row.Scan(
		&workspace.Meta.ID,
		&workspace.Title,
		&workspace.Description,
		&config,
		&workspace.Meta.CreatedAt,
		&workspace.Meta.UpdatedAt,
	); err != nil {
		return WorkspaceConfiguration{}, err
	}
	if config == nil {
		return WorkspaceConfiguration{}, fmt.Errorf("workspace configuration config is null")
	}
	workspace.Config = JSONData(config)
	return workspace, nil
}

func scanMemoryBucket(row pgx.Row) (MemoryBucket, error) {
	var bucket MemoryBucket
	var config []byte
	if err := row.Scan(
		&bucket.Meta.ID,
		&bucket.Title,
		&bucket.Description,
		&config,
		&bucket.Meta.CreatedAt,
		&bucket.Meta.UpdatedAt,
	); err != nil {
		return MemoryBucket{}, err
	}
	if config == nil {
		return MemoryBucket{}, fmt.Errorf("memory bucket config is null")
	}
	bucket.Config = JSONData(config)
	return bucket, nil
}

func scanAttachment(row pgx.Row) (Attachment, error) {
	var attachment Attachment
	if err := row.Scan(
		&attachment.Meta.ID,
		&attachment.Kind,
		&attachment.SourceType,
		&attachment.SourceID,
		&attachment.TargetType,
		&attachment.TargetID,
		&attachment.Meta.CreatedAt,
		&attachment.Meta.UpdatedAt,
	); err != nil {
		return Attachment{}, err
	}
	return attachment, nil
}

func (s *Store) CreateAgent(ctx context.Context, input AgentInput) (Agent, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO agents (title, description, config)
         VALUES ($1, $2, $3)
         RETURNING id, title, description, config, created_at, updated_at`,
		input.Title,
		input.Description,
		input.Config,
	)
	agent, err := scanAgent(row)
	if err != nil {
		return Agent{}, err
	}
	return agent, nil
}

func (s *Store) GetAgent(ctx context.Context, id uuid.UUID) (Agent, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, title, description, config, created_at, updated_at
         FROM agents
         WHERE id = $1`,
		id,
	)
	agent, err := scanAgent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Agent{}, ErrAgentNotFound
		}
		return Agent{}, err
	}
	return agent, nil
}

func (s *Store) UpdateAgent(ctx context.Context, id uuid.UUID, update AgentUpdate) (Agent, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	paramIndex := 1

	if update.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramIndex))
		args = append(args, *update.Title)
		paramIndex++
	}
	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, *update.Description)
		paramIndex++
	}
	if update.Config != nil {
		setClauses = append(setClauses, fmt.Sprintf("config = $%d", paramIndex))
		args = append(args, *update.Config)
		paramIndex++
	}

	if len(setClauses) == 0 {
		return Agent{}, fmt.Errorf("agent update requires at least one field")
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE agents SET %s WHERE id = $%d RETURNING id, title, description, config, created_at, updated_at",
		strings.Join(setClauses, ", "),
		paramIndex,
	)
	row := s.pool.QueryRow(ctx, query, args...)
	agent, err := scanAgent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Agent{}, ErrAgentNotFound
		}
		return Agent{}, err
	}
	return agent, nil
}

func (s *Store) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM agents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAgentNotFound
	}
	return nil
}

func (s *Store) ListAgents(ctx context.Context, filter AgentFilter, pageSize int32, cursor *PageCursor) (AgentListResult, error) {
	limit := NormalizePageSize(pageSize)

	query := strings.Builder{}
	query.WriteString(`SELECT id, title, description, config, created_at, updated_at FROM agents`)

	args := make([]any, 0, 3)
	clauses := make([]string, 0, 2)
	paramIndex := 1

	if filter.Query != "" {
		clauses = append(clauses, fmt.Sprintf("(config->>'name' ILIKE $%d OR config->>'role' ILIKE $%d)", paramIndex, paramIndex))
		args = append(args, "%"+filter.Query+"%")
		paramIndex++
	}
	if cursor != nil {
		clauses = append(clauses, fmt.Sprintf("id::text > $%d", paramIndex))
		args = append(args, cursor.AfterID.String())
		paramIndex++
	}

	if len(clauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
	}
	query.WriteString(fmt.Sprintf(" ORDER BY id::text ASC LIMIT $%d", paramIndex))
	args = append(args, int(limit)+1)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return AgentListResult{}, err
	}
	defer rows.Close()

	agents := make([]Agent, 0, limit)
	var (
		nextCursor *PageCursor
		lastID     uuid.UUID
		hasMore    bool
	)
	for rows.Next() {
		if int32(len(agents)) == limit {
			hasMore = true
			break
		}
		agent, err := scanAgent(rows)
		if err != nil {
			return AgentListResult{}, err
		}
		agents = append(agents, agent)
		lastID = agent.Meta.ID
	}
	if err := rows.Err(); err != nil {
		return AgentListResult{}, err
	}
	if hasMore {
		nextCursor = &PageCursor{AfterID: lastID}
	}
	return AgentListResult{Agents: agents, NextCursor: nextCursor}, nil
}

func (s *Store) CreateTool(ctx context.Context, input ToolInput) (Tool, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO tools (type, name, description, config)
         VALUES ($1, $2, $3, $4)
         RETURNING id, type, name, description, config, created_at, updated_at`,
		input.Type,
		input.Name,
		input.Description,
		input.Config,
	)
	tool, err := scanTool(row)
	if err != nil {
		return Tool{}, err
	}
	return tool, nil
}

func (s *Store) GetTool(ctx context.Context, id uuid.UUID) (Tool, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, type, name, description, config, created_at, updated_at
         FROM tools
         WHERE id = $1`,
		id,
	)
	tool, err := scanTool(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tool{}, ErrToolNotFound
		}
		return Tool{}, err
	}
	return tool, nil
}

func (s *Store) UpdateTool(ctx context.Context, id uuid.UUID, update ToolUpdate) (Tool, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	paramIndex := 1

	if update.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", paramIndex))
		args = append(args, *update.Name)
		paramIndex++
	}
	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, *update.Description)
		paramIndex++
	}
	if update.Config != nil {
		setClauses = append(setClauses, fmt.Sprintf("config = $%d", paramIndex))
		args = append(args, *update.Config)
		paramIndex++
	}

	if len(setClauses) == 0 {
		return Tool{}, fmt.Errorf("tool update requires at least one field")
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE tools SET %s WHERE id = $%d RETURNING id, type, name, description, config, created_at, updated_at",
		strings.Join(setClauses, ", "),
		paramIndex,
	)
	row := s.pool.QueryRow(ctx, query, args...)
	tool, err := scanTool(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tool{}, ErrToolNotFound
		}
		return Tool{}, err
	}
	return tool, nil
}

func (s *Store) DeleteTool(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM tools WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrToolNotFound
	}
	return nil
}

func (s *Store) ListTools(ctx context.Context, filter ToolFilter, pageSize int32, cursor *PageCursor) (ToolListResult, error) {
	limit := NormalizePageSize(pageSize)

	query := strings.Builder{}
	query.WriteString(`SELECT id, type, name, description, config, created_at, updated_at FROM tools`)

	args := make([]any, 0, 3)
	clauses := make([]string, 0, 2)
	paramIndex := 1

	if filter.Type != nil {
		clauses = append(clauses, fmt.Sprintf("type = $%d", paramIndex))
		args = append(args, *filter.Type)
		paramIndex++
	}
	if cursor != nil {
		clauses = append(clauses, fmt.Sprintf("id::text > $%d", paramIndex))
		args = append(args, cursor.AfterID.String())
		paramIndex++
	}

	if len(clauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
	}
	query.WriteString(fmt.Sprintf(" ORDER BY id::text ASC LIMIT $%d", paramIndex))
	args = append(args, int(limit)+1)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return ToolListResult{}, err
	}
	defer rows.Close()

	tools := make([]Tool, 0, limit)
	var (
		nextCursor *PageCursor
		lastID     uuid.UUID
		hasMore    bool
	)
	for rows.Next() {
		if int32(len(tools)) == limit {
			hasMore = true
			break
		}
		tool, err := scanTool(rows)
		if err != nil {
			return ToolListResult{}, err
		}
		tools = append(tools, tool)
		lastID = tool.Meta.ID
	}
	if err := rows.Err(); err != nil {
		return ToolListResult{}, err
	}
	if hasMore {
		nextCursor = &PageCursor{AfterID: lastID}
	}
	return ToolListResult{Tools: tools, NextCursor: nextCursor}, nil
}

func (s *Store) CreateMcpServer(ctx context.Context, input McpServerInput) (McpServer, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO mcp_servers (title, description, config)
         VALUES ($1, $2, $3)
         RETURNING id, title, description, config, created_at, updated_at`,
		input.Title,
		input.Description,
		input.Config,
	)
	server, err := scanMcpServer(row)
	if err != nil {
		return McpServer{}, err
	}
	return server, nil
}

func (s *Store) GetMcpServer(ctx context.Context, id uuid.UUID) (McpServer, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, title, description, config, created_at, updated_at
         FROM mcp_servers
         WHERE id = $1`,
		id,
	)
	server, err := scanMcpServer(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return McpServer{}, ErrMcpServerNotFound
		}
		return McpServer{}, err
	}
	return server, nil
}

func (s *Store) UpdateMcpServer(ctx context.Context, id uuid.UUID, update McpServerUpdate) (McpServer, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	paramIndex := 1

	if update.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramIndex))
		args = append(args, *update.Title)
		paramIndex++
	}
	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, *update.Description)
		paramIndex++
	}
	if update.Config != nil {
		setClauses = append(setClauses, fmt.Sprintf("config = $%d", paramIndex))
		args = append(args, *update.Config)
		paramIndex++
	}

	if len(setClauses) == 0 {
		return McpServer{}, fmt.Errorf("mcp server update requires at least one field")
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE mcp_servers SET %s WHERE id = $%d RETURNING id, title, description, config, created_at, updated_at",
		strings.Join(setClauses, ", "),
		paramIndex,
	)
	row := s.pool.QueryRow(ctx, query, args...)
	server, err := scanMcpServer(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return McpServer{}, ErrMcpServerNotFound
		}
		return McpServer{}, err
	}
	return server, nil
}

func (s *Store) DeleteMcpServer(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM mcp_servers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrMcpServerNotFound
	}
	return nil
}

func (s *Store) ListMcpServers(ctx context.Context, pageSize int32, cursor *PageCursor) (McpServerListResult, error) {
	limit := NormalizePageSize(pageSize)

	query := strings.Builder{}
	query.WriteString(`SELECT id, title, description, config, created_at, updated_at FROM mcp_servers`)

	args := make([]any, 0, 2)
	paramIndex := 1
	if cursor != nil {
		query.WriteString(fmt.Sprintf(" WHERE id::text > $%d", paramIndex))
		args = append(args, cursor.AfterID.String())
		paramIndex++
	}
	query.WriteString(fmt.Sprintf(" ORDER BY id::text ASC LIMIT $%d", paramIndex))
	args = append(args, int(limit)+1)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return McpServerListResult{}, err
	}
	defer rows.Close()

	servers := make([]McpServer, 0, limit)
	var (
		nextCursor *PageCursor
		lastID     uuid.UUID
		hasMore    bool
	)
	for rows.Next() {
		if int32(len(servers)) == limit {
			hasMore = true
			break
		}
		server, err := scanMcpServer(rows)
		if err != nil {
			return McpServerListResult{}, err
		}
		servers = append(servers, server)
		lastID = server.Meta.ID
	}
	if err := rows.Err(); err != nil {
		return McpServerListResult{}, err
	}
	if hasMore {
		nextCursor = &PageCursor{AfterID: lastID}
	}
	return McpServerListResult{McpServers: servers, NextCursor: nextCursor}, nil
}

func (s *Store) CreateWorkspaceConfiguration(ctx context.Context, input WorkspaceConfigurationInput) (WorkspaceConfiguration, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO workspace_configurations (title, description, config)
         VALUES ($1, $2, $3)
         RETURNING id, title, description, config, created_at, updated_at`,
		input.Title,
		input.Description,
		input.Config,
	)
	workspace, err := scanWorkspaceConfiguration(row)
	if err != nil {
		return WorkspaceConfiguration{}, err
	}
	return workspace, nil
}

func (s *Store) GetWorkspaceConfiguration(ctx context.Context, id uuid.UUID) (WorkspaceConfiguration, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, title, description, config, created_at, updated_at
         FROM workspace_configurations
         WHERE id = $1`,
		id,
	)
	workspace, err := scanWorkspaceConfiguration(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspaceConfiguration{}, ErrWorkspaceConfigurationNotFound
		}
		return WorkspaceConfiguration{}, err
	}
	return workspace, nil
}

func (s *Store) UpdateWorkspaceConfiguration(ctx context.Context, id uuid.UUID, update WorkspaceConfigurationUpdate) (WorkspaceConfiguration, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	paramIndex := 1

	if update.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramIndex))
		args = append(args, *update.Title)
		paramIndex++
	}
	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, *update.Description)
		paramIndex++
	}
	if update.Config != nil {
		setClauses = append(setClauses, fmt.Sprintf("config = $%d", paramIndex))
		args = append(args, *update.Config)
		paramIndex++
	}

	if len(setClauses) == 0 {
		return WorkspaceConfiguration{}, fmt.Errorf("workspace configuration update requires at least one field")
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE workspace_configurations SET %s WHERE id = $%d RETURNING id, title, description, config, created_at, updated_at",
		strings.Join(setClauses, ", "),
		paramIndex,
	)
	row := s.pool.QueryRow(ctx, query, args...)
	workspace, err := scanWorkspaceConfiguration(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspaceConfiguration{}, ErrWorkspaceConfigurationNotFound
		}
		return WorkspaceConfiguration{}, err
	}
	return workspace, nil
}

func (s *Store) DeleteWorkspaceConfiguration(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM workspace_configurations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrWorkspaceConfigurationNotFound
	}
	return nil
}

func (s *Store) ListWorkspaceConfigurations(ctx context.Context, pageSize int32, cursor *PageCursor) (WorkspaceConfigurationListResult, error) {
	limit := NormalizePageSize(pageSize)

	query := strings.Builder{}
	query.WriteString(`SELECT id, title, description, config, created_at, updated_at FROM workspace_configurations`)

	args := make([]any, 0, 2)
	paramIndex := 1
	if cursor != nil {
		query.WriteString(fmt.Sprintf(" WHERE id::text > $%d", paramIndex))
		args = append(args, cursor.AfterID.String())
		paramIndex++
	}
	query.WriteString(fmt.Sprintf(" ORDER BY id::text ASC LIMIT $%d", paramIndex))
	args = append(args, int(limit)+1)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return WorkspaceConfigurationListResult{}, err
	}
	defer rows.Close()

	workspaces := make([]WorkspaceConfiguration, 0, limit)
	var (
		nextCursor *PageCursor
		lastID     uuid.UUID
		hasMore    bool
	)
	for rows.Next() {
		if int32(len(workspaces)) == limit {
			hasMore = true
			break
		}
		workspace, err := scanWorkspaceConfiguration(rows)
		if err != nil {
			return WorkspaceConfigurationListResult{}, err
		}
		workspaces = append(workspaces, workspace)
		lastID = workspace.Meta.ID
	}
	if err := rows.Err(); err != nil {
		return WorkspaceConfigurationListResult{}, err
	}
	if hasMore {
		nextCursor = &PageCursor{AfterID: lastID}
	}
	return WorkspaceConfigurationListResult{WorkspaceConfigurations: workspaces, NextCursor: nextCursor}, nil
}

func (s *Store) CreateMemoryBucket(ctx context.Context, input MemoryBucketInput) (MemoryBucket, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO memory_buckets (title, description, config)
         VALUES ($1, $2, $3)
         RETURNING id, title, description, config, created_at, updated_at`,
		input.Title,
		input.Description,
		input.Config,
	)
	bucket, err := scanMemoryBucket(row)
	if err != nil {
		return MemoryBucket{}, err
	}
	return bucket, nil
}

func (s *Store) GetMemoryBucket(ctx context.Context, id uuid.UUID) (MemoryBucket, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, title, description, config, created_at, updated_at
         FROM memory_buckets
         WHERE id = $1`,
		id,
	)
	bucket, err := scanMemoryBucket(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MemoryBucket{}, ErrMemoryBucketNotFound
		}
		return MemoryBucket{}, err
	}
	return bucket, nil
}

func (s *Store) UpdateMemoryBucket(ctx context.Context, id uuid.UUID, update MemoryBucketUpdate) (MemoryBucket, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	paramIndex := 1

	if update.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramIndex))
		args = append(args, *update.Title)
		paramIndex++
	}
	if update.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, *update.Description)
		paramIndex++
	}
	if update.Config != nil {
		setClauses = append(setClauses, fmt.Sprintf("config = $%d", paramIndex))
		args = append(args, *update.Config)
		paramIndex++
	}

	if len(setClauses) == 0 {
		return MemoryBucket{}, fmt.Errorf("memory bucket update requires at least one field")
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE memory_buckets SET %s WHERE id = $%d RETURNING id, title, description, config, created_at, updated_at",
		strings.Join(setClauses, ", "),
		paramIndex,
	)
	row := s.pool.QueryRow(ctx, query, args...)
	bucket, err := scanMemoryBucket(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MemoryBucket{}, ErrMemoryBucketNotFound
		}
		return MemoryBucket{}, err
	}
	return bucket, nil
}

func (s *Store) DeleteMemoryBucket(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM memory_buckets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrMemoryBucketNotFound
	}
	return nil
}

func (s *Store) ListMemoryBuckets(ctx context.Context, pageSize int32, cursor *PageCursor) (MemoryBucketListResult, error) {
	limit := NormalizePageSize(pageSize)

	query := strings.Builder{}
	query.WriteString(`SELECT id, title, description, config, created_at, updated_at FROM memory_buckets`)

	args := make([]any, 0, 2)
	paramIndex := 1
	if cursor != nil {
		query.WriteString(fmt.Sprintf(" WHERE id::text > $%d", paramIndex))
		args = append(args, cursor.AfterID.String())
		paramIndex++
	}
	query.WriteString(fmt.Sprintf(" ORDER BY id::text ASC LIMIT $%d", paramIndex))
	args = append(args, int(limit)+1)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return MemoryBucketListResult{}, err
	}
	defer rows.Close()

	buckets := make([]MemoryBucket, 0, limit)
	var (
		nextCursor *PageCursor
		lastID     uuid.UUID
		hasMore    bool
	)
	for rows.Next() {
		if int32(len(buckets)) == limit {
			hasMore = true
			break
		}
		bucket, err := scanMemoryBucket(rows)
		if err != nil {
			return MemoryBucketListResult{}, err
		}
		buckets = append(buckets, bucket)
		lastID = bucket.Meta.ID
	}
	if err := rows.Err(); err != nil {
		return MemoryBucketListResult{}, err
	}
	if hasMore {
		nextCursor = &PageCursor{AfterID: lastID}
	}
	return MemoryBucketListResult{MemoryBuckets: buckets, NextCursor: nextCursor}, nil
}

func (s *Store) CreateAttachment(ctx context.Context, input AttachmentInput) (Attachment, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO attachments (kind, source_type, source_id, target_type, target_id)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, kind, source_type, source_id, target_type, target_id, created_at, updated_at`,
		input.Kind,
		input.SourceType,
		input.SourceID,
		input.TargetType,
		input.TargetID,
	)
	attachment, err := scanAttachment(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Attachment{}, ErrAttachmentExists
		}
		return Attachment{}, err
	}
	return attachment, nil
}

func (s *Store) GetAttachment(ctx context.Context, id uuid.UUID) (Attachment, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, kind, source_type, source_id, target_type, target_id, created_at, updated_at
         FROM attachments
         WHERE id = $1`,
		id,
	)
	attachment, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Attachment{}, ErrAttachmentNotFound
		}
		return Attachment{}, err
	}
	return attachment, nil
}

func (s *Store) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

func (s *Store) ListAttachments(ctx context.Context, filter AttachmentFilter, pageSize int32, cursor *PageCursor) (AttachmentListResult, error) {
	limit := NormalizePageSize(pageSize)

	query := strings.Builder{}
	query.WriteString(`SELECT id, kind, source_type, source_id, target_type, target_id, created_at, updated_at FROM attachments`)

	args := make([]any, 0, 6)
	clauses := make([]string, 0, 5)
	paramIndex := 1

	if filter.Kind != nil {
		clauses = append(clauses, fmt.Sprintf("kind = $%d", paramIndex))
		args = append(args, *filter.Kind)
		paramIndex++
	}
	if filter.SourceType != nil {
		clauses = append(clauses, fmt.Sprintf("source_type = $%d", paramIndex))
		args = append(args, *filter.SourceType)
		paramIndex++
	}
	if filter.SourceID != nil {
		clauses = append(clauses, fmt.Sprintf("source_id = $%d", paramIndex))
		args = append(args, *filter.SourceID)
		paramIndex++
	}
	if filter.TargetType != nil {
		clauses = append(clauses, fmt.Sprintf("target_type = $%d", paramIndex))
		args = append(args, *filter.TargetType)
		paramIndex++
	}
	if filter.TargetID != nil {
		clauses = append(clauses, fmt.Sprintf("target_id = $%d", paramIndex))
		args = append(args, *filter.TargetID)
		paramIndex++
	}
	if cursor != nil {
		clauses = append(clauses, fmt.Sprintf("id::text > $%d", paramIndex))
		args = append(args, cursor.AfterID.String())
		paramIndex++
	}

	if len(clauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
	}
	query.WriteString(fmt.Sprintf(" ORDER BY id::text ASC LIMIT $%d", paramIndex))
	args = append(args, int(limit)+1)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return AttachmentListResult{}, err
	}
	defer rows.Close()

	attachments := make([]Attachment, 0, limit)
	var (
		nextCursor *PageCursor
		lastID     uuid.UUID
		hasMore    bool
	)
	for rows.Next() {
		if int32(len(attachments)) == limit {
			hasMore = true
			break
		}
		attachment, err := scanAttachment(rows)
		if err != nil {
			return AttachmentListResult{}, err
		}
		attachments = append(attachments, attachment)
		lastID = attachment.Meta.ID
	}
	if err := rows.Err(); err != nil {
		return AttachmentListResult{}, err
	}
	if hasMore {
		nextCursor = &PageCursor{AfterID: lastID}
	}
	return AttachmentListResult{Attachments: attachments, NextCursor: nextCursor}, nil
}
