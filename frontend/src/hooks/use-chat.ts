import { useCallback, useEffect, useRef, useState } from 'react'
import { cancelChat, deleteSession, duplicateSession, getSession, getSessions, renameSession, resolveApproval, searchSessions, setSessionArchived, setSessionPinned, startChat, streamUrl, truncateSession, uploadAttachment, type ApprovalChoice } from '../lib/api-client'
import { initialChatState, normalizeSessionMessages, parseInflightTurn, reduceChatEvent, type ChatEvent, type ChatEventType, type ChatMessage, type ChatState, type SessionSummary } from '../lib/chat-contract'
import { appendCompletedAssistant, dedupeRestoredAssistant, isCurrentConversation, isCurrentPump, normalizeClientError, normalizeRestoreError, pollSessionUntilSettled, queuedTurnBaseline } from '../lib/conversation-runtime'
import { planTurn, type PendingTurn, type TurnMode } from '../lib/turn-control'

const supportedEvents: ChatEventType[] = ['token', 'reasoning', 'tool', 'tool_complete', 'subagent', 'approval', 'usage', 'done', 'cancel', 'apperror']
const newId = () => crypto.randomUUID()
const inflightTurnKey = 'hermes-web-studio:inflight-turn'
type QueuedTurn = { content: string; files: File[]; attachmentNames: string[] }
type TurnOptions = { model?: string; provider?: string }

export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [streamState, setStreamState] = useState<ChatState>(initialChatState)
  const [activeSessionId, setActiveSessionId] = useState<string>(newId)
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [sessionLoading, setSessionLoading] = useState(true)
  const [sessionError, setSessionError] = useState<string>()
  const [queuedMessages, setQueuedMessages] = useState<PendingTurn[]>([])
  const [draft, setDraft] = useState('')
  const sourceRef = useRef<EventSource | null>(null)
  const streamIdRef = useRef<string | null>(null)
  const queueRef = useRef<(QueuedTurn & { options?: TurnOptions })[]>([])
  const activeSessionRef = useRef(activeSessionId)
  const answerRef = useRef('')
  const terminalRef = useRef<string | null>(null)
  const lastEventIdRef = useRef(0)
  const turnBaselineRef = useRef(0)
  const fallbackStreamRef = useRef<string | null>(null)
  const pollControllerRef = useRef<AbortController | null>(null)
  const pumpControllerRef = useRef<AbortController | null>(null)
  const pendingUserIdRef = useRef<string | null>(null)
  const sessionEpochRef = useRef(0)
  const chatStateRef = useRef<ChatState>(initialChatState)
  const messagesRef = useRef<ChatMessage[]>([])
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
  useEffect(() => { messagesRef.current = messages }, [messages])
  useEffect(() => {
    const controller = new AbortController()
    refreshSessions(controller.signal).then(() => setSessionLoading(false)).catch((error) => {
      if (error?.name !== 'AbortError') setSessionError(normalizeRestoreError(error))
      setSessionLoading(false)
    })
    return () => controller.abort()
  }, [refreshSessions])
  useEffect(() => () => { closeSource(); pollControllerRef.current?.abort(); pumpControllerRef.current?.abort() }, [closeSource])

  const finish = useCallback((streamId: string, sessionId: string, epoch: number, state: ChatState, status: ChatMessage['status'] = 'complete') => {
    if (terminalRef.current === streamId) return
    if (!isCurrentConversation(streamId, sessionId, epoch, { streamId: streamIdRef.current, sessionId: activeSessionRef.current, epoch: sessionEpochRef.current })) return
    terminalRef.current = streamId; closeSource(); streamIdRef.current = null; pollControllerRef.current?.abort(); pollControllerRef.current = null; fallbackStreamRef.current = null; pumpControllerRef.current = null
    if (parseInflightTurn(window.localStorage.getItem(inflightTurnKey))?.stream_id === streamId) window.localStorage.removeItem(inflightTurnKey)
    const completedMessages = activeSessionRef.current === sessionId
      ? appendCompletedAssistant(messagesRef.current, state.answer, status)
      : messagesRef.current
    messagesRef.current = completedMessages
    if (state.answer.trim() && activeSessionRef.current === sessionId) setMessages(completedMessages)
    pendingUserIdRef.current = null
    chatStateRef.current = initialChatState
    setStreamState(status === 'error' ? { ...initialChatState, status: 'error', error: state.error } : initialChatState)
    void refreshSessions().catch(() => undefined)
    const next = queueRef.current.shift()
    setQueuedMessages(queueRef.current.map(({ content, attachmentNames }) => ({ content, attachmentNames })))
    if (next) pumpRef.current(next.content, next.files, completedMessages, next.options)
  }, [closeSource])

  const connectStream = useCallback((streamId: string, sessionId: string, epoch = sessionEpochRef.current) => {
    const source = new EventSource(streamUrl(streamId, lastEventIdRef.current ? String(lastEventIdRef.current) : undefined))
    sourceRef.current = source
    supportedEvents.forEach((type) => source.addEventListener(type, (raw) => {
      const event = raw as MessageEvent<string>
      if (!isCurrentConversation(streamId, sessionId, epoch, { streamId: streamIdRef.current, sessionId: activeSessionRef.current, epoch: sessionEpochRef.current }) || sourceRef.current !== source || terminalRef.current === streamId) return
      const eventID = Number(event.lastEventId)
      if (Number.isFinite(eventID) && eventID > 0) {
        if (eventID <= lastEventIdRef.current) return
        lastEventIdRef.current = eventID
        const journal = parseInflightTurn(window.localStorage.getItem(inflightTurnKey))
        if (journal?.stream_id === streamId) window.localStorage.setItem(inflightTurnKey, JSON.stringify({ ...journal, last_event_id: eventID }))
      }
      let data: Record<string, unknown>
      try { data = JSON.parse(event.data) as Record<string, unknown> } catch { data = { message: event.data } }
      if (!isCurrentConversation(streamId, sessionId, epoch, { streamId: streamIdRef.current, sessionId: activeSessionRef.current, epoch: sessionEpochRef.current }) || sourceRef.current !== source || terminalRef.current === streamId) return
      const next = reduceChatEvent(chatStateRef.current, { type, data } as ChatEvent)
      chatStateRef.current = next
      if (type === 'token' && typeof data.text === 'string') answerRef.current += data.text
      if (type === 'done') { finish(streamId, sessionId, epoch, { ...next, answer: answerRef.current || next.answer }); return }
      if (type === 'cancel') { finish(streamId, sessionId, epoch, next, 'cancelled'); return }
      if (type === 'apperror') { finish(streamId, sessionId, epoch, next, 'error'); return }
      if (activeSessionRef.current === sessionId) setStreamState(next)
    }))
    source.onerror = () => {
      if (!isCurrentConversation(streamId, sessionId, epoch, { streamId: streamIdRef.current, sessionId: activeSessionRef.current, epoch: sessionEpochRef.current }) || sourceRef.current !== source || (source.readyState !== EventSource.CLOSED && source.readyState !== EventSource.CONNECTING) || terminalRef.current === streamId || fallbackStreamRef.current === streamId) return
      source.close()
      fallbackStreamRef.current = streamId
      pollControllerRef.current?.abort()
      const controller = new AbortController()
      pollControllerRef.current = controller
      setStreamState((state) => ({ ...state, status: 'error', error: 'The Hermes stream closed unexpectedly. Reconnecting…' }))
      void pollSessionUntilSettled(
        async (signal) => getSession(sessionId, signal),
        turnBaselineRef.current,
        6,
        1000,
        controller.signal,
      ).then((assistant) => {
        if (!isCurrentConversation(streamId, sessionId, epoch, { streamId: streamIdRef.current, sessionId: activeSessionRef.current, epoch: sessionEpochRef.current }) || terminalRef.current === streamId) return
        if (assistant) {
          const restored = dedupeRestoredAssistant(messagesRef.current, assistant)
          messagesRef.current = restored
          setMessages(restored)
          finish(streamId, sessionId, epoch, { ...initialChatState, answer: '' })
        } else finish(streamId, sessionId, epoch, { ...initialChatState, error: 'Hermes stream did not settle before polling timed out.' }, 'error')
      }).catch((error) => {
        if (!controller.signal.aborted && isCurrentConversation(streamId, sessionId, epoch, { streamId: streamIdRef.current, sessionId: activeSessionRef.current, epoch: sessionEpochRef.current }) && terminalRef.current !== streamId) {
          finish(streamId, sessionId, epoch, { ...initialChatState, error: normalizeRestoreError(error) }, 'error')
        }
      }).finally(() => {
        if (fallbackStreamRef.current === streamId) fallbackStreamRef.current = null
        if (pollControllerRef.current === controller) pollControllerRef.current = null
      })
    }
  }, [finish])

  const pump = useCallback(async (content: string, files: File[] = [], baseMessages?: ChatMessage[], options?: TurnOptions) => {
    const clean = content.trim(); if (!clean) return
    const epoch = sessionEpochRef.current
    const sessionId = activeSessionRef.current
    pumpControllerRef.current?.abort()
    const controller = new AbortController()
    pumpControllerRef.current = controller
    closeSource(); pollControllerRef.current?.abort(); answerRef.current = ''; terminalRef.current = null; fallbackStreamRef.current = null; lastEventIdRef.current = 0; chatStateRef.current = initialChatState
    turnBaselineRef.current = queuedTurnBaseline(baseMessages || messagesRef.current)
    const pendingUserId = newId()
    pendingUserIdRef.current = pendingUserId
    const pendingMessages = [...(baseMessages || messagesRef.current), { id: pendingUserId, role: 'user' as const, content: clean, status: 'complete' as const }]
    messagesRef.current = pendingMessages
    setMessages(pendingMessages)
    setStreamState({ ...initialChatState, status: 'streaming' })

    const removePending = () => {
      if (pendingUserIdRef.current !== pendingUserId) return
      const next = messagesRef.current.filter((message) => message.id !== pendingUserId)
      messagesRef.current = next
      setMessages(next)
      pendingUserIdRef.current = null
    }
    try {
      const uploaded = await Promise.all(files.map((file) => uploadAttachment(file, sessionId, controller.signal)))
      if (controller.signal.aborted || sessionEpochRef.current !== epoch || activeSessionRef.current !== sessionId) { removePending(); return }
      const started = await startChat({ session_id: sessionId, message: clean, model: options?.model, provider: options?.provider, attachment_ids: uploaded.map((file) => file.id) }, controller.signal)
      if (controller.signal.aborted || sessionEpochRef.current !== epoch || activeSessionRef.current !== sessionId) { removePending(); return }
      streamIdRef.current = started.stream_id
      const returnedSessionId = started.session_id || sessionId
      activeSessionRef.current = returnedSessionId
      setActiveSessionId(returnedSessionId)
      window.localStorage.setItem(inflightTurnKey, JSON.stringify({ stream_id: started.stream_id, session_id: returnedSessionId, last_event_id: 0 }))
      connectStream(started.stream_id, returnedSessionId, epoch)
    } catch (error) {
      const ownsLifecycle = isCurrentPump(controller, pumpControllerRef.current, sessionId, epoch, { sessionId: activeSessionRef.current, epoch: sessionEpochRef.current })
      if (!ownsLifecycle) return
      if (controller.signal.aborted) {
        removePending()
        setStreamState(initialChatState)
      } else {
        setStreamState({ ...initialChatState, status: 'error', error: normalizeClientError(error) })
      }
    }
    finally { if (pumpControllerRef.current === controller) pumpControllerRef.current = null }
  }, [closeSource, connectStream, messages, refreshSessions])
  pumpRef.current = pump

  useEffect(() => {
    const journal = parseInflightTurn(window.localStorage.getItem(inflightTurnKey))
    if (!journal) {
      if (window.localStorage.getItem(inflightTurnKey)) window.localStorage.removeItem(inflightTurnKey)
      return
    }
    const epoch = sessionEpochRef.current
    activeSessionRef.current = journal.session_id
    const controller = new AbortController()
    pollControllerRef.current = controller
    void getSession(journal.session_id, controller.signal).then((detail) => {
      if (!isCurrentConversation(journal.stream_id, journal.session_id, epoch, { streamId: journal.stream_id, sessionId: activeSessionRef.current, epoch }) || activeSessionRef.current !== journal.session_id) return
      setActiveSessionId(journal.session_id)
      const restored = normalizeSessionMessages(detail.messages)
      setMessages(restored)
      turnBaselineRef.current = restored.length
      setStreamState({ ...initialChatState, status: 'streaming' })
      streamIdRef.current = journal.stream_id
      terminalRef.current = null
      lastEventIdRef.current = journal.last_event_id || 0
      if (pollControllerRef.current === controller) pollControllerRef.current = null
      connectStream(journal.stream_id, journal.session_id, epoch)
    }).catch((error) => {
      if (controller.signal.aborted || sessionEpochRef.current !== epoch || activeSessionRef.current !== journal.session_id) return
      if (parseInflightTurn(window.localStorage.getItem(inflightTurnKey))?.stream_id === journal.stream_id) window.localStorage.removeItem(inflightTurnKey)
      if (pollControllerRef.current === controller) pollControllerRef.current = null
      streamIdRef.current = null
      fallbackStreamRef.current = null
      chatStateRef.current = initialChatState
      setStreamState({ ...initialChatState, status: 'error', error: normalizeRestoreError(error) })
      setSessionError(normalizeRestoreError(error))
    })
    return () => { controller.abort(); sessionEpochRef.current += 1; closeSource(); pumpControllerRef.current?.abort(); pollControllerRef.current?.abort() }
  }, [closeSource, connectStream])

  const send = useCallback((content: string, files: File[] = [], options?: TurnOptions, mode: TurnMode = 'queue') => {
    const clean = content.trim(); if (!clean) return
    setDraft('')
    if (streamState.status === 'streaming' || streamIdRef.current) {
      const next = { content: clean, files, attachmentNames: files.map((file) => file.name), options }
      const planned = planTurn(mode, queueRef.current, next)
      queueRef.current = planned
      setQueuedMessages(planned.map(({ content: itemContent, attachmentNames }) => ({ content: itemContent, attachmentNames })))
      if (mode !== 'queue') {
        const streamId = streamIdRef.current
        if (streamId) {
          void cancelChat(streamId).catch(() => undefined)
          finish(streamId, activeSessionRef.current, sessionEpochRef.current, { ...initialChatState, status: 'cancelled' }, 'cancelled')
        } else {
          pumpControllerRef.current?.abort()
          const pendingId = pendingUserIdRef.current
          if (pendingId) {
            const nextMessages = messagesRef.current.filter((message) => message.id !== pendingId)
            messagesRef.current = nextMessages
            setMessages(nextMessages)
            pendingUserIdRef.current = null
          }
          chatStateRef.current = initialChatState
          setStreamState(initialChatState)
          queueRef.current = []
          setQueuedMessages([])
          pump(clean, files, messagesRef.current, options)
        }
      }
    } else pump(clean, files, undefined, options)
  }, [finish, pump, streamState.status])
  const retry = useCallback(async (message: ChatMessage) => {
    const index = messages.findIndex((item) => item.id === message.id)
    if (index < 0) return
    try { await truncateSession(activeSessionRef.current, index) } catch (error) {
      setStreamState({ ...initialChatState, status: 'error', error: normalizeClientError(error) })
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
      setStreamState({ ...initialChatState, status: 'error', error: normalizeClientError(error) })
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
    pumpControllerRef.current?.abort()
    const streamId = streamIdRef.current
    if (!streamId) {
      const pendingId = pendingUserIdRef.current
      if (pendingId) {
        const nextMessages = messagesRef.current.filter((message) => message.id !== pendingId)
        messagesRef.current = nextMessages
        setMessages(nextMessages)
        pendingUserIdRef.current = null
      }
      chatStateRef.current = initialChatState
      setStreamState(initialChatState)
      const next = queueRef.current.shift()
      setQueuedMessages(queueRef.current.map(({ content, attachmentNames }) => ({ content, attachmentNames })))
      if (next) pumpRef.current(next.content, next.files, messagesRef.current, next.options)
      return
    }
    await cancelChat(streamId).catch(() => undefined)
    finish(streamId, activeSessionRef.current, sessionEpochRef.current, { ...initialChatState, status: 'cancelled' }, 'cancelled')
  }, [finish])

  const removeQueued = useCallback((index: number) => {
    queueRef.current = queueRef.current.filter((_, itemIndex) => itemIndex !== index)
    setQueuedMessages(queueRef.current.map(({ content, attachmentNames }) => ({ content, attachmentNames })))
  }, [])
  const selectSession = useCallback(async (sessionId: string) => {
    const epoch = sessionEpochRef.current + 1
    sessionEpochRef.current = epoch; activeSessionRef.current = sessionId; closeSource(); pollControllerRef.current?.abort(); pumpControllerRef.current?.abort(); streamIdRef.current = null; fallbackStreamRef.current = null; setActiveSessionId(sessionId); chatStateRef.current = initialChatState; setStreamState(initialChatState); queueRef.current = []; setQueuedMessages([]); setSessionLoading(true); setSessionError(undefined)
    const controller = new AbortController()
    pollControllerRef.current?.abort(); pollControllerRef.current = controller
    try {
      const detail = await getSession(sessionId, controller.signal)
      if (sessionEpochRef.current !== epoch || activeSessionRef.current !== sessionId) return
      if (pollControllerRef.current === controller) pollControllerRef.current = null
      const restored = normalizeSessionMessages(detail.messages)
      setMessages(restored)
      turnBaselineRef.current = restored.length
      const journal = parseInflightTurn(window.localStorage.getItem(inflightTurnKey))
      if (journal?.session_id === sessionId) {
        streamIdRef.current = journal.stream_id
        setStreamState({ ...initialChatState, status: 'streaming' })
        connectStream(journal.stream_id, sessionId, epoch)
      }
    } catch (error) {
      if (sessionEpochRef.current !== epoch || activeSessionRef.current !== sessionId) return
      setMessages([]); setSessionError(normalizeRestoreError(error))
    } finally {
      if (pollControllerRef.current === controller) pollControllerRef.current = null
      if (sessionEpochRef.current === epoch) setSessionLoading(false)
    }
  }, [closeSource, connectStream])
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
  const reset = useCallback(() => { sessionEpochRef.current += 1; const nextSessionId = newId(); activeSessionRef.current = nextSessionId; closeSource(); pollControllerRef.current?.abort(); pumpControllerRef.current?.abort(); pollControllerRef.current = null; streamIdRef.current = null; fallbackStreamRef.current = null; window.localStorage.removeItem(inflightTurnKey); queueRef.current = []; setQueuedMessages([]); setMessages([]); chatStateRef.current = initialChatState; setStreamState(initialChatState); setDraft(''); setActiveSessionId(nextSessionId) }, [closeSource])

  return { messages, streamState, send, cancel, removeQueued, reset, retry, edit, approve, draft, setDraft, sessions, selectSession, searchSessions: searchSessionList, rename, pin, archive, remove, duplicate, activeSessionId, sessionLoading, sessionError, queuedMessages, isStreaming: streamState.status === 'streaming' }
}
