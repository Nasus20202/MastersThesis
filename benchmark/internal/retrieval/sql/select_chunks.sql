SELECT chunks.id, documents.path, documents.title, chunks.headings, chunks.body
FROM chunks
JOIN documents ON documents.id = chunks.document_id
WHERE chunks.id IN (SELECT value FROM json_each(?));
