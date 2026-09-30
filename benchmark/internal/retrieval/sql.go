package retrieval

import _ "embed"

var (
	//go:embed sql/schema.sql
	schemaSQL string
	//go:embed sql/create_vectors.sql
	createVectorsSQL string
	//go:embed sql/insert_document.sql
	insertDocumentSQL string
	//go:embed sql/insert_chunk.sql
	insertChunkSQL string
	//go:embed sql/insert_chunk_text.sql
	insertChunkTextSQL string
	//go:embed sql/insert_vector.sql
	insertVectorSQL string
	//go:embed sql/insert_metadata.sql
	insertMetadataSQL string
	//go:embed sql/select_metadata.sql
	selectMetadataSQL string
	//go:embed sql/search_lexical.sql
	searchLexicalSQL string
	//go:embed sql/search_semantic.sql
	searchSemanticSQL string
	//go:embed sql/select_chunks.sql
	selectChunksSQL string
)
