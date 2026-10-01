SELECT rowid, distance
FROM chunk_vectors
WHERE embedding MATCH ? AND k = ?
ORDER BY distance;
