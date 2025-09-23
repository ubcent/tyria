-- +goose Up
-- +goose StatementBegin

-- Update user role enum to support owner, admin, viewer roles
ALTER TABLE users 
DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users 
ADD CONSTRAINT users_role_check 
CHECK (role IN ('owner', 'admin', 'viewer'));

-- Update existing users to use new role system
-- Convert 'admin' to 'owner' for first user, others to 'admin'
UPDATE users 
SET role = CASE 
  WHEN id = (SELECT MIN(id) FROM users WHERE role = 'admin') THEN 'owner'
  WHEN role = 'admin' THEN 'admin'
  WHEN role = 'user' THEN 'viewer'
  ELSE 'viewer'
END;

-- Set default role to 'viewer' for new users
ALTER TABLE users 
ALTER COLUMN role SET DEFAULT 'viewer';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert role enum back to original
ALTER TABLE users 
DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users 
ADD CONSTRAINT users_role_check 
CHECK (role IN ('admin', 'user'));

-- Convert roles back
UPDATE users 
SET role = CASE 
  WHEN role = 'owner' THEN 'admin'
  WHEN role = 'admin' THEN 'admin'
  WHEN role = 'viewer' THEN 'user'
  ELSE 'user'
END;

-- Set default role back to 'admin'
ALTER TABLE users 
ALTER COLUMN role SET DEFAULT 'admin';

-- +goose StatementEnd