-- +goose Up
-- Adiciona suporte a multi-tenancy

CREATE TABLE tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER NOT NULL DEFAULT 1
);

ALTER TABLE connection_data ADD COLUMN tenant_id TEXT NOT NULL DEFAULT '';
ALTER TABLE connection_data ADD COLUMN name TEXT;
ALTER TABLE connection_data ADD COLUMN description TEXT;
ALTER TABLE connection_data ADD COLUMN max_connections INTEGER DEFAULT 10;
ALTER TABLE connection_data ADD COLUMN is_active INTEGER NOT NULL DEFAULT 1;

DROP INDEX IF EXISTS idx_connection_data_dsn;
CREATE INDEX idx_connection_data_tenant ON connection_data(tenant_id, is_active);
CREATE UNIQUE INDEX idx_connection_data_tenant_name ON connection_data(tenant_id, name);

CREATE TABLE connection_data_new (
     id TEXT PRIMARY KEY,
     tenant_id TEXT NOT NULL,
     name TEXT NOT NULL,
     description TEXT,
     dsn TEXT NOT NULL,
     dialect TEXT NOT NULL,
     dek TEXT NOT NULL,
     readonly INTEGER NOT NULL DEFAULT 0,
     max_connections INTEGER DEFAULT 10,
     is_active INTEGER NOT NULL DEFAULT 1,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

INSERT INTO tenants (id, name) VALUES ('default', 'Default Tenant');
INSERT INTO connection_data_new (id, tenant_id, name, dsn, dialect, dek, readonly, created_at, updated_at)
SELECT
    id,
    'default' as tenant_id,
    'legacy_' || id as name,
    dsn,
    dialect,
    dek,
    readonly,
    created_at,
    updated_at
FROM connection_data;

DROP TABLE connection_data;
ALTER TABLE connection_data_new RENAME TO connection_data;

CREATE INDEX idx_connection_data_tenant ON connection_data(tenant_id, is_active);
CREATE UNIQUE INDEX idx_connection_data_tenant_name ON connection_data(tenant_id, name);

CREATE TABLE connection_audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id TEXT NOT NULL,
  connection_id TEXT NOT NULL,
  action TEXT NOT NULL, -- 'query', 'schema', 'connect'
  query_hash TEXT,
  success INTEGER NOT NULL,
  error_message TEXT,
  execution_time_ms INTEGER,
  rows_affected INTEGER,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  FOREIGN KEY (connection_id) REFERENCES connection_data(id)
);

CREATE INDEX idx_audit_tenant_time ON connection_audit_log(tenant_id, timestamp);
CREATE INDEX idx_audit_connection_time ON connection_audit_log(connection_id, timestamp);

CREATE TABLE jwt_tokens (
                            jti TEXT PRIMARY KEY,
                            tenant_id TEXT NOT NULL,
                            connection_ids TEXT NOT NULL,
                            issued_at DATETIME NOT NULL,
                            expires_at DATETIME NOT NULL,
                            revoked INTEGER NOT NULL DEFAULT 0,
                            revoked_at DATETIME,
                            FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE INDEX idx_jwt_tenant_expiry ON jwt_tokens(tenant_id, expires_at, revoked);

-- +goose Down
DROP TABLE IF EXISTS jwt_tokens;
DROP TABLE IF EXISTS connection_audit_log;
DROP INDEX IF EXISTS idx_connection_data_tenant;
DROP INDEX IF EXISTS idx_connection_data_tenant_name;

CREATE TABLE connection_data_old (
     id TEXT PRIMARY KEY,
     dsn TEXT NOT NULL,
     dialect TEXT NOT NULL,
     readonly INTEGER NOT NULL DEFAULT 0,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
     dek TEXT NOT NULL
);

INSERT INTO connection_data_old (id, dsn, dialect, readonly, created_at, updated_at, dek)
SELECT id, dsn, dialect, readonly, created_at, updated_at, dek
FROM connection_data;

DROP TABLE connection_data;
ALTER TABLE connection_data_old RENAME TO connection_data;
CREATE INDEX idx_connection_data_dsn ON connection_data(dsn);

DROP TABLE IF EXISTS tenants;