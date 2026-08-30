import { useEffect, useRef, useState, type ChangeEvent, type KeyboardEvent } from 'react'
import { ArrowUp, FileText, Paperclip, Square, TerminalSquare, X } from 'lucide-react'
import { Button } from '../ui/button'

export function Composer({ onSend, onCancel, isStreaming, draft, onDraftChange, queuedMessages }: { onSend: (value: string, attachments?: File[]) => void; onCancel: () => void; isStreaming: boolean; draft?: string; onDraftChange?: (value: string) => void; queuedMessages?: string[] }) {
  const [localValue, setLocalValue] = useState('')
  const value = draft ?? localValue
  const setValue = onDraftChange ?? setLocalValue
  const [attachments, setAttachments] = useState<File[]>([])
  const ref = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    const input = ref.current
    if (!input) return
    input.style.height = 'auto'
    input.style.height = `${Math.min(input.scrollHeight, 180)}px`
  }, [value])

  function submit() {
    if (!value.trim()) return
    onSend(value, attachments)
    setValue('')
    setAttachments([])
  }

  function chooseFiles(event: ChangeEvent<HTMLInputElement>) {
    setAttachments((current) => [...current, ...Array.from(event.target.files || [])].slice(0, 5))
    event.target.value = ''
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
        {attachments.length > 0 && <div className="flex flex-wrap gap-2 px-2 pt-2" aria-label="Selected attachments">{attachments.map((file, index) => <span key={`${file.name}-${index}`} className="attachment-chip"><FileText size={13} />{file.name}<button type="button" onClick={() => setAttachments((current) => current.filter((_, item) => item !== index))} aria-label={`Remove ${file.name}`}><X size={13} /></button></span>)}</div>}
        <textarea ref={ref} value={value} onChange={(event) => setValue(event.target.value)} onKeyDown={onKeyDown} rows={1} placeholder={isStreaming ? 'Queue a message for Hermes…' : 'Message Hermes…'} aria-label="Message Hermes" className="max-h-44 min-h-12 w-full resize-none bg-transparent px-3 py-3 text-sm leading-6 outline-none placeholder:text-muted-foreground" />
        <div className="flex items-center gap-1 px-1 pb-1">
          <label className="inline-flex size-9 cursor-pointer items-center justify-center rounded-lg text-muted-foreground hover:bg-accent hover:text-accent-foreground focus-within:ring-2 focus-within:ring-ring" aria-label="Attach files"><Paperclip size={16} /><input type="file" multiple accept="image/png,image/jpeg,image/gif,image/webp,application/pdf,text/plain" className="sr-only" onChange={chooseFiles} /></label>
          <Button type="button" variant="ghost" size="sm" disabled className="gap-1.5"><TerminalSquare size={14} /> Tools</Button>
          <span className="ml-auto mr-2 hidden text-[11px] text-muted-foreground sm:inline">Enter to send · Shift Enter for newline</span>
          {isStreaming && <Button type="button" size="icon" variant="outline" onClick={onCancel} aria-label="Stop Hermes"><Square size={13} fill="currentColor" /></Button>}
          <Button type="button" size="icon" onClick={submit} disabled={!value.trim()} aria-label={isStreaming ? 'Queue message' : 'Send message'}><ArrowUp size={17} /></Button>
        </div>
      </div>
      <p className="mt-2 text-center text-[10px] text-muted-foreground">{queuedMessages?.length ? `${queuedMessages.length} message${queuedMessages.length === 1 ? '' : 's'} queued` : attachments.length ? `${attachments.length} attachment${attachments.length === 1 ? '' : 's'} ready` : 'Hermes can make mistakes and run tools. Review important actions.'}</p>
    </div>
  )
}
