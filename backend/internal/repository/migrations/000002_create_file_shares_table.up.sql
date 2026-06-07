-- Version: 000002_create_file_shares_table.up.sql

CREATE TABLE IF NOT EXISTS file_shares (
    id VARCHAR(36) PRIMARY KEY,
    file_id VARCHAR(36) NOT NULL,
    shared_with_user_id VARCHAR(36) NOT NULL,
    permission_level VARCHAR(20) NOT NULL DEFAULT 'read', -- Options: 'read', 'read_write'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Enforce absolute multi-tenant relational integrity limits
    CONSTRAINT fk_share_file FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    CONSTRAINT fk_share_user FOREIGN KEY (shared_with_user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Compound Unique Index Constraint: Prevents duplicating the exact same user-to-file share pairing row block
    CONSTRAINT unique_user_file_share UNIQUE (file_id, shared_with_user_id)
);

CREATE INDEX IF NOT EXISTS idx_file_shares_user ON file_shares(shared_with_user_id);