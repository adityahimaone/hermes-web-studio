export type ChatEventType =
  | 'token'
  | 'reasoning'
  | 'tool'
  | 'tool_complete'
  | 'done'
  | 'cancel'
  | 'apperror'
  | 'subagent'
  | 'approval'
  | 'usage'

export interface ChatEvent {
  type: ChatEventType
  data: Record<string, unknown>
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  status?: 'streaming' | 'complete' | 'error' | 'cancelled'
  created_at?: string
}

export interface SessionSummary {
  session_id: string
  title?: string
  updated_at?: string
  created_at?: string
  pinned?: boolean
  archived?: boolean
  project_id?: string
  project?: string
  tags?: string[]
  [key: string]: unknown
}

export interface SessionGroup {
  label: string
  sessions: SessionSummary[]
}

function sessionSearchText(session: SessionSummary): string {
  return [session.session_id, session.title, session.project_id, session.project, session.tags]
    .filter((value) => typeof value === 'string' || Array.isArray(value))
    .flatMap((value) => Array.isArray(value) ? value : [value])
    .join(' ')
    .toLocaleLowerCase()
}

export function filterSessions(sessions: SessionSummary[], query: string): SessionSummary[] {
  const needle = query.trim().toLocaleLowerCase()
  if (!needle) return sessions
  return sessions.filter((session) => sessionSearchText(session).includes(needle))
}

function sessionDate(session: SessionSummary): Date | null {
  const raw = session.updated_at || session.created_at
  if (!raw) return null
  const date = new Date(raw)
  return Number.isNaN(date.getTime()) ? null : date
}

export function groupSessionsByDate(sessions: SessionSummary[], now = new Date()): SessionGroup[] {
  const start = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate())
  const groups = new Map<string, SessionSummary[]>()
  for (const session of [...sessions].sort((a, b) => (sessionDate(b)?.getTime() || 0) - (sessionDate(a)?.getTime() || 0))) {
    const date = sessionDate(session)
    const age = date ? Math.floor((start - Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate())) / 86400000) : -1
    const label = age === 0 ? 'Today' : age === 1 ? 'Yesterday' : age >= 2 && age < 7 ? 'This week' : 'Earlier'
    const list = groups.get(label) || []
    list.push(session)
    groups.set(label, list)
  }
  return ['Today', 'Yesterday', 'This week', 'Earlier'].flatMap((label) => {
    const list = groups.get(label)
    return list?.length ? [{ label, sessions: list }] : []
  })
}

export function replaceSession(sessions: SessionSummary[], updated: SessionSummary): SessionSummary[] {
  const found = sessions.some((session) => session.session_id === updated.session_id)
  return (found ? sessions.map((session) => session.session_id === updated.session_id ? updated : session) : [updated, ...sessions])
    .sort((a, b) => Number(Boolean(b.pinned)) - Number(Boolean(a.pinned)) || (new Date(b.updated_at || b.created_at || 0).getTime() - new Date(a.updated_at || a.created_at || 0).getTime()))
}

export function removeSession(sessions: SessionSummary[], sessionId: string): SessionSummary[] {
  return sessions.filter((session) => session.session_id !== sessionId)
}

export function editMessageHistory(messages: ChatMessage[], messageId: string): { messages: ChatMessage[]; content: string } | null {
  const index = messages.findIndex((message) => message.id === messageId && message.role === 'user')
  if (index < 0) return null
  return { messages: messages.slice(0, index), content: messages[index].content }
}

export function retryMessageHistory(messages: ChatMessage[], messageId?: string): { messages: ChatMessage[]; content: string } | null {
  const index = messageId ? messages.findIndex((message) => message.id === messageId && message.role === 'user') : [...messages].map((message) => message.role).lastIndexOf('user')
  if (index < 0) return null
  return { messages: messages.slice(0, index), content: messages[index].content }
}

export interface SessionDetail extends SessionSummary {
  messages: ChatMessage[]
}

export function normalizeSessionMessage(message: unknown, index: number): ChatMessage | null {
  if (!message || typeof message !== 'object') return null
  const value = message as Record<string, unknown>
  const role = value.role === 'assistant' || value.role === 'user' ? value.role : null
  if (!role || typeof value.content !== 'string') return null
  return { id: typeof value.id === 'string' ? value.id : `history-${index}`, role, content: value.content, status: 'complete', created_at: typeof value.created_at === 'string' ? value.created_at : undefined }
}

export function normalizeSessionMessages(messages: unknown): ChatMessage[] {
  if (!Array.isArray(messages)) return []
  return messages.flatMap((message, index) => {
    const normalized = normalizeSessionMessage(message, index)
    return normalized ? [normalized] : []
  })
}

export interface ToolActivity {
  id: string
  name: string
  status: 'running' | 'complete' | 'error'
  args?: unknown
  result?: unknown
}

export interface SubagentActivity { id: string; name: string; status: 'running' | 'complete' | 'error'; task?: string }
export interface ApprovalRequest { id: string; name: string; command?: string; reason?: string; status: 'pending' | 'approved' | 'denied' }
export interface TokenUsage { input?: number; output?: number; total?: number; contextLimit?: number }

export interface ChatState {
  answer: string
  reasoning: string
  tools: ToolActivity[]
  subagents: SubagentActivity[]
  approvals: ApprovalRequest[]
  usage?: TokenUsage
  status: 'idle' | 'streaming' | 'complete' | 'error' | 'cancelled'
  error?: string
}

export const initialChatState: ChatState = {
  answer: '',
  reasoning: '',
  tools: [],
  subagents: [],
  approvals: [],
  status: 'idle',
}

function eventText(data: Record<string, unknown>, key: string) {
  return typeof data[key] === 'string' ? data[key] : ''
}

export function reduceChatEvent(state: ChatState, event: ChatEvent): ChatState {
  switch (event.type) {
    case 'token':
      return { ...state, status: 'streaming', answer: state.answer + eventText(event.data, 'text') }
    case 'reasoning':
      return { ...state, status: 'streaming', reasoning: state.reasoning + eventText(event.data, 'text') }
    case 'tool': {
      const eventID = eventText(event.data, 'tid')
      const existing = state.tools.find((tool) => eventID ? tool.id === eventID : tool.name === eventText(event.data, 'name') && tool.status === 'running')
      const id = eventID || existing?.id || `${eventText(event.data, 'name')}-${state.tools.length}`
      if (existing) {
        return { ...state, status: 'streaming', tools: state.tools.map((tool) => tool.id === id ? { ...tool, args: event.data.args ?? tool.args, result: event.data.result ?? tool.result } : tool) }
      }
      return {
        ...state,
        status: 'streaming',
        tools: [...state.tools, { id, name: eventText(event.data, 'name') || 'Hermes tool', status: 'running', args: event.data.args }],
      }
    }
    case 'tool_complete': {
      const id = eventText(event.data, 'tid')
      const name = eventText(event.data, 'name')
      let matched = false
      const tools = state.tools.map((tool) => {
        if (!matched && ((id && tool.id === id) || (!id && tool.name === name && tool.status === 'running'))) {
          matched = true
          return { ...tool, status: event.data.is_error ? 'error' : 'complete', result: event.data.result } as ToolActivity
        }
        return tool
      })
      return { ...state, tools }
    }
    case 'subagent': {
      const eventID = eventText(event.data, 'sid') || eventText(event.data, 'id')
      const existing = state.subagents.find((agent) => eventID ? agent.id === eventID : agent.name === eventText(event.data, 'name') && agent.status === 'running')
      const id = eventID || existing?.id || `subagent-${state.subagents.length}`
      if (existing) return { ...state, status: 'streaming', subagents: state.subagents.map((agent) => agent.id === id ? { ...agent, name: eventText(event.data, 'name') || agent.name, status: eventText(event.data, 'status') === 'error' ? 'error' : eventText(event.data, 'status') === 'complete' ? 'complete' : agent.status, task: eventText(event.data, 'task') || agent.task } : agent) }
      return { ...state, status: 'streaming', subagents: [...state.subagents, { id, name: eventText(event.data, 'name') || 'Hermes subagent', status: eventText(event.data, 'status') === 'error' ? 'error' : 'running', task: eventText(event.data, 'task') || undefined }] }
    }
    case 'approval': {
      const id = eventText(event.data, 'approval_id') || eventText(event.data, 'run_id') || eventText(event.data, 'id') || `approval-${state.approvals.length}`
      return { ...state, status: 'streaming', approvals: [...state.approvals, { id, name: eventText(event.data, 'name') || 'Action approval', command: eventText(event.data, 'command') || undefined, reason: eventText(event.data, 'reason') || undefined, status: 'pending' }] }
    }
    case 'usage': {
      const number = (...keys: string[]) => keys.map((key) => event.data[key]).find((value): value is number => typeof value === 'number')
      return { ...state, usage: { input: number('input_tokens', 'prompt_tokens', 'input'), output: number('output_tokens', 'completion_tokens', 'output'), total: number('total_tokens', 'total'), contextLimit: number('context_limit', 'context_window', 'contextLimit') } }
    }
    case 'done':
      return {
        ...state,
        answer: state.answer || eventText(event.data, 'answer'),
        status: 'complete',
      }
    case 'cancel':
      return { ...state, status: 'cancelled' }
    case 'apperror':
      return { ...state, status: 'error', error: eventText(event.data, 'message') || 'Hermes could not complete the request.' }
  }
}
