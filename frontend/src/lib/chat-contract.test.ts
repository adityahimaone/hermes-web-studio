import { describe, expect, it } from 'vitest'
import { initialChatState, normalizeSessionMessages, reduceChatEvent } from './chat-contract'

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
})
