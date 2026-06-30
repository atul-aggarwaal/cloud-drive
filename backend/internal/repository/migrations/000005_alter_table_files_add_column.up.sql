-- 1. Update the parent files table tracking container
ALTER TABLE files
ADD COLUMN lifecycle_status TEXT
CHECK (
    lifecycle_status IN (
        'ACTIVE',
        'DELETE_REQUESTED',
        'DELETED'
    )
)
NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE files ADD COLUMN created_by VARCHAR(255) DEFAULT NULL;
ALTER TABLE files ADD COLUMN updated_by VARCHAR(255) DEFAULT NULL;
ALTER TABLE files ADD COLUMN delete_requested_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;
ALTER TABLE files ADD COLUMN deleted_requested_by VARCHAR(255) DEFAULT NULL;

CREATE INDEX idx_files_delete_requested
ON files(lifecycle_status, delete_requested_at);
