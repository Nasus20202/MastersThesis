package llamaenv

const (
	envLlamaModelName       = "LLAMA_MODEL_NAME"
	envLlamaClientHost      = "LLAMA_CLIENT_HOST"
	envLlamaPort            = "LLAMA_PORT"
	envLlamaParallel        = "LLAMA_PARALLEL"
	envLlamaModelRepository = "LLAMA_MODEL_REPOSITORY"
	envLlamaModelRevision   = "LLAMA_MODEL_REVISION"
	envLlamaModelFile       = "LLAMA_MODEL_FILE"
	envLlamaModelQuant      = "LLAMA_MODEL_QUANTIZATION"
	envLlamaModelSHA256     = "LLAMA_MODEL_SHA256"
	envLlamaDraftRepository = "LLAMA_MODEL_DRAFT_REPOSITORY"
	envLlamaDraftRevision   = "LLAMA_MODEL_DRAFT_REVISION"
	envLlamaDraftFile       = "LLAMA_MODEL_DRAFT_FILE"
	envLlamaDraftSHA256     = "LLAMA_MODEL_DRAFT_SHA256"
	envLlamaSpecType        = "LLAMA_SPEC_TYPE"
	envLlamaSpecDraftNMax   = "LLAMA_SPEC_DRAFT_N_MAX"
	envLlamaKVUnified       = "LLAMA_KV_UNIFIED_PER_SLOT"
	envLlamaGPULayers       = "LLAMA_GPU_LAYERS"
	envLlamaDevice          = "LLAMA_DEVICE"
	envLlamaVulkanDevice    = "LLAMA_VULKAN_DEVICE"
	envLlamaFlashAttention  = "LLAMA_FLASH_ATTN"
	envLlamaCacheTypeK      = "LLAMA_CACHE_TYPE_K"
	envLlamaCacheTypeV      = "LLAMA_CACHE_TYPE_V"
	envLlamaModelsMax       = "LLAMA_MODELS_MAX"
	envLlamaReasoning       = "LLAMA_REASONING"
	envLlamaReasoningBudget = "LLAMA_REASONING_BUDGET"
	envLlamaHost            = "LLAMA_HOST"
	envLlamaBaseURL         = "LLAMA_BASE_URL"

	envInferenceProvider = "INFERENCE_PROVIDER"
	envOpenAIBaseURL     = "OPENAI_BASE_URL"
	envOpenAIAPIKey      = "OPENAI_API_KEY" //nolint:gosec // variable name, not a credential
	envOpenAIModel       = "OPENAI_MODEL"

	envEmbeddingModelName  = "EMBEDDING_MODEL_NAME"
	envEmbeddingPort       = "EMBEDDING_PORT"
	envEmbeddingBaseURL    = "EMBEDDING_BASE_URL"
	envEmbeddingRepository = "EMBEDDING_MODEL_REPOSITORY"
	envEmbeddingRevision   = "EMBEDDING_MODEL_REVISION"
	envEmbeddingFile       = "EMBEDDING_MODEL_FILE"
	envEmbeddingQuant      = "EMBEDDING_MODEL_QUANTIZATION"
	envEmbeddingSHA256     = "EMBEDDING_MODEL_SHA256"
)

var llamaRuntimeEnvNames = [...]string{
	envLlamaKVUnified,
	envLlamaGPULayers,
	envLlamaDevice,
	envLlamaVulkanDevice,
	envLlamaParallel,
	envLlamaFlashAttention,
	envLlamaCacheTypeK,
	envLlamaCacheTypeV,
	envLlamaModelsMax,
	envLlamaReasoning,
	envLlamaReasoningBudget,
	envLlamaHost,
	envLlamaPort,
	envLlamaClientHost,
	envLlamaSpecType,
	envLlamaSpecDraftNMax,
	envLlamaDraftRepository,
	envLlamaDraftRevision,
	envLlamaDraftFile,
	envLlamaDraftSHA256,
}
