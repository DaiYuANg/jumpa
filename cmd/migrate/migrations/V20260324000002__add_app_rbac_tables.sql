CREATE TABLE app_roles (
  id              VARCHAR(64) PRIMARY KEY,
  name            VARCHAR(128) NOT NULL,
  description     VARCHAR(512),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app_permission_groups (
  id              VARCHAR(64) PRIMARY KEY,
  name            VARCHAR(128) NOT NULL,
  description     VARCHAR(512),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app_permissions (
  id              VARCHAR(64) PRIMARY KEY,
  name            VARCHAR(128) NOT NULL,
  code            VARCHAR(128) NOT NULL UNIQUE,
  group_id        VARCHAR(64),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app_role_permission_groups (
  role_id               VARCHAR(64) NOT NULL,
  permission_group_id   VARCHAR(64) NOT NULL,
  PRIMARY KEY (role_id, permission_group_id)
);

CREATE TABLE app_user_roles (
  user_id          BIGINT NOT NULL,
  role_id          VARCHAR(64) NOT NULL,
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE app_auth_principals (
  id               VARCHAR(64) PRIMARY KEY,
  email            VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE app_auth_principal_roles (
  principal_id     VARCHAR(64) NOT NULL,
  role             VARCHAR(128) NOT NULL,
  PRIMARY KEY (principal_id, role)
);

CREATE TABLE app_auth_principal_permissions (
  principal_id     VARCHAR(64) NOT NULL,
  permission       VARCHAR(128) NOT NULL,
  PRIMARY KEY (principal_id, permission)
);

INSERT INTO app_auth_principals (id, email) VALUES
  ('u-admin', 'admin@example.com'),
  ('u-readonly', 'readonly@example.com'),
  ('u-users', 'users@example.com'),
  ('u-roles', 'roles@example.com'),
  ('u-guest', 'guest@example.com');

INSERT INTO app_auth_principal_roles (principal_id, role) VALUES
  ('u-admin', 'admin'),
  ('u-readonly', 'readonly'),
  ('u-users', 'users-manager'),
  ('u-roles', 'roles-manager'),
  ('u-guest', 'guest');

INSERT INTO app_auth_principal_permissions (principal_id, permission) VALUES
  ('u-admin', 'users:read'),
  ('u-admin', 'users:write'),
  ('u-admin', 'roles:read'),
  ('u-admin', 'roles:write'),
  ('u-admin', 'permissions:read'),
  ('u-admin', 'permissions:write'),
  ('u-admin', 'permission-groups:read'),
  ('u-admin', 'permission-groups:write'),
  ('u-readonly', 'users:read'),
  ('u-readonly', 'roles:read'),
  ('u-readonly', 'permissions:read'),
  ('u-readonly', 'permission-groups:read'),
  ('u-users', 'users:read'),
  ('u-users', 'users:write'),
  ('u-users', 'roles:read'),
  ('u-roles', 'roles:read'),
  ('u-roles', 'roles:write');
