import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { ArrowUp, Paperclip, Square, TerminalSquare } from 'lucide-react'
import { Button } from '../ui/button'

export function Composer({ onSend, onCancel, isStreaming }: { onSend: (value: string) => void; onCancel: () => void; isStreaming: boolean }) {
  const [value, setValue] = useState('')
  const ref = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    const input = ref.current
    if (!input) return
    input.style.height = 'auto'
    input.style.height = `${Math.min(input.scrollHeight, 180)}px`
  }, [value])

  function submit() {
    if (!value.trim() || isStreaming) return
    onSend(value)
    setValue('')
  }

  function onKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      submit()
    }
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-3 pb-4 sm:px-6">
      <div className="rounded-2xl border bg-card/90 p-2 shadow-2xl shadow-black/30 backdrop-blur-xl focus-within:border-primary/35 focus-within:ring-1 focus-within:ring-primary/20">
        <textarea ref={ref} value={value} onChange={(event) => setValue(event.target.value)} onKeyDown={onKeyDown} disabled={isStreaming} rows={1} placeholder="Message Hermes…" aria-label="Message Hermes" className="max-h-44 min-h-12 w-full resize-none bg-transparent px-3 py-3 text-sm leading-6 outline-none placeholder:text-muted-foreground disabled:opacity-60" />
        <div className="flex items-center gap-1 px-1 pb-1">
          <Button type="button" variant="ghost" size="icon" aria-label="Attach file (planned)" disabled><Paperclip size={16} /></Button>
          <Button type="button" variant="ghost" size="sm" disabled className="gap-1.5"><TerminalSquare size={14} /> Tools</Button>
          <span className="ml-auto mr-2 hidden text-[11px] text-muted-foreground sm:inline">Enter to send · Shift Enter for newline</span>
          {isStreaming ? (
            <Button type="button" size="icon" variant="outline" onClick={onCancel} aria-label="Stop Hermes"><Square size={13} fill="currentColor" /></Button>
          ) : (
            <Button type="button" size="icon" onClick={submit} disabled={!value.trim()} aria-label="Send message"><ArrowUp size={17} /></Button>
          )}
        </div>
      </div>
      <p className="mt-2 text-center text-[10px] text-muted-foreground">Hermes can make mistakes and run tools. Review important actions.</p>
    </div>
  )
}

