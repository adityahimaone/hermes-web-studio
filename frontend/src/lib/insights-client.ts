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

import { readJson } from './api-client'

export async function getInsights(signal?: AbortSignal): Promise<InsightsResponse> {
  const response = await fetch('/api/operator/insights', { signal })
  return readJson<InsightsResponse>(response)
}
