import { describe, expect, it } from 'vitest'
import { groupModelCatalog, normalizeModelCatalog, searchModelCatalog } from './model-catalog'

describe('model catalog', () => {
  const models = [
    { id: 'gpt-test', label: 'GPT Test', provider: 'openai', aliases: ['fast'], capabilities: [], available: true },
    { id: 'claude-test', label: 'Claude Test', provider: 'anthropic', aliases: [], capabilities: ['reasoning'], available: true },
  ]

  it('groups available models by provider', () => {
    expect(groupModelCatalog(models)).toEqual([
      { provider: 'anthropic', models: [models[1]] },
      { provider: 'openai', models: [models[0]] },
    ])
  })

  it('searches model ids, labels, providers, and aliases', () => {
    expect(searchModelCatalog(models, 'fast')).toEqual([models[0]])
  })

  it('normalizes and deduplicates model IDs and aliases', () => {
    expect(normalizeModelCatalog([
      { id: ' model-a ', label: ' Model A\n', provider: ' openai ', aliases: ['fast\n', 'fast', '  '], capabilities: [], available: true },
      { id: 'model-a', label: 'other', provider: 'openai', aliases: [], capabilities: [], available: true },
    ])).toEqual([{ id: 'model-a', label: 'Model A', provider: 'openai', aliases: ['fast'], capabilities: [], available: true }])
  })

  it('reports no visible results after filtering', () => {
    expect(searchModelCatalog(models, 'missing')).toHaveLength(0)
  })
})
