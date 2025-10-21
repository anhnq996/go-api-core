-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop index
DROP INDEX IF EXISTS idx_users_deleted_at;

-- Drop table
DROP TABLE IF EXISTS users;

