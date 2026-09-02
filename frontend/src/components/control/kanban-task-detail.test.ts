import { describe, expect, it, vi } from 'vitest'
import { isKanbanMutationAvailable, runTaskDetailAction, runTaskDetailMutation } from './kanban-view'

describe('task detail mutation handling', () => {
  it('reports mutations unavailable unless capabilities explicitly allow them', () => {
    expect(isKanbanMutationAvailable(undefined)).toBe(false)
    expect(isKanbanMutationAvailable({ available: false } as any)).toBe(false)
    expect(isKanbanMutationAvailable({ available: true } as any)).toBe(true)
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
