-- Record whether a project originated in the application or through MCP.
ALTER TABLE user_projects ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'app';

ALTER TABLE user_projects DROP CONSTRAINT IF EXISTS user_projects_source_check;
ALTER TABLE user_projects ADD CONSTRAINT user_projects_source_check
    CHECK (source IN ('app', 'mcp'));
