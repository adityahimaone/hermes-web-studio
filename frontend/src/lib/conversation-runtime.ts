import { normalizeSessionMessages, type ChatMessage, type SessionSummary } from './chat-contract'

export type ActivityMode = 'compact' | 'transparent' | 'final'
export type LifecycleKind = 'normal' | 'error' | 'cancel' | 'switch' | 'reload' | 'reconnect' | 'compression' | 'recovery'

export interface DisclosurePreference { mode: ActivityMode; showSettledActivity: boolean }
export const defaultDisclosure: DisclosurePreference = { mode: 'compact', showSettledActivity: true }

export function normalizeActivityMode(value: unknown): ActivityMode {
  return value === 'transparent' || value === 'final' ? value : 'compact'
}

export function disclosureFor(mode: unknown, showSettledActivity = true): DisclosurePreference {
  return { mode: normalizeActivityMode(mode), showSettledActivity: Boolean(showSettledActivity) }
}

export function shouldRenderActivity(mode: ActivityMode, settled: boolean, showSettledActivity: boolean): boolean {
  return mode !== 'final' && (!settled || showSettledActivity)
}

export interface CompactionBarrier { beforeMessageCount: number; summary: string; createdAt: string; source: 'gateway' | 'local-recovery' }
export interface TranscriptRepair { messages: ChatMessage[]; removed: ChatMessage[]; reason: string; auditId: string }

export function createCompactionBarrier(messages: ChatMessage[], summary: string, createdAt: string, source: CompactionBarrier['source'] = 'local-recovery'): CompactionBarrier | null {
  const clean = summary.trim()
  return clean && messages.length ? { beforeMessageCount: messages.length, summary: clean, createdAt, source } : null
}

export function repairPartialTranscript(messages: ChatMessage[], auditId: string): TranscriptRepair {
  const valid = messages.filter((message) => message.content.trim() && (message.role === 'user' || message.role === 'assistant'))
  const removed = messages.filter((message) => !valid.includes(message))
  return { messages: valid.map((message) => ({ ...message, status: message.status === 'streaming' ? 'cancelled' : message.status })), removed, reason: removed.length ? 'Removed empty or unfinished transcript entries.' : 'Transcript is already valid.', auditId }
}

export function isCurrentConversation(streamId: string, sessionId: string, epoch: number, current: { streamId: string | null; sessionId: string; epoch: number }): boolean {
  return current.streamId === streamId && current.sessionId === sessionId && current.epoch === epoch
}

export function claimInflightTurn(streamId: string, sessionId: string, epoch: number, current: { streamId: string | null; sessionId: string; epoch: number }): boolean {
  return current.sessionId === sessionId && current.epoch === epoch && (current.streamId === null || current.streamId === streamId)
}

export function isCurrentPump(controller: AbortController, currentController: AbortController | null, sessionId: string, epoch: number, current: { sessionId: string; epoch: number }): boolean {
  return controller === currentController && current.sessionId === sessionId && current.epoch === epoch
}

export function canMutatePumpState(controller: AbortController, currentController: AbortController | null, sessionId: string, epoch: number, current: { sessionId: string; epoch: number }): boolean {
  return !controller.signal.aborted && isCurrentPump(controller, currentController, sessionId, epoch, current)
}

export function resetAnswerAtSessionBoundary(answer: { current: string }): void {
  answer.current = ''
}

export function resetCursorAtSessionBoundary(cursor: { current: number }): void {
  cursor.current = 0
}

export function resetConversationRuntimeState(state: {
  cursor: { current: number }
  pump: { current: AbortController | null }
  pendingUser: { current: string | null }
  answer: { current: string }
}): void {
  resetCursorAtSessionBoundary(state.cursor)
  resetOwnedPumpState(state.pump, state.pendingUser)
  resetAnswerAtSessionBoundary(state.answer)
}

export function releaseOwnedController(controller: AbortController, currentController: AbortController | null): AbortController | null {
  if (currentController !== controller) return currentController
  controller.abort()
  return null
}

export function resetOwnedPumpState(pump: { current: AbortController | null }, pendingUser: { current: string | null }): void {
  pump.current?.abort()
  pump.current = null
  pendingUser.current = null
}

export function normalizeRestoreError(_error: unknown): string {
  return 'Unable to restore Hermes session.'
}

export function normalizeClientError(_error: unknown): string {
  return 'Unable to complete Hermes request.'
}

export function shouldReplacePendingPump(controller: AbortController, currentController: AbortController | null, sessionId: string, epoch: number, current: { sessionId: string; epoch: number }): boolean {
  return currentController !== null && controller !== currentController && current.sessionId === sessionId && current.epoch === epoch
}

export function queuedTurnBaseline(messages: ChatMessage[]): number {
  return messages.length + 1
}

export function dedupeRestoredAssistant(messages: ChatMessage[], assistant: ChatMessage): ChatMessage[] {
  const index = messages.findIndex((message) => message.id === assistant.id)
  return index < 0 ? [...messages, assistant] : messages.map((message, itemIndex) => itemIndex === index ? assistant : message)
}

export function appendCompletedAssistant(messages: ChatMessage[], content: string, status: ChatMessage['status'] = 'complete'): ChatMessage[] {
  if (!content.trim()) return messages
  const last = messages.at(-1)
  if (last?.role === 'assistant' && last.content === content) return messages
  return [...messages, { id: crypto.randomUUID(), role: 'assistant', content, status }]
}

export async function pollSessionUntilSettled<T extends { messages?: unknown[] }>(getSession: (signal?: AbortSignal) => Promise<T>, baselineMessageCount: number, maxPolls: number, intervalMs = 0, signal?: AbortSignal): Promise<ChatMessage | null> {
  for (let poll = 0; poll < maxPolls; poll += 1) {
    if (signal?.aborted) return null
    if (poll > 0 && intervalMs > 0) {
      await new Promise<void>((resolve) => {
        const timer = setTimeout(resolve, intervalMs)
        signal?.addEventListener('abort', () => { clearTimeout(timer); resolve() }, { once: true })
      })
      if (signal?.aborted) return null
    }
    const detail = await getSession(signal)
    if (signal?.aborted) return null
    const messages = normalizeSessionMessages(detail.messages)
    const assistant = messages.slice(baselineMessageCount).filter((message) => message.role === 'assistant' && message.content.trim()).at(-1)
    if (assistant) return assistant
  }
  return null
}

export interface Lineage { id: string; parentId?: string; branch: number; action: 'root' | 'fork' | 'undo' | 'duplicate' }

export function branchFromTurn(messages: ChatMessage[], messageId: string, lineage: Lineage): { prefix: ChatMessage[]; lineage: Lineage } | null {
  const index = messages.findIndex((message) => message.id === messageId)
  if (index < 0) return null
  return { prefix: messages.slice(0, index + 1), lineage: { ...lineage, parentId: messageId, branch: lineage.branch + 1, action: 'fork' } }
}

export function undoToTurn(messages: ChatMessage[], messageId: string): ChatMessage[] | null {
  const index = messages.findIndex((message) => message.id === messageId)
  return index < 0 ? null : messages.slice(0, index)
}

export interface SessionProjection extends SessionSummary { source: 'webui' | 'cli' | 'cron' | 'webhook' | 'gateway' | 'unknown'; external: boolean; clusterKey: string }

function stringField(item: SessionSummary, ...keys: string[]): string { for (const key of keys) { const value = item[key]; if (typeof value === 'string' && value.trim()) return value.trim() } return '' }
function sourceOf(item: SessionSummary): SessionProjection['source'] { const source = stringField(item, 'source', 'origin', 'session_source', 'session_type').toLowerCase(); return source.includes('cli') ? 'cli' : source.includes('cron') ? 'cron' : source.includes('webhook') ? 'webhook' : source.includes('gateway') ? 'gateway' : source.includes('webui') ? 'webui' : 'unknown' }

export function projectSessions(items: SessionSummary[]): SessionProjection[] {
  const byId = new Map<string, SessionProjection>()
  for (const item of items) {
    if (!item.session_id) continue
    const source = sourceOf(item)
    const projected = { ...item, source, external: source !== 'webui' && source !== 'unknown', clusterKey: stringField(item, 'conversation_id', 'thread_id', 'session_id') }
    const current = byId.get(item.session_id)
    if (!current || new Date(projected.updated_at || projected.created_at || 0).getTime() > new Date(current.updated_at || current.created_at || 0).getTime()) byId.set(item.session_id, projected)
  }
  return [...byId.values()].sort((a, b) => Number(Boolean(b.pinned)) - Number(Boolean(a.pinned)) || new Date(b.updated_at || b.created_at || 0).getTime() - new Date(a.updated_at || a.created_at || 0).getTime())
}

export interface ModelOption { id: string; label: string; provider: string; aliases: string[]; capabilities: string[]; available: boolean }
export function discoverModels(value: unknown): ModelOption[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((entry) => {
    if (!entry || typeof entry !== 'object') return []
    const item = entry as Record<string, unknown>; const id = typeof item.id === 'string' ? item.id.trim() : ''
    if (!id) return []
    const aliases = Array.isArray(item.aliases) ? item.aliases.filter((alias): alias is string => typeof alias === 'string') : []
    const capabilities = Array.isArray(item.capabilities) ? item.capabilities.filter((capability): capability is string => typeof capability === 'string') : []
    return [{ id, label: typeof item.name === 'string' ? item.name : id, provider: typeof item.provider === 'string' ? item.provider : 'unknown', aliases, capabilities, available: item.available !== false }]
  })
}

export function searchModels(models: ModelOption[], query: string): ModelOption[] { const needle = query.trim().toLowerCase(); return needle ? models.filter((model) => [model.id, model.label, model.provider, ...model.aliases].join(' ').toLowerCase().includes(needle)) : models }

export interface ExportDocument { format: 'markdown' | 'json' | 'html'; filename: string; content: string; safe: true }
export function exportSession(title: string, messages: ChatMessage[], format: ExportDocument['format']): ExportDocument {
  const safeTitle = title.replace(/[\\/:*?"<>|]/g, '-').trim() || 'hermes-session'
  const json = JSON.stringify({ title, messages }, null, 2)
  if (format === 'json') return { format, filename: `${safeTitle}.json`, content: json, safe: true }
  const markdown = `# ${title || 'Hermes session'}\n\n${messages.map((message) => `## ${message.role === 'assistant' ? 'Hermes' : 'User'}\n\n${message.content}`).join('\n\n')}`
  if (format === 'markdown') return { format, filename: `${safeTitle}.md`, content: markdown, safe: true }
  return { format, filename: `${safeTitle}.html`, content: `<!doctype html><meta charset="utf-8"><title>${escapeHTML(title)}</title><main>${messages.map((message) => `<article><h2>${message.role === 'assistant' ? 'Hermes' : 'User'}</h2><p>${escapeHTML(message.content)}</p></article>`).join('')}</main>`, safe: true }
}

function escapeHTML(value: string): string { return value.replace(/[&<>"']/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[character] || character)) }

export interface LifecycleRow { kind: LifecycleKind; expected: 'stream' | 'settled' | 'safe-error' | 'restored' | 'capability-unavailable' }
export const lifecycleRows: LifecycleRow[] = [
  { kind: 'normal', expected: 'settled' }, { kind: 'error', expected: 'safe-error' }, { kind: 'cancel', expected: 'settled' }, { kind: 'switch', expected: 'restored' },
  { kind: 'reload', expected: 'restored' }, { kind: 'reconnect', expected: 'stream' }, { kind: 'compression', expected: 'stream' }, { kind: 'recovery', expected: 'restored' },
]

export const unavailableUpstreamCapabilities = ['public-share', 'external-import', 'adaptive-title', 'live-model-quota'] as const
