export interface InsightsUsage {
  input_tokens: number
  output_tokens: number
  total_tokens: number
  available: boolean
}

export interface ProviderInsight {
  provider: string
  model?: string
  sessions: number
  messages: number
  usage: InsightsUsage
}

export interface InsightsResponse {
  generated_at: string
  source: string
  synchronization: { status: string; sessions_scanned: number; sessions_read: number; last_activity_at?: string | null }
  summary: { sessions: number; messages: number; user_messages: number; assistant_messages: number }
  usage: InsightsUsage
  provider_history: ProviderInsight[]
  cost: { available: boolean; reason?: string }
}

export async function getInsights(signal?: AbortSignal): Promise<InsightsResponse> {
  const response = await fetch('/api/operator/insights', { signal })
  const data = await response.json() as Partial<InsightsResponse> & { message?: string }
  if (!response.ok) throw new Error(data.message || 'Insights are unavailable.')
  return data as InsightsResponse
}
