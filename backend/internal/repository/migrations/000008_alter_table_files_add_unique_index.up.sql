
CREATE UNIQUE INDEX uq_files_owner_id_name 
ON files(owner_id, file_name)
WHERE lifecycle_status = 'ACTIVE'