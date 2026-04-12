type CacheMetricRowLike = {
  cache_creation_tokens?: number | null
  cache_read_tokens?: number | null
  upstream_endpoint?: string | null
  service_tier?: string | null
  openai_ws_mode?: boolean | null
  model?: string | null
  upstream_model?: string | null
}

const openAIModelPattern =
  /\b(gpt|codex|computer-use-preview|o1|o3|o4|o4-mini|o3-mini)\b/i

export function isLikelyOpenAICacheCreationMetricUnavailable(
  row: CacheMetricRowLike | null | undefined
): boolean {
  if (!row) return false
  if ((row.cache_creation_tokens ?? 0) > 0) return false
  if ((row.cache_read_tokens ?? 0) <= 0) return false

  if (row.openai_ws_mode) return true
  if ((row.upstream_endpoint ?? '').trim() === '/v1/responses') return true
  if ((row.service_tier ?? '').trim() !== '') return true

  const combinedModel = `${row.model ?? ''} ${row.upstream_model ?? ''}`.trim()
  return openAIModelPattern.test(combinedModel)
}
