import { useCallback, useEffect, useRef, useState } from 'react'
import { cancelChat, deleteSession, getSession, getSessions, renameSession, resolveApproval, setSessionArchived, setSessionPinned, startChat, truncateSession, uploadAttachment } from '../lib/api-client'
import { initialChatState, normalizeSessionMessages, reduceChatEvent, type ChatEvent, type ChatEventType, type ChatMessage, type ChatState, type SessionSummary } from '../lib/chat-contract'

const supportedEvents: ChatEventType[] = ['token', 'reasoning', 'tool', 'tool_complete', 'subagent', 'approval', 'usage', 'done', 'cancel', 'apperror']
const newId = () => crypto.randomUUID()

export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [streamState, setStreamState] = useState<ChatState>(initialChatState)
  const [activeSessionId, setActiveSessionId] = useState<string>(newId)
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [sessionLoading, setSessionLoading] = useState(true)
  const [sessionError, setSessionError] = useState<string>()
  const [queuedMessages, setQueuedMessages] = useState<string[]>([])
  const [draft, setDraft] = useState('')
  const sourceRef = useRef<EventSource | null>(null)
  const streamIdRef = useRef<string | null>(null)
  const queueRef = useRef<string[]>([])
  const activeSessionRef = useRef(activeSessionId)
  const answerRef = useRef('')
  const terminalRef = useRef<string | null>(null)
  const pumpRef = useRef<(content: string) => void>(() => undefined)

  const closeSource = useCallback(() => { sourceRef.current?.close(); sourceRef.current = null }, [])
  const refreshSessions = useCallback(async (signal?: AbortSignal) => {
    const result = await getSessions(signal)
    setSessions(result.sessions || [])
    return result.sessions || []
  }, [])
  useEffect(() => { activeSessionRef.current = activeSessionId }, [activeSessionId])
  useEffect(() => {
    const controller = new AbortController()
    refreshSessions(controller.signal).then(() => setSessionLoading(false)).catch((error) => {
      if (error?.name !== 'AbortError') setSessionError(error instanceof Error ? error.message : 'Unable to load sessions.')
      setSessionLoading(false)
    })
    return () => controller.abort()
  }, [refreshSessions])
  useEffect(() => closeSource, [closeSource])

  const finish = useCallback((streamId: string, state: ChatState, status: ChatMessage['status'] = 'complete') => {
    if (terminalRef.current === streamId) return
    terminalRef.current = streamId; closeSource(); streamIdRef.current = null
    if (state.answer || status !== 'complete') setMessages((current) => [...current, { id: newId(), role: 'assistant', content: state.answer, status }])
    setStreamState(initialChatState)
    void refreshSessions().catch(() => undefined)
    const next = queueRef.current.shift()
    setQueuedMessages([...queueRef.current])
    if (next) pumpRef.current(next)
  }, [closeSource])

  const pump = useCallback(async (content: string, files: File[] = [], baseMessages?: ChatMessage[]) => {
    const clean = content.trim(); if (!clean) return
    closeSource(); answerRef.current = ''; terminalRef.current = null
    setMessages((current) => [...(baseMessages || current), { id: newId(), role: 'user', content: clean, status: 'complete' }])
    setStreamState({ ...initialChatState, status: 'streaming' })
    try {
      const uploaded = await Promise.all(files.map((file) => uploadAttachment(file, activeSessionRef.current)))
      const started = await startChat({ session_id: activeSessionRef.current, message: clean, attachment_ids: uploaded.map((file) => file.id) })
      streamIdRef.current = started.stream_id
      setActiveSessionId(started.session_id || activeSessionRef.current)
      const source = new EventSource(`/api/chat/stream?stream_id=${encodeURIComponent(started.stream_id)}`)
      sourceRef.current = source
      supportedEvents.forEach((type) => source.addEventListener(type, (raw) => {
        const event = raw as MessageEvent<string>
        let data: Record<string, unknown>
        try { data = JSON.parse(event.data) as Record<string, unknown> } catch { data = { message: event.data } }
        setStreamState((current) => {
          if (terminalRef.current === started.stream_id) return current
          const next = reduceChatEvent(current, { type, data } as ChatEvent)
          if (type === 'token' && typeof data.text === 'string') answerRef.current += data.text
          if (type === 'done') finish(started.stream_id, { ...next, answer: answerRef.current || next.answer })
          if (type === 'cancel') finish(started.stream_id, next, 'cancelled')
          if (type === 'apperror') finish(started.stream_id, next, 'error')
          return type === 'done' || type === 'cancel' || type === 'apperror' ? current : next
        })
      }))
      source.onerror = () => {
        if (source.readyState === EventSource.CLOSED && terminalRef.current !== started.stream_id) setStreamState((state) => ({ ...state, status: 'error', error: 'The Hermes stream closed unexpectedly. Reconnecting…' }))
      }
    } catch (error) { setStreamState((state) => ({ ...state, status: 'error', error: error instanceof Error ? error.message : 'Unable to start Hermes.' })) }
  }, [closeSource, finish, refreshSessions])
  pumpRef.current = pump

  const send = useCallback((content: string, files: File[] = []) => {
    const clean = content.trim(); if (!clean) return
    setDraft('')
    if (streamState.status === 'streaming' || streamIdRef.current) { queueRef.current.push(clean); setQueuedMessages([...queueRef.current]) } else pump(clean, files)
  }, [pump, streamState.status])
  const retry = useCallback((message: ChatMessage) => {
    const index = messages.findIndex((item) => item.id === message.id)
    if (index < 0) return
    void truncateSession(activeSessionRef.current, index).catch(() => undefined)
    setMessages((current) => current.slice(0, index))
    setDraft('')
    pump(message.content, [], messages.slice(0, index))
  }, [messages, pump])
  const edit = useCallback((message: ChatMessage) => {
    const index = messages.findIndex((item) => item.id === message.id)
    if (index < 0) return
    void truncateSession(activeSessionRef.current, index).catch(() => undefined)
    setMessages((current) => current.slice(0, index))
    setDraft(message.content)
  }, [messages])
  const approve = useCallback(async (id: string, decision: 'approved' | 'denied') => {
    await resolveApproval(id, decision)
    setStreamState((state) => ({ ...state, approvals: state.approvals.map((item) => item.id === id ? { ...item, status: decision } : item) }))
  }, [])
  const cancel = useCallback(async () => {
    const streamId = streamIdRef.current; if (!streamId) return
    await cancelChat(streamId).catch(() => undefined); finish(streamId, { ...initialChatState, status: 'cancelled' }, 'cancelled')
  }, [finish])
  const selectSession = useCallback(async (sessionId: string) => {
    closeSource(); setActiveSessionId(sessionId); setStreamState(initialChatState); queueRef.current = []; setQueuedMessages([]); setSessionLoading(true); setSessionError(undefined)
    try { const detail = await getSession(sessionId); setMessages(normalizeSessionMessages(detail.messages)) } catch (error) { setMessages([]); setSessionError(error instanceof Error ? error.message : 'Unable to load this session.') } finally { setSessionLoading(false) }
  }, [closeSource])
  const rename = useCallback(async (sessionId: string) => {
    const current = sessions.find((item) => item.session_id === sessionId)
    const title = window.prompt('Rename session', current?.title || '')?.trim()
    if (!title) return
    const updated = await renameSession(sessionId, title)
    setSessions((items) => items.map((item) => item.session_id === sessionId ? updated : item))
  }, [sessions])
  const pin = useCallback(async (sessionId: string, pinned: boolean) => {
    const updated = await setSessionPinned(sessionId, pinned)
    setSessions((items) => items.map((item) => item.session_id === sessionId ? updated : item))
  }, [])
  const archive = useCallback(async (sessionId: string, archived: boolean) => {
    const updated = await setSessionArchived(sessionId, archived)
    setSessions((items) => items.map((item) => item.session_id === sessionId ? updated : item))
  }, [])
  const remove = useCallback(async (sessionId: string) => {
    if (!window.confirm('Delete this session?')) return
    await deleteSession(sessionId)
    setSessions((items) => items.filter((item) => item.session_id !== sessionId))
    if (activeSessionRef.current === sessionId) reset()
  }, [])
  const reset = useCallback(() => { closeSource(); queueRef.current = []; setQueuedMessages([]); setMessages([]); setStreamState(initialChatState); setDraft(''); setActiveSessionId(newId()) }, [closeSource])

  return { messages, streamState, send, cancel, reset, retry, edit, approve, draft, setDraft, sessions, selectSession, rename, pin, archive, remove, activeSessionId, sessionLoading, sessionError, queuedMessages, isStreaming: streamState.status === 'streaming' }
}
