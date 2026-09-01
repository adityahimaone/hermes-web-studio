import { describe, expect, it } from 'vitest'
import { utf8Length } from './model-catalog'
import { findCatalogModel, groupModelCatalog, normalizeModelCatalog, searchModelCatalog, validModelSelection } from './model-catalog'

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

  it('truncates by UTF-8 bytes without splitting characters', () => {
    expect(utf8Length('🙂🙂🙂', 8)).toBe('🙂🙂')
  })

  it('rejects unavailable model/provider selections', () => {
    expect(validModelSelection(models, 'missing', 'openai')).toBe(false)
    expect(validModelSelection(models, 'gpt-test', 'anthropic')).toBe(false)
    expect(validModelSelection(models, 'gpt-test', 'openai')).toBe(true)
  })

  it('resolves raw catalog entries through normalized model/provider mapping', () => {
    expect(findCatalogModel([{ id: ' gpt-test ', label: 'GPT', provider: ' openai ', aliases: [], capabilities: [], available: true }], 'gpt-test', 'openai')).toMatchObject({ id: 'gpt-test', provider: 'openai' })
  })
})
