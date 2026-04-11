CREATE TABLE bastion_hosts (
  id                   BIGINT PRIMARY KEY,
  name                 VARCHAR(128) NOT NULL UNIQUE,
  address              VARCHAR(255) NOT NULL,
  port                 INTEGER NOT NULL DEFAULT 22,
  protocol             VARCHAR(32) NOT NULL DEFAULT 'ssh',
  environment          VARCHAR(64),
  platform             VARCHAR(32),
  authentication_type  VARCHAR(32) NOT NULL DEFAULT 'managed',
  credential_ref       VARCHAR(128),
  jump_enabled         BOOLEAN NOT NULL DEFAULT TRUE,
  recording_policy     VARCHAR(32) NOT NULL DEFAULT 'required',
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bastion_host_accounts (
  id                   BIGINT PRIMARY KEY,
  host_id              BIGINT NOT NULL,
  account_name         VARCHAR(128) NOT NULL,
  authentication_type  VARCHAR(32) NOT NULL DEFAULT 'managed',
  credential_ref       VARCHAR(128),
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (host_id, account_name),
  FOREIGN KEY (host_id) REFERENCES bastion_hosts(id) ON DELETE CASCADE
);

CREATE TABLE bastion_access_policies (
  id                   BIGINT PRIMARY KEY,
  name                 VARCHAR(128) NOT NULL UNIQUE,
  subject_type         VARCHAR(32) NOT NULL,
  subject_ref          VARCHAR(128) NOT NULL,
  target_type          VARCHAR(32) NOT NULL,
  target_ref           VARCHAR(128) NOT NULL,
  account_pattern      VARCHAR(128) NOT NULL DEFAULT '*',
  protocol             VARCHAR(32) NOT NULL DEFAULT 'ssh',
  approval_required    BOOLEAN NOT NULL DEFAULT FALSE,
  recording_required   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bastion_sessions (
  id                   BIGINT PRIMARY KEY,
  host_id              BIGINT NOT NULL,
  host_account_id      BIGINT,
  principal_id         VARCHAR(64) NOT NULL,
  protocol             VARCHAR(32) NOT NULL DEFAULT 'ssh',
  status               VARCHAR(32) NOT NULL,
  source_addr          VARCHAR(255),
  started_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ended_at             TIMESTAMP NULL,
  FOREIGN KEY (host_id) REFERENCES bastion_hosts(id) ON DELETE RESTRICT,
  FOREIGN KEY (host_account_id) REFERENCES bastion_host_accounts(id) ON DELETE SET NULL
);

CREATE TABLE bastion_session_events (
  id                   BIGINT PRIMARY KEY,
  session_id           BIGINT NOT NULL,
  event_type           VARCHAR(64) NOT NULL,
  payload              TEXT,
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES bastion_sessions(id) ON DELETE CASCADE
);

CREATE TABLE bastion_session_commands (
  id                   BIGINT PRIMARY KEY,
  session_id           BIGINT NOT NULL,
  command_text         TEXT NOT NULL,
  risk_level           VARCHAR(32) NOT NULL DEFAULT 'normal',
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES bastion_sessions(id) ON DELETE CASCADE
);

CREATE INDEX idx_bastion_hosts_lookup
  ON bastion_hosts (protocol, environment, platform);

CREATE INDEX idx_bastion_access_policies_lookup
  ON bastion_access_policies (subject_type, subject_ref, target_type, target_ref);

CREATE INDEX idx_bastion_sessions_lookup
  ON bastion_sessions (principal_id, status, started_at);
