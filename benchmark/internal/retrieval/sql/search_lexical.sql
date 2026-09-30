SELECT rowid, bm25(chunk_text) AS rank
FROM chunk_text
WHERE chunk_text MATCH ?
ORDER BY rank, rowid
LIMIT ?;
