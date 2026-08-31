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

  it('turns a non-JSON legacy proxy response into an actionable HTTP error', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('404 page not found\n', { status: 404 }))

    await expect(getInsights()).rejects.toThrow('Request failed (404): 404 page not found')
  })
})
