CREATE TABLE IF NOT EXISTS users (
    user_id       INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    DATETIME
);

CREATE TABLE IF NOT EXISTS demands (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    vk_id         INTEGER NOT NULL,
    for_week      DATETIME NOT NULL,
    first_name    TEXT,
    last_name     TEXT,
    players_count INTEGER NOT NULL,
    slots         BLOB,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    vk_username   TEXT,
    for_week_str  TEXT NOT NULL,
    comment       TEXT,
    UNIQUE (vk_id, for_week_str)
);
CREATE INDEX IF NOT EXISTS idx_demands_for_week_str ON demands (vk_id, for_week_str);

CREATE TABLE IF NOT EXISTS slots (
    id           INTEGER NOT NULL,
    name         TEXT NOT NULL,
    valid_from   DATETIME NOT NULL,
    valid_until  DATETIME NOT NULL,
    for_week_str TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_slots_id_for_week_str ON slots (id, for_week_str);

CREATE TABLE IF NOT EXISTS slots_default (
    id           INTEGER PRIMARY KEY,
    name         TEXT NOT NULL,
    valid_from   DATETIME NOT NULL,
    valid_until  DATETIME NOT NULL,
    for_week_str TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_slots_default_id_for_week_str ON slots_default (id, for_week_str);

CREATE TABLE IF NOT EXISTS masters (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    vk_id      INTEGER NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
