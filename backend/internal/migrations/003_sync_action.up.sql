-- Expand the action CHECK constraint to include sync_from_node.
-- SQLite does not support ALTER COLUMN, so we recreate the table.
CREATE TABLE audit_log_new (
    id              INTEGER  PRIMARY KEY AUTOINCREMENT,
    actor           TEXT     NOT NULL,
    action          TEXT     NOT NULL CHECK (action IN (
                        'add_domain_rule',
                        'remove_domain_rule',
                        'set_cluster_blocking',
                        'set_node_blocking',
                        'sync_from_node'
                    )),
    target_domain   TEXT,
    rule_type       TEXT,
    rule_kind       TEXT,
    blocking_enabled INTEGER,
    blocking_timer  INTEGER,
    target_node_id  INTEGER,
    target_node_name TEXT,
    node_results    TEXT     NOT NULL DEFAULT '[]',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO audit_log_new SELECT * FROM audit_log;
DROP TABLE audit_log;
ALTER TABLE audit_log_new RENAME TO audit_log;
CREATE INDEX IF NOT EXISTS audit_log_created_at_idx ON audit_log(created_at);
