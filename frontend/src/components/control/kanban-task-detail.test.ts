import { describe, expect, it, vi } from 'vitest'
import { runTaskDetailAction } from './kanban-view'

describe('task detail mutation handling', () => {
  it('surfaces refresh rejection through visible error callback', async () => {
    const onError = vi.fn()
    await runTaskDetailAction(
      () => Promise.resolve(),
      () => Promise.reject(new Error('refresh failed')),
      onError,
    )
    expect(onError).toHaveBeenCalledWith('refresh failed')
  })
})
