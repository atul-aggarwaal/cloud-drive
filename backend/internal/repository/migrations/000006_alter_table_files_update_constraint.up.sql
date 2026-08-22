-- 1. Update the parent files table tracking container
ALTER TABLE files DROP CONSTRAINT files_lifecycle_status_check;

ALTER TABLE files
ADD CONSTRAINT files_lifecycle_status_check
CHECK (
    lifecycle_status IN (
        'ACTIVE',
        'DELETE_REQUESTED',
        'DELETING',
        'DELETED'
    )
);

