CREATE TABLE bastion_access_requests (
  id                   BIGINT PRIMARY KEY,
  policy_id            BIGINT NOT NULL,
  principal_name       VARCHAR(128) NOT NULL,
  principal_email      VARCHAR(255),
  host_name            VARCHAR(128) NOT NULL,
  host_account         VARCHAR(128) NOT NULL,
  protocol             VARCHAR(32) NOT NULL DEFAULT 'ssh',
  status               VARCHAR(32) NOT NULL,
  requested_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  reviewed_at          TIMESTAMP NULL,
  reviewed_by          VARCHAR(128),
  review_comment       TEXT,
  FOREIGN KEY (policy_id) REFERENCES bastion_access_policies(id) ON DELETE CASCADE
);

CREATE INDEX idx_bastion_access_requests_lookup
  ON bastion_access_requests (policy_id, principal_name, host_name, host_account, protocol, requested_at);
