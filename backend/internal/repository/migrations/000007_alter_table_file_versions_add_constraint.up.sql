ALTER TABLE file_versions 
ADD CONSTRAINT fk_unique_file_version 
UNIQUE (file_id, version_num);
