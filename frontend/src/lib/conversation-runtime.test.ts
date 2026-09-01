import { describe, expect, it, vi } from 'vitest'
import { appendCompletedAssistant, branchFromTurn, claimInflightTurn, createCompactionBarrier, defaultDisclosure, dedupeRestoredAssistant, discoverModels, exportSession, isCurrentConversation, isCurrentPump, lifecycleRows, normalizeActivityMode, normalizeClientError, normalizeRestoreError, pollSessionUntilSettled, projectSessions, releaseOwnedController, repairPartialTranscript, resetAnswerAtSessionBoundary, searchModels, shouldRenderActivity, shouldReplacePendingPump, queuedTurnBaseline, undoToTurn } from './conversation-runtime'
import type { ChatMessage } from './chat-contract'

const messages: ChatMessage[] = [{ id: 'u1', role: 'user', content: 'hello', status: 'complete' }, { id: 'a1', role: 'assistant', content: 'hi', status: 'complete' }]

describe('conversation runtime contracts', () => {
  it('normalizes disclosure and settles activity without hiding transparent work', () => {
    expect(defaultDisclosure).toEqual({ mode: 'compact', showSettledActivity: true })
    expect(normalizeActivityMode('bad')).toBe('compact')
    expect(shouldRenderActivity('final', false, true)).toBe(false)
    expect(shouldRenderActivity('transparent', true, false)).toBe(false)
  })

  it('creates barriers and repairs partial transcript entries deterministically', () => {
    expect(createCompactionBarrier(messages, 'summary', '2026-08-30T00:00:00Z')).toMatchObject({ beforeMessageCount: 2, source: 'local-recovery' })
    const repair = repairPartialTranscript([...messages, { id: 'bad', role: 'assistant', content: ' ', status: 'streaming' }], 'audit-1')
    expect(repair.messages).toEqual(messages)
    expect(repair.removed.map((item) => item.id)).toEqual(['bad'])
  })

  it('preserves explicit branch and undo lineage', () => {
    expect(branchFromTurn(messages, 'a1', { id: 'session-1', branch: 0, action: 'root' })?.prefix).toEqual(messages)
    expect(branchFromTurn(messages, 'a1', { id: 'session-1', branch: 0, action: 'root' })?.lineage).toMatchObject({ parentId: 'a1', branch: 1, action: 'fork' })
    expect(undoToTurn(messages, 'a1')?.map((item) => item.id)).toEqual(['u1'])
  })

  it('deduplicates projections and keeps external sessions read-only metadata', () => {
    const projected = projectSessions([{ session_id: 'one', title: 'old', source: 'cli', updated_at: '2026-08-29T00:00:00Z' }, { session_id: 'one', title: 'new', source: 'cli', updated_at: '2026-08-30T00:00:00Z' }])
    expect(projected).toHaveLength(1)
    expect(projected[0]).toMatchObject({ title: 'new', source: 'cli', external: true })
  })

  it('discovers searchable models without inventing unavailable data', () => {
    const models = discoverModels([{ id: 'gpt', provider: 'local', aliases: ['fast'], capabilities: ['tools'] }, { provider: 'missing' }])
    expect(searchModels(models, 'FAST')[0].id).toBe('gpt')
    expect(discoverModels({ models: [] })).toEqual([])
  })

  it('exports safe markdown, JSON, and escaped HTML', () => {
    expect(exportSession('a/b', messages, 'markdown').filename).toBe('a-b.md')
    expect(exportSession('x', messages, 'json').content).toContain('"messages"')
    expect(exportSession('<x>', [{ ...messages[0], content: '<script>' }], 'html').content).toContain('&lt;script&gt;')
  })

  it('polls restored session until expired stream turn has persisted assistant reply', async () => {
    const getSession = vi.fn().mockResolvedValueOnce({ messages: [{ role: 'user', content: 'hello' }] }).mockResolvedValueOnce({ messages: [...messages, { role: 'assistant', content: 'done' }] })
    await expect(pollSessionUntilSettled(getSession, 2, 2)).resolves.toMatchObject({ role: 'assistant', content: 'done' })
    expect(getSession).toHaveBeenCalledTimes(2)
  })

  it('propagates non-abort polling failures for lifecycle cleanup', async () => {
    const failure = new Error('session read failed')
    await expect(pollSessionUntilSettled(async () => { throw failure }, 0, 1)).rejects.toThrow('session read failed')
  })

  it('merges restored assistant into latest queued transcript by identity', () => {
    const queued = [...messages, { id: 'u2', role: 'user' as const, content: 'next', status: 'complete' as const }]
    expect(dedupeRestoredAssistant(queued, { id: 'a2', role: 'assistant', content: 'done', status: 'complete' })).toEqual([...queued, { id: 'a2', role: 'assistant', content: 'done', status: 'complete' }])
  })

  it('aborts polling and ignores pending completion', async () => {
    const controller = new AbortController()
    let resolve!: (value: { messages: ChatMessage[] }) => void
    const getSession = vi.fn(() => new Promise<{ messages: ChatMessage[] }>((r) => { resolve = r }))
    const pending = pollSessionUntilSettled(getSession, 0, 2, 0, controller.signal)
    controller.abort()
    resolve({ messages: [...messages, { id: 'same', role: 'assistant', content: 'done', status: 'complete' }] })
    await expect(pending).resolves.toBeNull()
  })

  it('times out expired stream polling without inventing transcript content', async () => {
    const getSession = vi.fn().mockResolvedValue({ messages })
    await expect(pollSessionUntilSettled(getSession, 2, 1)).resolves.toBeNull()
    expect(getSession).toHaveBeenCalledTimes(1)
  })

  it('claims hard-reload journal before stream assignment and rejects stale restore', () => {
    expect(claimInflightTurn('stream-1', 'session-1', 2, { streamId: null, sessionId: 'session-1', epoch: 2 })).toBe(true)
    expect(claimInflightTurn('stream-1', 'session-1', 2, { streamId: 'stream-2', sessionId: 'session-1', epoch: 2 })).toBe(false)
    expect(claimInflightTurn('stream-1', 'session-1', 1, { streamId: null, sessionId: 'session-1', epoch: 2 })).toBe(false)
  })

  it('resets answer accumulator only at session boundary', () => {
    const answer = { current: 'old streamed answer' }
    resetAnswerAtSessionBoundary(answer)
    expect(answer.current).toBe('')
  })

  it('aborts and releases only owned restore controller', () => {
    const owned = new AbortController()
    const replacement = new AbortController()
    expect(releaseOwnedController(owned, owned)).toBeNull()
    expect(owned.signal.aborted).toBe(true)
    expect(releaseOwnedController(owned, replacement)).toBe(replacement)
    expect(replacement.signal.aborted).toBe(false)
  })

  it('guards conversation callbacks by epoch, stream, and session identity', () => {
    expect(isCurrentConversation('stream-1', 'session-1', 2, { streamId: 'stream-1', sessionId: 'session-1', epoch: 2 })).toBe(true)
    expect(isCurrentConversation('stream-1', 'session-1', 1, { streamId: 'stream-1', sessionId: 'session-1', epoch: 2 })).toBe(false)
    expect(isCurrentConversation('stream-1', 'session-1', 2, { streamId: 'stream-1', sessionId: 'session-2', epoch: 2 })).toBe(false)
  })

  it('isolates cancelled old pump from queued replacement lifecycle', () => {
    const oldController = new AbortController()
    const newController = new AbortController()
    expect(isCurrentPump(oldController, newController, 'session-1', 3, { sessionId: 'session-1', epoch: 3 })).toBe(false)
    expect(isCurrentPump(newController, newController, 'session-1', 3, { sessionId: 'session-1', epoch: 3 })).toBe(true)
  })

  it('normalizes restore failures without exposing private details', () => {
    expect(normalizeRestoreError(new Error('/home/adityahimaone/.hermes/token secret=abc'))).toBe('Unable to restore Hermes session.')
  })

  it('normalizes client failures without exposing backend details', () => {
    expect(normalizeClientError(new Error('gateway token=abc /private/path'))).toBe('Unable to complete Hermes request.')
  })

  it('marks pending pump replacement safe when no stream ID exists', () => {
    const oldController = new AbortController()
    const currentController = new AbortController()
    expect(shouldReplacePendingPump(oldController, currentController, 'session-1', 2, { sessionId: 'session-1', epoch: 2 })).toBe(true)
    expect(shouldReplacePendingPump(oldController, null, 'session-1', 2, { sessionId: 'session-1', epoch: 2 })).toBe(false)
  })

  it('computes queued turn baseline from completed transcript', () => {
    expect(queuedTurnBaseline(messages)).toBe(3)
    expect(queuedTurnBaseline([...messages, { id: 'a2', role: 'assistant', content: 'next', status: 'complete' }])).toBe(4)
  })

  it('keeps identical assistant replies distinct across turns', async () => {
    const getSession = vi.fn().mockResolvedValue({ messages: [...messages, { id: 'a2', role: 'assistant', content: 'hi', status: 'complete' }] })
    await expect(pollSessionUntilSettled(getSession, 2, 1)).resolves.toMatchObject({ id: 'a2', content: 'hi' })
  })

  it('deduplicates replayed completion after restored assistant without merging turns', () => {
    const restored = [...messages, { id: 'u2', role: 'user' as const, content: 'again', status: 'complete' as const }, { id: 'a2', role: 'assistant' as const, content: 'done', status: 'complete' as const }]
    expect(appendCompletedAssistant(restored, 'done')).toEqual(restored)
    expect(appendCompletedAssistant([...messages, { id: 'u2', role: 'user' as const, content: 'again', status: 'complete' as const }], 'hi')).toHaveLength(4)
  })

  it('preserves restored assistant identity while deduplicating same turn', () => {
    const restored = { id: 'turn-1', role: 'assistant' as const, content: 'done', status: 'complete' as const }
    expect(dedupeRestoredAssistant([...messages, { ...restored, content: 'old' }], restored)).toEqual([...messages, restored])
    expect(dedupeRestoredAssistant(messages, restored)).toEqual([...messages, restored])
  })

  it('keeps the lifecycle matrix explicit', () => { expect(lifecycleRows.map((row) => row.kind)).toEqual(['normal', 'error', 'cancel', 'switch', 'reload', 'reconnect', 'compression', 'recovery']) })
})
