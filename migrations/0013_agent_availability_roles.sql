ALTER TABLE agents
    ADD COLUMN availability TEXT NOT NULL DEFAULT 'internal';

ALTER TABLE agents
    ADD CONSTRAINT agents_availability_check CHECK (availability IN ('internal', 'private'));

CREATE TABLE agent_roles (
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    identity_id UUID NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'maintainer', 'participant')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (agent_id, identity_id)
);

CREATE INDEX idx_agent_roles_identity ON agent_roles (identity_id);
