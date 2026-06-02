CREATE UNIQUE INDEX IF NOT EXISTS users_single_groom ON users (role) WHERE role = 'groom';
CREATE UNIQUE INDEX IF NOT EXISTS users_single_bride ON users (role) WHERE role = 'bride';
