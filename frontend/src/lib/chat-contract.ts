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
  [key: string]: unknown
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
      const id = eventText(event.data, 'tid') || `${eventText(event.data, 'name')}-${state.tools.length}`
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
      const id = eventText(event.data, 'sid') || eventText(event.data, 'id') || `subagent-${state.subagents.length}`
      return { ...state, status: 'streaming', subagents: [...state.subagents, { id, name: eventText(event.data, 'name') || 'Hermes subagent', status: eventText(event.data, 'status') === 'error' ? 'error' : 'running', task: eventText(event.data, 'task') || undefined }] }
    }
    case 'approval': {
      const id = eventText(event.data, 'approval_id') || eventText(event.data, 'id') || `approval-${state.approvals.length}`
      return { ...state, status: 'streaming', approvals: [...state.approvals, { id, name: eventText(event.data, 'name') || 'Action approval', command: eventText(event.data, 'command') || undefined, reason: eventText(event.data, 'reason') || undefined, status: 'pending' }] }
    }
    case 'usage': {
      const number = (...keys: string[]) => keys.map((key) => event.data[key]).find((value): value is number => typeof value === 'number')
      return { ...state, usage: { input: number('input_tokens', 'input'), output: number('output_tokens', 'output'), total: number('total_tokens', 'total'), contextLimit: number('context_limit', 'contextLimit') } }
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
