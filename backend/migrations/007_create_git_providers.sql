-- Create git_providers table for user-configurable Git OAuth providers
CREATE TABLE IF NOT EXISTS git_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,

    user_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    name TEXT NOT NULL,
    url TEXT,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,

    connected BOOLEAN DEFAULT FALSE,
    token TEXT,
    username TEXT,
    is_default BOOLEAN DEFAULT FALSE,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_git_providers_user_id ON git_providers(user_id);
CREATE INDEX IF NOT EXISTS idx_git_providers_deleted_at ON git_providers(deleted_at);
