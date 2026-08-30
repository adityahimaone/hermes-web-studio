import { useCallback, useEffect, useRef, useState } from 'react'
import { cancelChat, deleteSession, duplicateSession, getSession, getSessions, renameSession, resolveApproval, searchSessions, setSessionArchived, setSessionPinned, startChat, streamUrl, truncateSession, uploadAttachment, type ApprovalChoice } from '../lib/api-client'
import { initialChatState, normalizeSessionMessages, reduceChatEvent, type ChatEvent, type ChatEventType, type ChatMessage, type ChatState, type SessionSummary } from '../lib/chat-contract'

const supportedEvents: ChatEventType[] = ['token', 'reasoning', 'tool', 'tool_complete', 'subagent', 'approval', 'usage', 'done', 'cancel', 'apperror']
const newId = () => crypto.randomUUID()
type QueuedTurn = { content: string; files: File[] }
type TurnOptions = { model?: string; provider?: string }

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
  const queueRef = useRef<(QueuedTurn & { options?: TurnOptions })[]>([])
  const activeSessionRef = useRef(activeSessionId)
  const answerRef = useRef('')
  const terminalRef = useRef<string | null>(null)
  const lastEventIdRef = useRef(0)
  const chatStateRef = useRef<ChatState>(initialChatState)
  const pumpRef = useRef<(content: string, files?: File[], baseMessages?: ChatMessage[], options?: TurnOptions) => void>(() => undefined)

  const closeSource = useCallback(() => { sourceRef.current?.close(); sourceRef.current = null }, [])
  const refreshSessions = useCallback(async (signal?: AbortSignal) => {
    const result = await getSessions(signal)
    setSessions(result.sessions || [])
    return result.sessions || []
  }, [])
  const searchSessionList = useCallback(async (query: string) => {
    const result = await searchSessions(query)
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
    chatStateRef.current = initialChatState
    setStreamState(initialChatState)
    void refreshSessions().catch(() => undefined)
    const next = queueRef.current.shift()
    setQueuedMessages(queueRef.current.map((item) => item.content))
    if (next) pumpRef.current(next.content, next.files, undefined, next.options)
  }, [closeSource])

  const pump = useCallback(async (content: string, files: File[] = [], baseMessages?: ChatMessage[], options?: TurnOptions) => {
    const clean = content.trim(); if (!clean) return
    closeSource(); answerRef.current = ''; terminalRef.current = null; lastEventIdRef.current = 0; chatStateRef.current = initialChatState
    setMessages((current) => [...(baseMessages || current), { id: newId(), role: 'user', content: clean, status: 'complete' }])
    setStreamState({ ...initialChatState, status: 'streaming' })
    try {
      const uploaded = await Promise.all(files.map((file) => uploadAttachment(file, activeSessionRef.current)))
      const started = await startChat({ session_id: activeSessionRef.current, message: clean, model: options?.model, provider: options?.provider, attachment_ids: uploaded.map((file) => file.id) })
      streamIdRef.current = started.stream_id
      setActiveSessionId(started.session_id || activeSessionRef.current)
      const source = new EventSource(streamUrl(started.stream_id))
      sourceRef.current = source
      supportedEvents.forEach((type) => source.addEventListener(type, (raw) => {
        const event = raw as MessageEvent<string>
        const eventID = Number(event.lastEventId)
        if (Number.isFinite(eventID) && eventID > 0) {
          if (eventID <= lastEventIdRef.current) return
          lastEventIdRef.current = eventID
        }
        let data: Record<string, unknown>
        try { data = JSON.parse(event.data) as Record<string, unknown> } catch { data = { message: event.data } }
        if (terminalRef.current === started.stream_id) return
        const next = reduceChatEvent(chatStateRef.current, { type, data } as ChatEvent)
        chatStateRef.current = next
        if (type === 'token' && typeof data.text === 'string') answerRef.current += data.text
        if (type === 'done') { finish(started.stream_id, { ...next, answer: answerRef.current || next.answer }); return }
        if (type === 'cancel') { finish(started.stream_id, next, 'cancelled'); return }
        if (type === 'apperror') { finish(started.stream_id, next, 'error'); return }
        setStreamState(next)
      }))
      source.onerror = () => {
        if (source.readyState === EventSource.CLOSED && terminalRef.current !== started.stream_id) setStreamState((state) => ({ ...state, status: 'error', error: 'The Hermes stream closed unexpectedly. Reconnecting…' }))
      }
    } catch (error) { setStreamState((state) => ({ ...state, status: 'error', error: error instanceof Error ? error.message : 'Unable to start Hermes.' })) }
  }, [closeSource, finish, refreshSessions])
  pumpRef.current = pump

  const send = useCallback((content: string, files: File[] = [], options?: TurnOptions) => {
    const clean = content.trim(); if (!clean) return
    setDraft('')
    if (streamState.status === 'streaming' || streamIdRef.current) {
      queueRef.current.push({ content: clean, files, options })
      setQueuedMessages(queueRef.current.map((item) => item.content))
    } else pump(clean, files, undefined, options)
  }, [pump, streamState.status])
  const retry = useCallback(async (message: ChatMessage) => {
    const index = messages.findIndex((item) => item.id === message.id)
    if (index < 0) return
    try { await truncateSession(activeSessionRef.current, index) } catch (error) {
      setStreamState({ ...initialChatState, status: 'error', error: error instanceof Error ? error.message : 'Unable to retry this message.' })
      return
    }
    setMessages((current) => current.slice(0, index))
    setDraft('')
    await pump(message.content, [], messages.slice(0, index))
  }, [messages, pump])
  const edit = useCallback(async (message: ChatMessage) => {
    const index = messages.findIndex((item) => item.id === message.id)
    if (index < 0) return
    try { await truncateSession(activeSessionRef.current, index) } catch (error) {
      setStreamState({ ...initialChatState, status: 'error', error: error instanceof Error ? error.message : 'Unable to edit this message.' })
      return
    }
    setMessages((current) => current.slice(0, index))
    setDraft(message.content)
  }, [messages])
  const approve = useCallback(async (id: string, choice: ApprovalChoice) => {
    await resolveApproval(id, choice)
    setStreamState((state) => ({ ...state, approvals: state.approvals.map((item) => item.id === id ? { ...item, status: choice === 'deny' ? 'denied' : 'approved' } : item) }))
  }, [])
  const cancel = useCallback(async () => {
    const streamId = streamIdRef.current; if (!streamId) return
    await cancelChat(streamId).catch(() => undefined); finish(streamId, { ...initialChatState, status: 'cancelled' }, 'cancelled')
  }, [finish])
  const selectSession = useCallback(async (sessionId: string) => {
    closeSource(); setActiveSessionId(sessionId); chatStateRef.current = initialChatState; setStreamState(initialChatState); queueRef.current = []; setQueuedMessages([]); setSessionLoading(true); setSessionError(undefined)
    try { const detail = await getSession(sessionId); setMessages(normalizeSessionMessages(detail.messages)) } catch (error) { setMessages([]); setSessionError(error instanceof Error ? error.message : 'Unable to load this session.') } finally { setSessionLoading(false) }
  }, [closeSource])
  const rename = useCallback(async (sessionId: string, title: string) => {
    if (!title.trim()) return
    const updated = await renameSession(sessionId, title.trim())
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
    await deleteSession(sessionId)
    setSessions((items) => items.filter((item) => item.session_id !== sessionId))
    if (activeSessionRef.current === sessionId) reset()
  }, [])
  const duplicate = useCallback(async (sessionId: string) => {
    const created = await duplicateSession(sessionId)
    setSessions((items) => [created, ...items])
  }, [])
  const reset = useCallback(() => { closeSource(); queueRef.current = []; setQueuedMessages([]); setMessages([]); chatStateRef.current = initialChatState; setStreamState(initialChatState); setDraft(''); setActiveSessionId(newId()) }, [closeSource])

  return { messages, streamState, send, cancel, reset, retry, edit, approve, draft, setDraft, sessions, selectSession, searchSessions: searchSessionList, rename, pin, archive, remove, duplicate, activeSessionId, sessionLoading, sessionError, queuedMessages, isStreaming: streamState.status === 'streaming' }
}
