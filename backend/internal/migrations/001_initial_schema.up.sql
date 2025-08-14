/* Initialization Status */

CREATE TABLE initialization_status (
    id INT PRIMARY KEY,
    user_created BOOLEAN NOT NULL DEFAULT false,
    pihole_status TEXT NOT NULL CHECK (pihole_status IN ('UNINITIALIZED', 'ADDED', 'SKIPPED')) DEFAULT 'UNINITIALIZED'
);
INSERT INTO initialization_status (id, user_created, pihole_status)
VALUES (1, 0, 'UNINITIALIZED');

/* Piholes */

CREATE TABLE piholes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scheme TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE, 
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
    OR OLD.name != NEW.name
    OR OLD.description IS NOT NEW.description
    OR OLD.password_enc != NEW.password_enc
BEGIN
    UPDATE piholes
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;

/* Sessions */

CREATE TABLE sessions (
	id TEXT PRIMARY KEY NOT NULL,
	user_id INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME NOT NULL
);

/* Users */

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_users_updated_at
AFTER UPDATE OF username, password_hash
ON users
FOR EACH ROW
WHEN OLD.username != NEW.username
    OR OLD.password_hash != NEW.password_hash
BEGIN
    UPDATE users
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;
