import { describe, expect, it } from 'vitest'
import { branchFromTurn, createCompactionBarrier, defaultDisclosure, discoverModels, exportSession, lifecycleRows, normalizeActivityMode, projectSessions, repairPartialTranscript, searchModels, shouldRenderActivity, undoToTurn } from './conversation-runtime'
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

  it('keeps the lifecycle matrix explicit', () => { expect(lifecycleRows.map((row) => row.kind)).toEqual(['normal', 'error', 'cancel', 'switch', 'reload', 'reconnect', 'compression', 'recovery']) })
})
