export interface HermesHealth {
  ok: boolean
  configured: boolean
  reachable: boolean
  base_url: string
  message?: string
}

export interface StartChatInput {
  session_id?: string
  message: string
  model?: string
  provider?: string
}

export interface StartChatResponse {
  stream_id: string
  session_id: string
}

async function readJson<T>(response: Response): Promise<T> {
  const data = (await response.json()) as T & { message?: string }
  if (!response.ok) throw new Error(data.message || `Request failed (${response.status})`)
  return data
}

export async function getHermesHealth(signal?: AbortSignal) {
  return readJson<HermesHealth>(await fetch('/api/health/hermes', { signal }))
}

export async function startChat(input: StartChatInput) {
  return readJson<StartChatResponse>(
    await fetch('/api/chat/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    }),
  )
}

export async function cancelChat(streamId: string) {
  return readJson<{ ok: boolean }>(
    await fetch(`/api/chat/cancel?stream_id=${encodeURIComponent(streamId)}`, { method: 'POST' }),
  )
}

