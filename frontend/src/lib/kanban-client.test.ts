import { describe, expect, it, vi } from 'vitest'
import { taskEdit, taskLink } from './kanban-client'

describe('kanban mutations', () => {
  it('posts edit payload to task action contract', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{"ok":true}', { status: 200 }))
    await taskEdit('board-a', 't_1', { result: 'done', summary: 'handoff' })
    expect(fetchMock.mock.calls[0][0]).toBe('/api/kanban/tasks/t_1/actions/edit?board=board-a')
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ result: 'done', summary: 'handoff' })
    fetchMock.mockRestore()
  })

  it('posts linked child ID to link action contract', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{"ok":true}', { status: 200 }))
    await taskLink('board-a', 't_1', 't_2')
    expect(fetchMock.mock.calls[0][0]).toBe('/api/kanban/tasks/t_1/actions/link?board=board-a')
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ child_id: 't_2' })
    fetchMock.mockRestore()
  })
})
