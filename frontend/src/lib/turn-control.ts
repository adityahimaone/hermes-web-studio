export type TurnMode = 'queue' | 'interrupt' | 'steer'

export interface PendingTurn {
  content: string
  attachmentNames: string[]
}

export function normalizeTurnMode(value: unknown): TurnMode {
  return value === 'interrupt' || value === 'steer' ? value : 'queue'
}

export function planTurn<T extends PendingTurn>(mode: TurnMode, pending: T[], next: T): T[] {
  if (mode === 'queue') return [...pending, next]
  if (mode === 'interrupt') return [next]
  return [...pending.filter((item) => item.content !== next.content), next]
}

export function clearPendingTurns(): PendingTurn[] {
  return []
}
