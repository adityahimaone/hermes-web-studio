import type { SessionDetail, SessionSummary } from './chat-contract'
import type { GitStatus, WorkspaceEntry, WorkspacePreview } from './workspace-contract'

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
  attachment_ids?: string[]
}

export interface StartChatResponse {
  stream_id: string
  session_id: string
}

export interface UploadedAttachment { id: string; name: string; mime: string; size: number }

export async function uploadAttachment(file: File, sessionId?: string) {
  const body = new FormData()
  body.append('file', file)
  if (sessionId) body.append('session_id', sessionId)
  return readJson<UploadedAttachment>(await fetch('/api/attachments', { method: 'POST', body }))
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

export async function getSessions(signal?: AbortSignal) {
  return readJson<{ sessions: SessionSummary[] }>(await fetch('/api/sessions', { signal }))
}

export async function getSession(sessionId: string, signal?: AbortSignal) {
  return readJson<SessionDetail>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}`, { signal }))
}

export async function renameSession(sessionId: string, title: string) {
  return readJson<SessionDetail>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}/rename`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ title }) }))
}

export async function setSessionPinned(sessionId: string, pinned: boolean) {
  return readJson<SessionDetail>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}/pin`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ pinned }) }))
}

export async function setSessionArchived(sessionId: string, archived: boolean) {
  return readJson<SessionDetail>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}/archive`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ archived }) }))
}

export async function deleteSession(sessionId: string) {
  return readJson<{ ok: boolean; session_id: string }>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}`, { method: 'DELETE' }))
}

export async function truncateSession(sessionId: string, count: number) {
  return readJson<SessionDetail>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}/truncate`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ count }) }))
}

export async function updateSession(sessionId: string, patch: Record<string, unknown>) {
  return readJson<SessionDetail>(await fetch(`/api/sessions/${encodeURIComponent(sessionId)}`, {
    method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(patch),
  }))
}

export type ApprovalChoice = 'once' | 'session' | 'always' | 'deny'

export async function resolveApproval(runId: string, choice: ApprovalChoice) {
  return readJson<{ ok: boolean }>(await fetch(`/api/runs/${encodeURIComponent(runId)}/approval`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ choice }) }))
}

export function streamUrl(streamId: string, lastEventId?: string): string {
  const params = new URLSearchParams({ stream_id: streamId })
  if (lastEventId) params.set('after', lastEventId)
  return `/api/chat/stream?${params.toString()}`
}

export async function getWorkspaceTree(path = '.', signal?: AbortSignal) { return readJson<{ root: string; path: string; entries: WorkspaceEntry[] }>(await fetch(`/api/workspace/tree?path=${encodeURIComponent(path)}`, { signal })) }
export async function getWorkspacePreview(path: string, signal?: AbortSignal) { return readJson<WorkspacePreview>(await fetch(`/api/workspace/preview?path=${encodeURIComponent(path)}`, { signal })) }
export async function getWorkspaceGit(path = '.', signal?: AbortSignal) { return readJson<GitStatus>(await fetch(`/api/workspace/git?path=${encodeURIComponent(path)}`, { signal })) }
export function workspaceDownloadUrl(path: string) { return `/api/workspace/download?path=${encodeURIComponent(path)}` }
export async function saveWorkspaceFile(path: string, content: string) { return readJson<WorkspacePreview>(await fetch('/api/workspace/file', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path, content }) })) }
export async function createWorkspaceItem(path: string, type: 'file' | 'directory', content = '') { return readJson<{ ok: boolean }>(await fetch('/api/workspace/item', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path, type, content }) })) }
export async function renameWorkspaceItem(path: string, name: string) { return readJson<{ ok: boolean }>(await fetch('/api/workspace/rename', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path, name }) })) }
export async function deleteWorkspaceItem(path: string) { return readJson<{ ok: boolean }>(await fetch(`/api/workspace/item?path=${encodeURIComponent(path)}`, { method: 'DELETE' })) }
export async function uploadWorkspaceFile(path: string, file: File) { const body = new FormData(); body.append('path', path); body.append('file', file); return readJson<WorkspaceEntry>(await fetch('/api/workspace/upload', { method: 'POST', body })) }
