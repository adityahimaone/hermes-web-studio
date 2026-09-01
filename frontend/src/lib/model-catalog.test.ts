import { describe, expect, it } from 'vitest'
import { groupModelCatalog, searchModelCatalog } from './model-catalog'

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
})
