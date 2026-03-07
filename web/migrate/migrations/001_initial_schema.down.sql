DROP TRIGGER IF EXISTS update_agents_updated_at ON agents;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS agent_events;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS org_members;
DROP TABLE IF EXISTS user_oauth_links;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

DROP EXTENSION IF EXISTS "pgcrypto";
