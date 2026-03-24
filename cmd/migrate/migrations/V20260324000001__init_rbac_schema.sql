-- RBAC baseline schema (portable SQL flavor).
-- Naming follows dbx/migrate convention: V<version>__<description>.sql

CREATE TABLE users (
  id              VARCHAR(64) PRIMARY KEY,
  username        VARCHAR(64) NOT NULL UNIQUE,
  email           VARCHAR(255) NOT NULL UNIQUE,
  password_hash   VARCHAR(255) NOT NULL,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE roles (
  id              VARCHAR(64) PRIMARY KEY,
  code            VARCHAR(64) NOT NULL UNIQUE,
  name            VARCHAR(128) NOT NULL,
  description     VARCHAR(512),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permissions (
  id              VARCHAR(64) PRIMARY KEY,
  resource        VARCHAR(128) NOT NULL,
  action          VARCHAR(64) NOT NULL,
  effect          VARCHAR(16) NOT NULL DEFAULT 'allow',
  description     VARCHAR(512),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (resource, action, effect)
);

CREATE TABLE role_permissions (
  role_id         VARCHAR(64) NOT NULL,
  permission_id   VARCHAR(64) NOT NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE TABLE user_roles (
  user_id         VARCHAR(64) NOT NULL,
  role_id         VARCHAR(64) NOT NULL,
  scope_type      VARCHAR(64) NOT NULL DEFAULT 'global',
  scope_id        VARCHAR(128),
  expires_at      TIMESTAMP NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, role_id, scope_type, scope_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE user_permissions (
  user_id         VARCHAR(64) NOT NULL,
  permission_id   VARCHAR(64) NOT NULL,
  effect          VARCHAR(16) NOT NULL DEFAULT 'allow',
  scope_type      VARCHAR(64) NOT NULL DEFAULT 'global',
  scope_id        VARCHAR(128),
  expires_at      TIMESTAMP NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, permission_id, scope_type, scope_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_roles_lookup
  ON user_roles (user_id, scope_type, scope_id);

CREATE INDEX idx_role_permissions_lookup
  ON role_permissions (role_id, permission_id);

CREATE INDEX idx_user_permissions_lookup
  ON user_permissions (user_id, scope_type, scope_id);
