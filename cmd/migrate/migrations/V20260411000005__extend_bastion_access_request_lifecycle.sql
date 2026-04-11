ALTER TABLE bastion_access_requests ADD COLUMN approved_until TIMESTAMP NULL;
ALTER TABLE bastion_access_requests ADD COLUMN consumed_at TIMESTAMP NULL;
ALTER TABLE bastion_access_requests ADD COLUMN consumed_session_id BIGINT NULL;

CREATE INDEX idx_bastion_access_requests_status
  ON bastion_access_requests (status, approved_until, consumed_at);
