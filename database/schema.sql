CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    email         TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT ''

);

CREATE TABLE IF NOT EXISTS sessions (
    uuid       TEXT     PRIMARY KEY,
    user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE,
    slug TEXT    NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS topics (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER  NOT NULL REFERENCES users(id),
    category_id INTEGER  NOT NULL REFERENCES categories(id),
    title       TEXT     NOT NULL,
    description TEXT     NOT NULL DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id   INTEGER  NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    user_id    INTEGER  NOT NULL REFERENCES users(id),
    content    TEXT     NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    image_url TEXT DEFAULT '',
    parent_id INTEGER REFERENCES posts(id)
);

CREATE TABLE IF NOT EXISTS reactions (
    user_id INTEGER NOT NULL REFERENCES users(id),
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    type    TEXT    NOT NULL CHECK(type IN ('like','dislike')),
    UNIQUE(user_id, post_id)
);

-- Catégories par défaut
INSERT OR IGNORE INTO categories (name, slug) VALUES
    ('Général',         'general'),
    ('Technologie',     'technologie'),
    ('Programmation',   'programmation'),
    ('Jeux vidéo',      'jeux-video'),
    ('Science',         'science'),
    ('Culture',         'culture');

