import { describe, expect, it, vi } from 'vitest'
import { isCurrentKanbanMutation, isCurrentKanbanRefresh, isKanbanMutationAvailable, runTaskDetailAction, runTaskDetailMutation } from './kanban-view'

describe('task detail mutation handling', () => {
  it('reports mutations unavailable unless capabilities explicitly allow them', () => {
    expect(isKanbanMutationAvailable(undefined)).toBe(false)
    expect(isKanbanMutationAvailable({ available: false } as any)).toBe(false)
    expect(isKanbanMutationAvailable({ available: true } as any)).toBe(true)
  })

  it('accepts only latest board refresh response', () => {
    expect(isCurrentKanbanRefresh(2, 2)).toBe(true)
    expect(isCurrentKanbanRefresh(1, 2)).toBe(false)
  })

  it('accepts only latest task mutation response', () => {
    expect(isCurrentKanbanMutation(2, 2)).toBe(true)
    expect(isCurrentKanbanMutation(1, 2)).toBe(false)
  })

  it('surfaces refresh rejection through visible error callback', async () => {
    const onError = vi.fn()
    await runTaskDetailAction(
      () => Promise.resolve(),
      () => Promise.reject(new Error('refresh failed')),
      onError,
    )
    expect(onError).toHaveBeenCalledWith('refresh failed')
  })

  it('reloads displayed detail after successful mutation', async () => {
    const setDetail = vi.fn()
    await runTaskDetailMutation(() => Promise.resolve(), () => Promise.resolve(), () => Promise.resolve({ id: 't_1', result: 'new' }), setDetail, vi.fn())
    expect(setDetail).toHaveBeenCalledWith({ id: 't_1', result: 'new' })
  })
})
