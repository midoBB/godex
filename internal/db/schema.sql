CREATE TABLE mangas (
    id INTEGER PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    mangadex_id TEXT UNIQUE,
    title TEXT,
    description TEXT,
    manga_path TEXT,
    cover_art_id INTEGER,
    FOREIGN KEY (cover_art_id) REFERENCES cover_arts (id)
);

CREATE TABLE cover_arts (
    id INTEGER PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    mangadex_id TEXT UNIQUE,
    filename TEXT,
    manga_id INTEGER,
    FOREIGN KEY (manga_id) REFERENCES mangas (id)
);

CREATE TABLE chapters (
    id INTEGER PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    mangadex_id TEXT UNIQUE,
    title TEXT,
    chapter TEXT,
    manga_id INTEGER,
    FOREIGN KEY (manga_id) REFERENCES mangas (id)
);
