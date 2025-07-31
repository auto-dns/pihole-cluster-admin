CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

CREATE TABLE piholes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scheme TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    description TEXT,
    password_enc TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (host, port)
);

CREATE TRIGGER update_piholes_updated_at
AFTER UPDATE OF scheme, host, port, description, password_enc
ON piholes
FOR EACH ROW
WHEN OLD.scheme != NEW.scheme
    OR OLD.host != NEW.host
    OR OLD.port != NEW.port
    OR OLD.description IS NOT NEW.description
    OR OLD.password_enc != NEW.password_enc
BEGIN
    UPDATE piholes
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;
