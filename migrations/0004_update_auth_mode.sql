-- +goose Up
-- +goose StatementBegin

-- Update auth_mode constraint to support 'basic' authentication
ALTER TABLE routes DROP CONSTRAINT IF EXISTS routes_auth_mode_check;
ALTER TABLE routes ADD CONSTRAINT routes_auth_mode_check CHECK (auth_mode IN ('none', 'api_key', 'basic'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to original constraint
ALTER TABLE routes DROP CONSTRAINT IF EXISTS routes_auth_mode_check;
ALTER TABLE routes ADD CONSTRAINT routes_auth_mode_check CHECK (auth_mode IN ('none', 'api_key', 'bearer'));

-- +goose StatementEnd