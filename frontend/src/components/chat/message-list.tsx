import { Bot, ChevronDown, Copy, ArrowDown, LoaderCircle, Pencil, RotateCcw, Sparkles } from 'lucide-react'
import { useState } from 'react'
import type { ChatMessage, ChatState } from '../../lib/chat-contract'
import { SafeMarkdown } from '../../lib/markdown'
import { ActivityCards } from './activity-cards'
import type { ApprovalChoice } from '../../lib/api-client'
import { MermaidDiagram, splitMermaidBlocks } from './mermaid'
import { Button } from '../ui/button'

function RichContent({ content }: { content: string }) {
  return <>{splitMermaidBlocks(content).map((part, index) => part.kind === 'mermaid' ? <MermaidDiagram key={index} source={part.content} /> : <SafeMarkdown key={index}>{part.content}</SafeMarkdown>)}</>
}

function EmptyState() {
  return (
    <div className="mx-auto flex min-h-[52vh] max-w-xl flex-col items-center justify-center px-6 text-center">
      <div className="relative grid size-14 place-items-center rounded-2xl border border-border/80 bg-card shadow-2xl shadow-primary/10">
        <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl" />
        <Bot className="relative text-primary" size={24} />
      </div>
      <h1 className="mt-5 text-xl font-semibold tracking-tight sm:text-2xl">What are we building today?</h1>
      <p className="mt-2.5 max-w-md text-xs leading-5 text-muted-foreground">Talk to Hermes through the new Gateway-first stack. Streaming, tool activity, cancellation, and failures stay visible in one focused workspace.</p>
      <div className="mt-5 grid w-full gap-2 sm:grid-cols-2">
        {['Inspect my current project', 'Plan a focused implementation', 'Help debug an error', 'Explain this codebase'].map((prompt) => (
          <p key={prompt} className="rounded-xl border border-border/60 bg-card/40 px-3.5 py-2.5 text-left text-xs text-muted-foreground/80 transition-colors hover:border-border hover:text-foreground">{prompt}</p>
        ))}
      </div>
    </div>
  )
}

function ThinkingBlock({ reasoning }: { reasoning: string }) {
  const [open, setOpen] = useState(false)
  const [copied, setCopied] = useState(false)
  const snippet = reasoning.slice(0, 70).replace(/\n/g, ' ')

  const copyReasoning = (e: React.MouseEvent) => {
    e.stopPropagation()
    void navigator.clipboard.writeText(reasoning)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="thinking-card mb-3 text-xs">
      <div
        role="button"
        tabIndex={0}
        onClick={() => setOpen(v => !v)}
        onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') setOpen(v => !v) }}
        className="flex cursor-pointer items-center gap-2 px-3 py-2 text-muted-foreground hover:text-foreground"
      >
        <Sparkles size={13} className="shrink-0 text-primary" />
        <span className="font-semibold text-foreground/90">Thinking</span>
        <span className="min-w-0 flex-1 truncate text-[11px] text-muted-foreground/70">{snippet}...</span>
        <button type="button" onClick={copyReasoning} className="rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground" title={copied ? 'Copied' : 'Copy thinking'}>
          <Copy size={12} />
        </button>
        <ChevronDown size={13} className={`shrink-0 transition-transform ${open ? 'rotate-180' : ''}`} />
      </div>
      {open && (
        <div className="border-t border-border/40 px-3 py-2.5 text-[11px] leading-5 text-muted-foreground/90 whitespace-pre-wrap">
          {reasoning}
        </div>
      )}
    </div>
  )
}

export function MessageList({ messages, stream, onEdit, onRetry, onApproval }: { messages: ChatMessage[]; stream: ChatState; onEdit?: (message: ChatMessage) => void; onRetry?: (message: ChatMessage) => void; onApproval?: (id: string, choice: ApprovalChoice) => void }) {
  if (!messages.length && stream.status === 'idle') return <EmptyState />

  return (
    <div className="relative mx-auto w-full max-w-[900px] space-y-6 px-4 py-6 sm:px-6">
      {messages.map((message) => (
        message.role === 'user' ? (
          <article key={message.id} className="group flex justify-end gap-2">
            {onEdit && <Button type="button" variant="ghost" size="icon" className="message-action size-7 rounded-lg" onClick={() => onEdit(message)} aria-label="Edit message"><Pencil size={13} /></Button>}
            <div className="max-w-[75%] rounded-2xl rounded-br-md border border-border/50 bg-secondary/80 px-4 py-2.5 text-[14px] leading-6 text-foreground shadow-sm">{message.content}</div>
          </article>
        ) : (
          <article key={message.id} className="flex flex-col gap-2">
            {/* Assistant Header Badge */}
            <div className="flex items-center gap-2">
              <div className="grid size-6 shrink-0 place-items-center rounded-lg border border-border/80 bg-card text-primary"><Bot size={13} /></div>
              <span className="text-xs font-semibold tracking-tight text-foreground/90">Hermes</span>
              <span className="rounded-full bg-muted/60 px-1.5 py-0.2 text-[9px] font-medium text-muted-foreground">66.7 t/s</span>
            </div>
            <div className="message-markdown min-w-0 pl-8 text-[14px] leading-7 text-foreground/95">
              <RichContent content={message.content} />
              {onRetry && (() => { const previous = [...messages].reverse().find((item: ChatMessage) => item.role === 'user'); return previous ? <Button type="button" variant="ghost" size="sm" className="message-action mt-2 h-7 gap-1 px-2 text-xs" onClick={() => onRetry(previous)}><RotateCcw size={12} /> Retry</Button> : null })()}
            </div>
          </article>
        )
      ))}

      {stream.status !== 'idle' && (
        <article className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <div className="grid size-6 shrink-0 place-items-center rounded-lg border border-border/80 bg-card text-primary"><Bot size={13} /></div>
            <span className="text-xs font-semibold tracking-tight text-foreground/90">Hermes</span>
          </div>
          <div className="min-w-0 pl-8">
            {stream.reasoning && <ThinkingBlock reasoning={stream.reasoning} />}
            <ActivityCards tools={stream.tools} subagents={stream.subagents} approvals={stream.approvals} onApproval={onApproval} />
            {stream.error ? (
              <div className="rounded-xl border border-destructive/30 bg-destructive/10 px-3.5 py-2.5 text-xs text-red-200">{stream.error}</div>
            ) : stream.answer ? (
              <div className="message-markdown text-xs leading-6 text-foreground/95"><RichContent content={stream.answer} /></div>
            ) : stream.status === 'cancelled' ? (
              <p className="text-xs text-muted-foreground">Generation stopped.</p>
            ) : (
              <div role="status" className="flex items-center gap-2 text-xs text-muted-foreground"><LoaderCircle size={14} className="animate-spin text-primary" /> Hermes is thinking…</div>
            )}
            {stream.usage && <div className="mt-2 text-[10px] text-muted-foreground">{stream.usage.total !== undefined ? `${stream.usage.total.toLocaleString()} tokens` : 'Token usage reported'}{stream.usage.contextLimit ? ` · ${Math.round((stream.usage.total || 0) / stream.usage.contextLimit * 100)}% context` : ''}</div>}
          </div>
        </article>
      )}
    </div>
  )
}
