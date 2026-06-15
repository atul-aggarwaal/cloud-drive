-- Version: 000003_alter_table_files_add_column.up.sql
alter table files add column updated_at timestamp not null default now();
