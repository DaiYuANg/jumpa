CREATE TABLE gateway_registry_nodes (
  id BIGINT PRIMARY KEY,
  node_key VARCHAR(191) NOT NULL,
  node_name VARCHAR(128) NOT NULL,
  runtime_type VARCHAR(32) NOT NULL,
  advertise_addr VARCHAR(255) NOT NULL,
  ssh_listen_addr VARCHAR(255) NOT NULL,
  zone_name VARCHAR(64) NOT NULL,
  tags_csv TEXT NULL,
  state VARCHAR(32) NOT NULL,
  registered_at TIMESTAMP NOT NULL,
  last_seen_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX uk_gateway_registry_nodes_node_key
  ON gateway_registry_nodes (node_key);

CREATE INDEX idx_gateway_registry_nodes_state_last_seen
  ON gateway_registry_nodes (state, last_seen_at);
