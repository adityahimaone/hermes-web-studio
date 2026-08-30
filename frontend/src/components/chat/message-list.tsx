import { Bot, LoaderCircle } from 'lucide-react'
import type { ChatMessage, ChatState } from '../../lib/chat-contract'
import { SafeMarkdown } from '../../lib/markdown'
import { ActivityCards } from './activity-cards'

function EmptyState() {
  return (
    <div className="mx-auto flex min-h-[58vh] max-w-xl flex-col items-center justify-center px-6 text-center">
      <div className="relative grid size-16 place-items-center rounded-2xl border bg-card shadow-2xl shadow-primary/10">
        <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl" />
        <Bot className="relative text-primary" size={28} />
      </div>
      <h1 className="mt-6 text-2xl font-semibold tracking-tight sm:text-3xl">What are we building today?</h1>
      <p className="mt-3 max-w-md text-sm leading-6 text-muted-foreground">Talk to Hermes through the new Gateway-first stack. Streaming, tool activity, cancellation, and failures stay visible in one focused workspace.</p>
      <div className="mt-6 grid w-full gap-2 sm:grid-cols-2">
        {['Inspect my current project', 'Plan a focused implementation', 'Help debug an error', 'Explain this codebase'].map((prompt) => (
          <p key={prompt} className="rounded-xl border bg-card/45 px-4 py-3 text-left text-xs text-muted-foreground">{prompt}</p>
        ))}
      </div>
    </div>
  )
}

export function MessageList({ messages, stream }: { messages: ChatMessage[]; stream: ChatState }) {
  if (!messages.length && stream.status === 'idle') return <EmptyState />

  return (
    <div className="mx-auto w-full max-w-3xl space-y-8 px-4 py-8 sm:px-8">
      {messages.map((message) => (
        message.role === 'user' ? (
          <article key={message.id} className="flex justify-end">
            <div className="max-w-[85%] rounded-2xl rounded-br-md bg-secondary px-4 py-3 text-sm leading-6 text-secondary-foreground shadow-sm">{message.content}</div>
          </article>
        ) : (
          <article key={message.id} className="flex gap-3">
            <div className="mt-0.5 grid size-8 shrink-0 place-items-center rounded-xl border bg-card text-primary"><Bot size={16} /></div>
            <div className="message-markdown min-w-0 flex-1 pt-1 text-sm leading-7 text-foreground/95"><SafeMarkdown>{message.content}</SafeMarkdown></div>
          </article>
        )
      ))}

      {stream.status !== 'idle' && (
        <article className="flex gap-3">
          <div className="mt-0.5 grid size-8 shrink-0 place-items-center rounded-xl border bg-card text-primary"><Bot size={16} /></div>
          <div className="min-w-0 flex-1 pt-1">
            {stream.reasoning && (
              <details className="mb-3 rounded-xl border bg-card/35 px-3 py-2 text-xs text-muted-foreground">
                <summary className="cursor-pointer font-medium text-foreground/80">Hermes reasoning</summary>
                <p className="mt-2 whitespace-pre-wrap leading-5">{stream.reasoning}</p>
              </details>
            )}
            <ActivityCards tools={stream.tools} subagents={stream.subagents} approvals={stream.approvals} />
            {stream.error ? (
              <div className="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-red-200">{stream.error}</div>
            ) : stream.answer ? (
              <div className="message-markdown text-sm leading-7 text-foreground/95"><SafeMarkdown>{stream.answer}</SafeMarkdown></div>
            ) : stream.status === 'cancelled' ? (
              <p className="text-sm text-muted-foreground">Generation stopped.</p>
            ) : (
              <div role="status" className="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle size={15} className="animate-spin text-primary" /> Hermes is thinking…</div>
            )}
            {stream.usage && <div className="mt-3 text-xs text-muted-foreground">{stream.usage.total !== undefined ? `${stream.usage.total.toLocaleString()} tokens` : 'Token usage reported'}{stream.usage.contextLimit ? ` · ${Math.round((stream.usage.total || 0) / stream.usage.contextLimit * 100)}% context` : ''}</div>}
          </div>
        </article>
      )}
    </div>
  )
}
