import { describe, expect, it } from 'vitest'
import { filterSessions, groupSessionsByDate, initialChatState, normalizeSessionMessages, parseInflightTurn, reduceChatEvent } from './chat-contract'

describe('chat event reducer', () => {
  it('builds an answer from token frames and settles on done', () => {
    let state = reduceChatEvent(initialChatState, { type: 'token', data: { text: 'Hello ' } })
    state = reduceChatEvent(state, { type: 'token', data: { text: 'Adit' } })
    state = reduceChatEvent(state, { type: 'done', data: {} })
    expect(state.answer).toBe('Hello Adit')
    expect(state.status).toBe('complete')
  })

  it('tracks a Hermes tool from start through completion', () => {
    let state = reduceChatEvent(initialChatState, { type: 'tool', data: { tid: 't1', name: 'terminal' } })
    state = reduceChatEvent(state, { type: 'tool_complete', data: { tid: 't1', name: 'terminal' } })
    expect(state.tools).toEqual([{ id: 't1', name: 'terminal', status: 'complete' }])
  })

  it('upserts progress frames and settles a subagent by its stable id', () => {
    let state = reduceChatEvent(initialChatState, { type: 'tool', data: { tid: 't1', name: 'terminal' } })
    state = reduceChatEvent(state, { type: 'tool', data: { tid: 't1', name: 'terminal', result: 'working' } })
    state = reduceChatEvent(state, { type: 'subagent', data: { id: 's1', name: 'research', status: 'running' } })
    state = reduceChatEvent(state, { type: 'subagent', data: { id: 's1', name: 'research', status: 'complete' } })
    expect(state.tools).toHaveLength(1)
    expect(state.tools[0].result).toBe('working')
    expect(state.subagents).toEqual([{ id: 's1', name: 'research', status: 'complete', task: undefined }])
  })

  it('surfaces safe application errors', () => {
    const state = reduceChatEvent(initialChatState, { type: 'apperror', data: { message: 'Gateway unavailable' } })
    expect(state.status).toBe('error')
    expect(state.error).toBe('Gateway unavailable')
  })

  it('normalizes legacy history and ignores unsupported message shapes', () => {
    expect(normalizeSessionMessages([
      { role: 'user', content: 'hello' },
      { role: 'assistant', content: 'world', created_at: '2026-08-30T00:00:00Z' },
      { role: 'system', content: 'hidden' },
    ])).toEqual([
      { id: 'history-0', role: 'user', content: 'hello', status: 'complete', created_at: undefined },
      { id: 'history-1', role: 'assistant', content: 'world', status: 'complete', created_at: '2026-08-30T00:00:00Z' },
    ])
  })

  it('accepts only complete inflight turn journal entries', () => {
    expect(parseInflightTurn('{"stream_id":"stream-1","session_id":"session-1"}')).toEqual({ stream_id: 'stream-1', session_id: 'session-1' })
    expect(parseInflightTurn('{"stream_id":"stream-1"}')).toBeNull()
    expect(parseInflightTurn('not-json')).toBeNull()
  })

  it('groups and searches sessions using metadata without changing order within a day', () => {
    const sessions = [
      { session_id: 'two', title: 'Build API', updated_at: '2026-08-30T12:00:00Z', tags: ['backend'] },
      { session_id: 'one', title: 'Design notes', updated_at: '2026-08-29T12:00:00Z', project: 'studio' },
    ]
    expect(filterSessions(sessions, 'backend')).toHaveLength(1)
    expect(groupSessionsByDate(sessions, new Date('2026-08-30T18:00:00Z')).map((group) => group.label)).toEqual(['Today', 'Yesterday'])
  })

  it('normalizes usage and approval events into visible state', () => {
    let state = reduceChatEvent(initialChatState, { type: 'usage', data: { prompt_tokens: 12, completion_tokens: 8, total_tokens: 20, context_window: 100 } })
    state = reduceChatEvent(state, { type: 'approval', data: { run_id: 'run-1', name: 'terminal' } })
    expect(state.usage).toEqual({ input: 12, output: 8, total: 20, contextLimit: 100 })
    expect(state.approvals[0].id).toBe('run-1')
  })
})
