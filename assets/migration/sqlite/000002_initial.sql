CREATE TABLE users (
    user_id TEXT NOT NULL PRIMARY KEY,
    user_active BOOLEAN NOT NULL DEFAULT true,
    user_created INTEGER NOT NULL,
    user_modified INTEGER,
    user_email TEXT,
    user_full_name TEXT NOT NULL,
    user_is_admin BOOLEAN NOT NULL DEFAULT false,
    user_date_joined INTEGER NOT NULL,
    user_last_login INTEGER,
    user_salt TEXT NOT NULL,
    user_hashed_password TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS ndx_user_email ON users(LOWER(user_email));
CREATE INDEX IF NOT EXISTS ndx_user_full_name ON users (user_full_name);
CREATE INDEX IF NOT EXISTS ndx_user_hashed_password ON users (user_hashed_password);
CREATE TABLE roles(
    role_id TEXT PRIMARY KEY,
    role_created INTEGER NOT NULL,
    role_modified INTEGER,
    role_name TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ndx_role_name ON roles(LOWER(role_name));
CREATE TABLE user_roles(
    user_role_user_id TEXT,
    user_role_role_id TEXT,
    user_role_created INTEGER NOT NULL,
    PRIMARY KEY (user_role_user_id, user_role_role_id),
    CONSTRAINT fk_users_user_id FOREIGN KEY (user_role_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_roles_role_id FOREIGN KEY (user_role_role_id) REFERENCES roles(role_id) ON DELETE CASCADE
);
CREATE TABLE permissions(
    permission_user_id TEXT,
    permission_role_id TEXT,
    permission_id TEXT NOT NULL,
    permission_resource_id TEXT,
    permission_created INTEGER NOT NULL,
    PRIMARY KEY (
        permission_user_id,
        permission_role_id,
        permission_id,
        permission_resource_id
    ),
    CONSTRAINT fk_permission_user FOREIGN KEY (permission_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_permission_role FOREIGN KEY (permission_role_id) REFERENCES roles(role_id) ON DELETE CASCADE
);