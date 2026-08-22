ALTER TABLE file_versions 
ADD CONSTRAINT uq_file_versions_file_id_version_num 
UNIQUE (file_id, version_num);
