export interface SlashCommand {
  name: string
  description: string
  usage?: string
}

export const slashCommands: readonly SlashCommand[] = [
  { name: '/help', description: 'Show available local commands' },
  { name: '/clear', description: 'Start with an empty conversation' },
]

export function slashCommandInput(value: string): { query: string; args: string } | null {
  const trimmed = value.trimStart()
  if (!trimmed.startsWith('/')) return null
  const match = trimmed.match(/^(\/[^\s]*)(?:\s+(.*))?$/s)
  if (!match) return { query: trimmed.toLocaleLowerCase(), args: '' }
  return { query: match[1].toLocaleLowerCase(), args: match[2] || '' }
}

export function slashCommandSuggestions(value: string): SlashCommand[] {
  const input = slashCommandInput(value)
  if (!input || input.args || /\s/.test(value.trimStart())) return []
  return slashCommands.filter((command) => command.name.startsWith(input.query))
}

export function localSlashCommand(value: string): SlashCommand | null {
  const input = slashCommandInput(value)
  if (!input || input.args) return null
  return slashCommands.find((command) => command.name === input.query) || null
}
