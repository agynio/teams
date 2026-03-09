package store

import "errors"

var (
	ErrAgentNotFound                  = errors.New("agent not found")
	ErrToolNotFound                   = errors.New("tool not found")
	ErrMcpServerNotFound              = errors.New("mcp server not found")
	ErrWorkspaceConfigurationNotFound = errors.New("workspace configuration not found")
	ErrMemoryBucketNotFound           = errors.New("memory bucket not found")
	ErrAttachmentNotFound             = errors.New("attachment not found")
	ErrAttachmentExists               = errors.New("attachment already exists")
)
