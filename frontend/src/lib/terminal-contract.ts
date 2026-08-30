export type TerminalCapability = {
  id: 'terminal'
  available: boolean
  state: 'available' | 'unavailable'
  reason: string
  message: string
}

export function terminalCapability(value: unknown): TerminalCapability {
  if (!value || typeof value !== 'object') throw new Error('Terminal capability could not be read.')
  const item = value as Record<string, unknown>
  if (item.id !== 'terminal' || item.available !== false || item.state !== 'unavailable') throw new Error('Terminal capability returned an unsafe state.')
  if (typeof item.reason !== 'string' || typeof item.message !== 'string') throw new Error('Terminal capability is incomplete.')
  return { id: 'terminal', available: false, state: 'unavailable', reason: item.reason, message: item.message }
}
