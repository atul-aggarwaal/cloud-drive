-- Version: 000002_create_file_shares_table.down.sql

DROP INDEX IF EXISTS idx_file_shares_user;
DROP TABLE IF EXISTS file_shares;