CREATE VIRTUAL TABLE chunk_vectors USING vec0 (
    embedding float[%d] distance_metric=cosine
);
