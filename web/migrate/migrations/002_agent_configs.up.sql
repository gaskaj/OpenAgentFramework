-- Agent type configuration templates (one per agent_type per org)
CREATE TABLE agent_type_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    agent_type      VARCHAR(50) NOT NULL,
    config          JSONB NOT NULL DEFAULT '{}',
    version         BIGINT NOT NULL DEFAULT 1,
    description     TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, agent_type)
);
CREATE INDEX idx_agent_type_configs_org ON agent_type_configs(org_id);
CREATE TRIGGER update_agent_type_configs_updated_at
    BEFORE UPDATE ON agent_type_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Per-agent configuration overrides
CREATE TABLE agent_config_overrides (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    config          JSONB NOT NULL DEFAULT '{}',
    version         BIGINT NOT NULL DEFAULT 1,
    description     TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id)
);
CREATE INDEX idx_agent_config_overrides_org ON agent_config_overrides(org_id);
CREATE INDEX idx_agent_config_overrides_agent ON agent_config_overrides(agent_id);
CREATE TRIGGER update_agent_config_overrides_updated_at
    BEFORE UPDATE ON agent_config_overrides
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Config change audit trail
CREATE TABLE config_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    target_type     VARCHAR(20) NOT NULL,
    target_id       UUID NOT NULL,
    changed_by      UUID REFERENCES users(id),
    previous_config JSONB,
    new_config      JSONB NOT NULL,
    version         BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_config_audit_org_time ON config_audit_log(org_id, created_at DESC);
