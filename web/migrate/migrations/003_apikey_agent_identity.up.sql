-- Add agent identity columns to api_keys so each key is bound to an agent name/type.
ALTER TABLE api_keys ADD COLUMN agent_type VARCHAR(64) NOT NULL DEFAULT 'developer';
ALTER TABLE api_keys ADD COLUMN agent_name VARCHAR(255) NOT NULL DEFAULT '';
