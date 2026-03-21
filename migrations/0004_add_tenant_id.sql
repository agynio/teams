ALTER TABLE agents
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE agents
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX agents_tenant_id_idx ON agents (tenant_id);

ALTER TABLE volumes
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE volumes
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX volumes_tenant_id_idx ON volumes (tenant_id);

ALTER TABLE mcps
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE mcps
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX mcps_tenant_agent_idx ON mcps (tenant_id, agent_id);

ALTER TABLE skills
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE skills
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX skills_tenant_agent_idx ON skills (tenant_id, agent_id);

ALTER TABLE hooks
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE hooks
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX hooks_tenant_agent_idx ON hooks (tenant_id, agent_id);

ALTER TABLE volume_attachments
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE volume_attachments
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX volume_attachments_tenant_id_idx ON volume_attachments (tenant_id);

ALTER TABLE envs
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE envs
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX envs_tenant_id_idx ON envs (tenant_id);

ALTER TABLE init_scripts
    ADD COLUMN tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE init_scripts
    ALTER COLUMN tenant_id DROP DEFAULT;
CREATE INDEX init_scripts_tenant_id_idx ON init_scripts (tenant_id);
