DROP INDEX idx_bastion_access_requests_status;

ALTER TABLE bastion_access_requests DROP COLUMN consumed_session_id;
ALTER TABLE bastion_access_requests DROP COLUMN consumed_at;
ALTER TABLE bastion_access_requests DROP COLUMN approved_until;
