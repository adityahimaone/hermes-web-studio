import { afterEach, describe, expect, it, vi } from 'vitest'
import { getInsights } from './insights-client'

describe('getInsights', () => {
  afterEach(() => vi.restoreAllMocks())

  it('reads the server-owned insights route', async () => {
    const response = { summary: { sessions: 0, messages: 0, user_messages: 0, assistant_messages: 0 } }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(response), { status: 200 }))

    await expect(getInsights()).resolves.toEqual(response)
    expect(fetchMock).toHaveBeenCalledWith('/api/operator/insights', { signal: undefined })
  })

  it('surfaces the server error without inventing fallback data', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ message: 'Insights state could not be read.' }), { status: 503 }))

    await expect(getInsights()).rejects.toThrow('Insights state could not be read.')
  })
})
