import { describe, expect, it, vi } from 'vitest'
import { getModelCatalog } from './api-client'

describe('model catalog API contract', () => {
  it('rejects unknown status and malformed model data', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ status: 'surprise', models: [{ arbitrary: true }] }), { status: 200 })))
    await expect(getModelCatalog()).rejects.toThrow('Invalid model catalog response')
    vi.unstubAllGlobals()
  })

  it('rejects malformed entries even when status is unavailable', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ status: 'unavailable', models: [{ arbitrary: true }] }), { status: 200 })))
    await expect(getModelCatalog()).rejects.toThrow('Invalid model catalog response')
    vi.unstubAllGlobals()
  })
})
