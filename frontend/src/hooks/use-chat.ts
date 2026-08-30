import { useCallback, useEffect, useRef, useState } from 'react'
import { cancelChat, startChat } from '../lib/api-client'
import { initialChatState, reduceChatEvent, type ChatEvent, type ChatEventType, type ChatMessage, type ChatState } from '../lib/chat-contract'

const supportedEvents: ChatEventType[] = ['token', 'reasoning', 'tool', 'tool_complete', 'done', 'cancel', 'apperror']

function id() {
  return crypto.randomUUID()
}

export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [streamState, setStreamState] = useState<ChatState>(initialChatState)
  const [sessionId] = useState(id)
  const sourceRef = useRef<EventSource | null>(null)
  const streamIdRef = useRef<string | null>(null)

  const closeSource = useCallback(() => {
    sourceRef.current?.close()
    sourceRef.current = null
  }, [])

  useEffect(() => closeSource, [closeSource])

  const send = useCallback(async (content: string) => {
    const clean = content.trim()
    if (!clean || streamState.status === 'streaming') return

    closeSource()
    setMessages((current) => [...current, { id: id(), role: 'user', content: clean, status: 'complete' }])
    setStreamState({ ...initialChatState, status: 'streaming' })

    try {
      const started = await startChat({ session_id: sessionId, message: clean })
      streamIdRef.current = started.stream_id
      const source = new EventSource(`/api/chat/stream?stream_id=${encodeURIComponent(started.stream_id)}`)
      sourceRef.current = source

      supportedEvents.forEach((type) => {
        source.addEventListener(type, (raw) => {
          const message = raw as MessageEvent<string>
          let data: Record<string, unknown> = {}
          try { data = JSON.parse(message.data) as Record<string, unknown> } catch { data = { message: message.data } }
          setStreamState((state) => {
            const next = reduceChatEvent(state, { type, data } as ChatEvent)
            if (type === 'done' && next.answer) {
              setMessages((current) => [...current, { id: id(), role: 'assistant', content: next.answer, status: 'complete' }])
              return initialChatState
            }
            return next
          })
          if (type === 'done' || type === 'cancel' || type === 'apperror') closeSource()
        })
      })

      source.onerror = () => {
        if (source.readyState === EventSource.CLOSED) {
          setStreamState((state) => state.status === 'streaming'
            ? { ...state, status: 'error', error: 'The Hermes stream closed unexpectedly.' }
            : state)
        }
      }
    } catch (error) {
      setStreamState({ ...initialChatState, status: 'error', error: error instanceof Error ? error.message : 'Unable to start Hermes.' })
    }
  }, [closeSource, sessionId, streamState.status])

  const cancel = useCallback(async () => {
    const streamId = streamIdRef.current
    if (!streamId) return
    await cancelChat(streamId).catch(() => undefined)
    closeSource()
    setStreamState((state) => ({ ...state, status: 'cancelled' }))
  }, [closeSource])

  const reset = useCallback(() => {
    closeSource()
    setMessages([])
    setStreamState(initialChatState)
  }, [closeSource])

  return { messages, streamState, send, cancel, reset, isStreaming: streamState.status === 'streaming' }
}
