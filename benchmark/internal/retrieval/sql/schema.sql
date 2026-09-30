CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT NOT NULL);

CREATE TABLE documents (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    blob_sha TEXT NOT NULL,
    title TEXT NOT NULL
);

CREATE TABLE chunks (
    id INTEGER PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES documents (id),
    headings TEXT NOT NULL,
    body TEXT NOT NULL
);

CREATE VIRTUAL TABLE chunk_text USING fts5 (
    title, headings, body, tokenize = 'porter unicode61'
);
