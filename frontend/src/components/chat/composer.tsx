import { useEffect, useRef, useState, type ChangeEvent, type KeyboardEvent } from 'react'
import { ArrowUp, Cpu, FileText, FolderOpen, Mic, Paperclip, Square, TerminalSquare, UserRound, X } from 'lucide-react'
import { Button } from '../ui/button'
import { Textarea } from '../ui/textarea'
import { Select } from '../ui/select'
import { Badge } from '../ui/badge'

type Profile = { id: string; name: string; model: string; provider?: string }
type Props = { onSend: (value: string, attachments?: File[], options?: { model?: string; provider?: string }) => void; onCancel: () => void; onCommand?: (command: string) => void; isStreaming: boolean; draft?: string; onDraftChange?: (value: string) => void; queuedMessages?: string[]; contextUsage?: { total?: number; contextLimit?: number }; workspacePath?: string; onWorkspaceOpen?: () => void }
export function Composer({ onSend, onCancel, onCommand, isStreaming, draft, onDraftChange, queuedMessages, contextUsage, workspacePath = '.', onWorkspaceOpen }: Props) {
  const [localValue, setLocalValue] = useState('')
  const value = draft ?? localValue
  const setValue = onDraftChange ?? setLocalValue
  const [attachments, setAttachments] = useState<File[]>([])
  const ref = useRef<HTMLTextAreaElement>(null)
  const [commandsOpen, setCommandsOpen] = useState(false)
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [profileId, setProfileId] = useState('default')
  const [model, setModel] = useState('default')
  const activeProfile = profiles.find(profile => profile.id === profileId)
  useEffect(() => { void fetch('/api/profiles').then(response => response.json()).then(data => { const next = (data.profiles || []) as Profile[]; setProfiles(next); const active = typeof data.active === 'string' ? data.active : next[0]?.id || 'default'; setProfileId(active); setModel(next.find(profile => profile.id === active)?.model || 'default') }).catch(() => undefined) }, [])

  useEffect(() => {
    const input = ref.current
    if (!input) return
    input.style.height = 'auto'
    input.style.height = `${Math.min(input.scrollHeight, 180)}px`
  }, [value])

  function submit() {
    if (!value.trim()) return
    if (value.trim().startsWith('/') && onCommand) { onCommand(value.trim()); setValue(''); setCommandsOpen(false); return }
    onSend(value, attachments, { model, provider: activeProfile?.provider })
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
    <div className="mobile-composer mx-auto w-full max-w-3xl px-3 pb-4 sm:px-6">
      <div className="rounded-2xl border bg-card/90 p-2 shadow-2xl shadow-black/30 backdrop-blur-xl focus-within:border-primary/35 focus-within:ring-1 focus-within:ring-primary/20">
        {attachments.length > 0 && <div className="flex flex-wrap gap-2 px-2 pt-2" aria-label="Selected attachments">{attachments.map((file, index) => <span key={`${file.name}-${index}`} className="attachment-chip"><FileText size={13} />{file.name}<Button type="button" variant="ghost" size="icon" className="size-5" onClick={() => setAttachments((current) => current.filter((_, item) => item !== index))} aria-label={`Remove ${file.name}`}><X size={13} /></Button></span>)}</div>}
        <Textarea ref={ref} value={value} onChange={(event) => { setValue(event.target.value); setCommandsOpen(event.target.value.startsWith('/')) }} onKeyDown={onKeyDown} rows={1} placeholder={isStreaming ? 'Queue a message for Hermes…' : 'Message Hermes…'} aria-label="Message Hermes" className="max-h-44 min-h-12 resize-none border-0 bg-transparent px-3 py-3 leading-6 shadow-none focus-visible:ring-0" />
        {commandsOpen && <div className="flex flex-wrap gap-1 px-2 pb-2" aria-label="Slash commands"><Button type="button" variant="outline" size="sm" onClick={() => { setValue('/help'); setCommandsOpen(false) }}>/help</Button><Button type="button" variant="outline" size="sm" onClick={() => { setValue('/clear'); setCommandsOpen(false) }}>/clear</Button></div>}
        <div className="composer-toolbar px-1 pb-1">
          <div className="composer-toolbar__controls">
          <Button type="button" variant="ghost" size="icon" aria-label="Attach files" onClick={() => document.getElementById('composer-attachments')?.click()}><Paperclip size={16} /></Button><input id="composer-attachments" type="file" multiple accept="image/png,image/jpeg,image/gif,image/webp,application/pdf,text/plain" className="sr-only" onChange={chooseFiles} />
          <Button type="button" variant="ghost" size="sm" disabled className="gap-1.5"><TerminalSquare size={14} /> Tools</Button>
          <Button type="button" variant="ghost" size="icon" aria-label="Use voice input" onClick={() => { const Speech = window.SpeechRecognition || window.webkitSpeechRecognition; if (!Speech) return; const recognition = new Speech(); recognition.onresult = (event: Event & { results: SpeechRecognitionResultList }) => setValue(`${value} ${event.results[0][0].transcript}`.trim()); recognition.start() }}><Mic size={15} /></Button>
          <span className="hidden h-5 w-px bg-border sm:block" aria-hidden="true" />
          <label className="sr-only" htmlFor="composer-profile">Profile</label><span className="hidden items-center gap-1 text-[10px] text-muted-foreground sm:flex"><UserRound size={12} /></span><Select id="composer-profile" aria-label="Conversation profile" value={profileId} onChange={event => { const next = profiles.find(profile => profile.id === event.target.value); setProfileId(event.target.value); setModel(next?.model || 'default'); void fetch('/api/profiles/active', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: event.target.value }) }) }} className="h-8 max-w-28 px-2 text-[11px]"><option value="default">Default</option>{profiles.filter(profile => profile.id !== 'default').map(profile => <option key={profile.id} value={profile.id}>{profile.name}</option>)}</Select>
          <label className="sr-only" htmlFor="composer-model">Model</label><span className="hidden items-center gap-1 text-[10px] text-muted-foreground sm:flex"><Cpu size={12} /></span><Select id="composer-model" aria-label="Conversation model" value={model} onChange={event => setModel(event.target.value)} className="h-8 max-w-32 px-2 text-[11px]"><option value="default">Default model</option>{activeProfile?.model && activeProfile.model !== 'default' && <option value={activeProfile.model}>{activeProfile.model}</option>}</Select>
          <Button type="button" variant="outline" size="sm" className="h-8 max-w-28 gap-1 px-2 text-[11px]" onClick={onWorkspaceOpen} aria-label={`Workspace ${workspacePath}`}><FolderOpen size={12} /><span className="truncate">{workspacePath}</span></Button>
          {contextUsage?.total !== undefined && contextUsage.contextLimit !== undefined && <Badge className="h-7 gap-1 px-2 text-[10px]" title="Context usage"><span className="inline-block size-2 rounded-full bg-primary" />{Math.min(100, Math.round((contextUsage.total / Math.max(1, contextUsage.contextLimit)) * 100))}%</Badge>}
          </div>
          <div className="composer-toolbar__actions">
          <span className="ml-auto mr-2 hidden text-[11px] text-muted-foreground sm:inline">Enter to send · Shift Enter for newline</span>
          {isStreaming && <Button type="button" size="icon" variant="outline" onClick={onCancel} aria-label="Stop Hermes"><Square size={13} fill="currentColor" /></Button>}
          <Button type="button" size="icon" onClick={submit} disabled={!value.trim()} aria-label={isStreaming ? 'Queue message' : 'Send message'}><ArrowUp size={17} /></Button>
          </div>
        </div>
      </div>
      <p className="mt-2 text-center text-[10px] text-muted-foreground">{queuedMessages?.length ? `${queuedMessages.length} message${queuedMessages.length === 1 ? '' : 's'} queued` : attachments.length ? `${attachments.length} attachment${attachments.length === 1 ? '' : 's'} ready` : 'Hermes can make mistakes and run tools. Review important actions.'}</p>
    </div>
  )
}
