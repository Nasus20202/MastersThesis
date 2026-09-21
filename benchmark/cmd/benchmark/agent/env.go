package agent

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
	envLlamaKVUnified       = "LLAMA_KV_UNIFIED_PER_SLOT"
	envLlamaGPULayers       = "LLAMA_GPU_LAYERS"
	envLlamaVulkanDevice    = "LLAMA_VULKAN_DEVICE"
	envLlamaFlashAttention  = "LLAMA_FLASH_ATTN"
	envLlamaCacheTypeK      = "LLAMA_CACHE_TYPE_K"
	envLlamaCacheTypeV      = "LLAMA_CACHE_TYPE_V"
	envLlamaModelsMax       = "LLAMA_MODELS_MAX"
	envLlamaReasoning       = "LLAMA_REASONING"
	envLlamaReasoningBudget = "LLAMA_REASONING_BUDGET"
	envLlamaHost            = "LLAMA_HOST"
)

var llamaRuntimeEnvNames = [...]string{
	envLlamaKVUnified,
	envLlamaGPULayers,
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
}
