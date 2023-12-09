CREATE TABLE mangas (
    mangadex_id TEXT PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT,
    description TEXT,
    manga_path TEXT,
    cover_art_id INTEGER,
    FOREIGN KEY (cover_art_id) REFERENCES cover_arts (id)
);

CREATE TABLE cover_arts (
    mangadex_id TEXT PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    filename TEXT,
    manga_id TEXT,
    FOREIGN KEY (manga_id) REFERENCES mangas (id)
);

CREATE TABLE chapters (
    mangadex_id TEXT PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT,
    chapter_number TEXT,
    manga_id TEXT,
    is_read INTEGER DEFAULT 0,
    read_at DATETIME,
    FOREIGN KEY (manga_id) REFERENCES mangas (id)
);
