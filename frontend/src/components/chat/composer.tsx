import { useEffect, useRef, useState, type ChangeEvent, type KeyboardEvent } from 'react'
import { ArrowUp, Bookmark, ChevronDown, Cpu, FileText, FolderOpen, Mic, Paperclip, Sparkles, Square, Timer, UserRound, X } from 'lucide-react'
import { Button } from '../ui/button'
import { Select } from '../ui/select'
import { Textarea } from '../ui/textarea'
import { normalizeTurnMode, type PendingTurn, type TurnMode } from '../../lib/turn-control'
import { localSlashCommand, slashCommandSuggestions } from '../../lib/slash-commands'
import { cn } from '../../lib/cn'
import { getModelCatalog, type ModelCatalogItem } from '../../lib/api-client'
import { findCatalogModel, groupModelCatalog, normalizeModelCatalog, searchModelCatalog } from '../../lib/model-catalog'
import { resolveComposerModel } from '../../lib/composer-state'

type Profile = { id: string; name: string; model: string; provider?: string }
type Props = { onSend: (value: string, attachments?: File[], options?: { model?: string; provider?: string }, mode?: TurnMode) => void; onCancel: () => void; onRemoveQueued?: (index: number) => void; onCommand?: (command: string) => void; isStreaming: boolean; draft?: string; onDraftChange?: (value: string) => void; queuedMessages?: PendingTurn[]; contextUsage?: { total?: number; contextLimit?: number }; workspacePath?: string; onWorkspaceOpen?: () => void }

const QUICK_PROMPTS = [
  { label: 'Inspect codebase', prompt: 'Inspect my current workspace and summarize its structure.' },
  { label: 'Plan implementation', prompt: 'Plan a focused implementation step by step.' },
  { label: 'Debug error', prompt: 'Help debug this error and propose a minimal fix:' },
  { label: 'Explain architecture', prompt: 'Explain the architecture and key contracts of this system.' },
]

export function Composer({ onSend, onCancel, onRemoveQueued, onCommand, isStreaming, draft, onDraftChange, queuedMessages, contextUsage, workspacePath = '.', onWorkspaceOpen }: Props) {
  const [localValue, setLocalValue] = useState('')
  const value = draft ?? localValue
  const setValue = onDraftChange ?? setLocalValue
  const [attachments, setAttachments] = useState<File[]>([])
  const ref = useRef<HTMLTextAreaElement>(null)
  const [commandsOpen, setCommandsOpen] = useState(false)
  const [promptsOpen, setPromptsOpen] = useState(false)
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [profileId, setProfileId] = useState('default')
  const [model, setModel] = useState('default')
  const [provider, setProvider] = useState('')
  const [turnMode, setTurnMode] = useState<TurnMode>('queue')
  const [catalog, setCatalog] = useState<ModelCatalogItem[]>([])
  const [catalogStatus, setCatalogStatus] = useState<'loading' | 'ready' | 'unavailable' | 'error'>('loading')
  const [modelSearch, setModelSearch] = useState('')
  const [contextOpen, setContextOpen] = useState(false)
  const activeProfile = profiles.find(profile => profile.id === profileId)
  const commandSuggestions = slashCommandSuggestions(value)

  useEffect(() => {
    const controller = new AbortController()
    void fetch('/api/profiles', { signal: controller.signal }).then(response => response.json()).then(data => {
      const next = (data.profiles || []) as Profile[]
      setProfiles(next)
      const active = typeof data.active === 'string' ? data.active : next[0]?.id || 'default'
      setProfileId(active)
      setModel(next.find(profile => profile.id === active)?.model || 'default')
      setProvider(next.find(profile => profile.id === active)?.provider || '')
    }).catch(() => undefined)
    void getModelCatalog(controller.signal).then(data => { setCatalog(data.models || []); setCatalogStatus(data.status) }).catch(() => setCatalogStatus('error'))
    return () => controller.abort()
  }, [])

  const normalizedCatalog = normalizeModelCatalog(catalog.map(item => ({ id: item.id, label: item.name, provider: item.provider || 'unknown', aliases: item.aliases || [], capabilities: [], available: item.available !== false })))
  const visibleModels = searchModelCatalog(normalizedCatalog, modelSearch)
  const modelGroups = groupModelCatalog(visibleModels)
  const staleProfileModel = resolveComposerModel({ model, provider }, normalizedCatalog, catalogStatus).stale

  useEffect(() => {
    const input = ref.current
    if (!input) return
    input.style.height = 'auto'
    input.style.height = `${Math.min(input.scrollHeight, 180)}px`
  }, [value])

  function submit() {
    if (!value.trim()) return
    if (staleProfileModel) return
    if (value.trim().startsWith('/') && onCommand && localSlashCommand(value.trim())) {
      onCommand(value.trim())
      setValue('')
      setCommandsOpen(false)
      return
    }
    onSend(value, attachments, { model, provider: provider || activeProfile?.provider }, turnMode)
    setValue('')
    setAttachments([])
  }

  function chooseFiles(event: ChangeEvent<HTMLInputElement>) {
    setAttachments((current) => [...current, ...Array.from(event.target.files || [])].slice(0, 5))
    event.target.value = ''
  }

  function onKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (commandsOpen && commandSuggestions.length > 0 && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
      event.preventDefault()
      const current = commandSuggestions.findIndex(command => command.name === value.trim().toLocaleLowerCase())
      const next = event.key === 'ArrowDown' ? (current + 1) % commandSuggestions.length : (current - 1 + commandSuggestions.length) % commandSuggestions.length
      setValue(`${commandSuggestions[next].name} `)
      return
    }
    if (commandsOpen && commandSuggestions.length > 0 && event.key === 'Tab') {
      event.preventDefault()
      setValue(`${commandSuggestions[0].name} `)
      return
    }
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      submit()
    }
  }

  const modelDisplay = model === 'default' ? (activeProfile?.model || 'Default model') : model
  const profileDisplay = activeProfile?.name || 'default'
  const workspaceDisplay = workspacePath === '.' ? 'Home' : workspacePath

  return (
    <div className="mobile-composer mx-auto w-full max-w-[1000px] px-3 pb-4 sm:px-6">
      <div className="relative rounded-2xl border border-border/70 bg-card/85 p-2 shadow-2xl shadow-black/40 backdrop-blur-xl transition-all focus-within:border-primary/40 focus-within:ring-1 focus-within:ring-primary/30">
        {attachments.length > 0 && (
          <div className="flex flex-wrap gap-1.5 px-2 pt-1 pb-2" aria-label="Selected attachments">
            {attachments.map((file, index) => (
              <span key={`${file.name}-${index}`} className="attachment-chip">
                <FileText size={12} className="text-primary" />
                <span className="max-w-40 truncate">{file.name}</span>
                <button type="button" onClick={() => setAttachments((current) => current.filter((_, item) => item !== index))} aria-label={`Remove ${file.name}`}>
                  <X size={11} />
                </button>
              </span>
            ))}
          </div>
        )}

        <Textarea
          ref={ref}
          value={value}
          onChange={(event) => {
            setValue(event.target.value)
            setCommandsOpen(event.target.value.startsWith('/'))
          }}
          onKeyDown={onKeyDown}
          rows={1}
          placeholder={isStreaming ? 'Queue a message for Hermes…' : 'Message Hermes…'}
          aria-label="Message Hermes"
          className="max-h-44 min-h-11 resize-none border-0 bg-transparent px-3 py-2 text-sm leading-6 placeholder:text-muted-foreground/60 shadow-none focus-visible:ring-0"
        />

        {commandsOpen && commandSuggestions.length > 0 && (
          <div className="grid gap-1 px-2 pb-2" aria-label="Slash commands" role="listbox">
            {commandSuggestions.map(command => (
              <Button key={command.name} type="button" variant="outline" size="sm" className="justify-start gap-2 text-left text-xs" role="option" onClick={() => { setValue(`${command.name} `); setCommandsOpen(false); ref.current?.focus() }}>
                <span className="font-mono text-primary">{command.name}</span>
                <span className="text-muted-foreground">{command.description}</span>
              </Button>
            ))}
          </div>
        )}

        {promptsOpen && (
          <div className="mb-2 grid grid-cols-1 gap-1.5 rounded-xl border border-border/80 bg-popover/90 p-2 text-xs sm:grid-cols-2">
            {QUICK_PROMPTS.map(item => (
              <button key={item.label} type="button" onClick={() => { setValue(item.prompt); setPromptsOpen(false); ref.current?.focus() }} className="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-left text-xs text-foreground/90 transition-colors hover:bg-accent hover:text-foreground">
                <Sparkles size={12} className="shrink-0 text-primary" />
                <span className="truncate">{item.label}</span>
              </button>
            ))}
          </div>
        )}

        <div className="composer-toolbar px-1 pt-1 pb-0.5">
          <div className="composer-toolbar__controls">
            {/* Attachment Button */}
            <Button type="button" variant="ghost" size="icon" className="size-7 rounded-lg text-muted-foreground hover:bg-accent/60 hover:text-foreground" aria-label="Attach files" title="Attach files" onClick={() => document.getElementById('composer-attachments')?.click()}>
              <Paperclip size={14} />
            </Button>
            <input id="composer-attachments" type="file" multiple accept="image/png,image/jpeg,image/gif,image/webp,application/pdf,text/plain" className="sr-only" onChange={chooseFiles} />

            {/* Voice Input */}
            <Button type="button" variant="ghost" size="icon" className="size-7 rounded-lg text-muted-foreground hover:bg-accent/60 hover:text-foreground" aria-label="Use voice input" title="Voice input" onClick={() => { const Speech = window.SpeechRecognition || window.webkitSpeechRecognition; if (!Speech) return; const recognition = new Speech(); recognition.onresult = (event: Event & { results: SpeechRecognitionResultList }) => setValue(`${value} ${event.results[0][0].transcript}`.trim()); recognition.start() }}>
              <Mic size={14} />
            </Button>

            {/* Prompts Bookmark */}
            <Button type="button" variant="ghost" size="icon" className={cn('size-7 rounded-lg transition-colors', promptsOpen ? 'bg-primary/20 text-primary' : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground')} aria-label="Prompt templates" title="Quick prompt templates" onClick={() => setPromptsOpen(v => !v)}>
              <Bookmark size={14} />
            </Button>

            {/* Profile Pill */}
            <div className="relative inline-flex items-center">
              <Select
                aria-label="Conversation profile"
                value={profileId}
                onChange={event => {
                  const next = profiles.find(p => p.id === event.target.value)
                  setProfileId(event.target.value)
                  setModel(next?.model || 'default')
                  setProvider(next?.provider || '')
                  void fetch('/api/profiles/active', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: event.target.value }) })
                }}
                className="select-menu-up h-7 min-h-7 max-w-28 rounded-full border-border/60 bg-muted/40 px-2 text-[11px] font-medium hover:border-border hover:bg-muted/70"
              >
                <option value="default">Default</option>
                {profiles.filter(p => p.id !== 'default').map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
              </Select>
            </div>

            {/* Workspace / Directory Pill */}
            <button type="button" onClick={onWorkspaceOpen} className="flex h-7 items-center gap-1 rounded-full border border-border/60 bg-muted/40 px-2 text-[11px] font-medium text-foreground/90 transition-colors hover:border-border hover:bg-muted/70" aria-label={`Workspace ${workspacePath}`} title={`Workspace ${workspacePath}`}>
              <FolderOpen size={11} className="text-muted-foreground" />
              <span className="max-w-24 truncate">{workspaceDisplay}</span>
              <ChevronDown size={10} className="text-muted-foreground" />
            </button>

            {/* Model Pill */}
            <div className="relative inline-flex items-center">
              <Select
                aria-label="Conversation model"
                value={model}
                onChange={event => { const selected = findCatalogModel(normalizedCatalog, event.target.value); setModel(selected?.id || 'default'); setProvider(selected?.provider || '') }}
                className="select-menu-up h-7 min-h-7 max-w-36 rounded-full border-border/60 bg-muted/40 px-2 text-[11px] font-medium hover:border-border hover:bg-muted/70"
              >
                <option value="default">{catalogStatus === 'loading' ? 'Loading models…' : catalogStatus === 'error' ? 'Models unavailable' : catalogStatus === 'unavailable' ? 'Catalog unavailable' : 'Default model'}</option>
                {activeProfile?.model && activeProfile.model !== 'default' && !findCatalogModel(normalizedCatalog, activeProfile.model, activeProfile.provider) && <option value={activeProfile.model}>{activeProfile.model} (unavailable)</option>}
                {modelGroups.map(group => <optgroup key={group.provider} label={group.provider}>{group.models.map(item => <option key={`${group.provider}:${item.id}`} value={item.id}>{item.label}</option>)}</optgroup>)}
              </Select>
            </div>
            {catalogStatus === 'ready' && catalog.length > 0 && <input aria-label="Search models" value={modelSearch} onChange={event => setModelSearch(event.target.value)} placeholder="Search models" className="h-7 w-28 rounded-full border border-border/60 bg-muted/40 px-2 text-[11px]" />}
            {catalogStatus === 'ready' && catalog.length > 0 && visibleModels.length === 0 && <span role="status" className="text-[10px] text-muted-foreground">No matching models</span>}
            {catalogStatus === 'loading' && <span role="status" className="text-[10px] text-muted-foreground">Loading model catalog…</span>}
            {catalogStatus === 'ready' && catalog.length === 0 && <span className="text-[10px] text-muted-foreground">No models available</span>}
            {catalogStatus === 'unavailable' && <span className="text-[10px] text-muted-foreground">Gateway catalog unavailable</span>}
            {catalogStatus === 'error' && <span role="status" className="text-[10px] text-destructive">Unable to load model catalog</span>}
            {staleProfileModel && <span role="alert" className="text-[10px] text-destructive">Selected profile model unavailable; choose a valid model</span>}

            {/* Reasoning Effort / Turn Mode Pill */}
            <div className="relative inline-flex items-center">
              <Select
                aria-label="Turn mode"
                value={turnMode}
                onChange={event => setTurnMode(normalizeTurnMode(event.target.value))}
                className="select-menu-up h-7 min-h-7 max-w-24 rounded-full border-border/60 bg-muted/40 px-2 text-[11px] font-medium hover:border-border hover:bg-muted/70"
              >
                <option value="queue">High</option>
                <option value="interrupt">Interrupt</option>
                <option value="steer">Steer</option>
              </Select>
            </div>

            {/* Context Token Window Pill */}
            {contextUsage?.total !== undefined && contextUsage.contextLimit !== undefined && (
              <div className="relative">
                <button type="button" className="inline-flex size-6 items-center justify-center rounded-full border border-border/60 bg-muted/60 text-[9px] font-semibold tabular-nums text-muted-foreground hover:border-primary/60 hover:text-foreground" aria-label="Context window usage" aria-expanded={contextOpen} onClick={() => setContextOpen(value => !value)}>
                  {Math.min(100, Math.round((contextUsage.total / Math.max(1, contextUsage.contextLimit)) * 100))}
                </button>
                {contextOpen && (
                  <div className="absolute bottom-[calc(100%+0.5rem)] right-0 z-[200] w-64 rounded-xl border border-border/80 bg-popover/95 p-3 text-left text-xs text-muted-foreground shadow-2xl shadow-black/50 backdrop-blur-xl animate-in fade-in-0 zoom-in-95">
                    <p className="text-sm font-semibold text-foreground">Context window</p>
                    <p className="mt-2">Context window: {Math.round((contextUsage.total / Math.max(1, contextUsage.contextLimit)) * 100)}% used</p>
                    <p className="mt-1">{contextUsage.total.toLocaleString()} / {contextUsage.contextLimit.toLocaleString()} tokens</p>
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="composer-toolbar__actions">
            <span className="mr-1 hidden text-[10px] text-muted-foreground/60 sm:inline">Enter to send · Shift Enter for newline</span>
            {isStreaming && (
              <Button type="button" size="icon" variant="outline" className="size-8 rounded-full border-border/80" onClick={onCancel} aria-label="Stop Hermes">
                <Square size={12} fill="currentColor" />
              </Button>
            )}
            <Button
              type="button"
              size="icon"
              onClick={submit}
              disabled={!value.trim() || staleProfileModel}
              aria-label={isStreaming ? 'Queue message' : 'Send message'}
              className="size-8 rounded-full bg-primary/90 text-primary-foreground transition-transform hover:scale-105 active:scale-95 disabled:opacity-30"
            >
              <ArrowUp size={15} />
            </Button>
          </div>
        </div>
      </div>

      <p className="mt-1.5 text-center text-[10px] text-muted-foreground/60">
        {queuedMessages?.length ? `${queuedMessages.length} message${queuedMessages.length === 1 ? '' : 's'} queued` : attachments.length ? `${attachments.length} attachment${attachments.length === 1 ? '' : 's'} ready` : 'Hermes can make mistakes and run tools. Review important actions.'}
      </p>

      {queuedMessages?.length ? (
        <div className="mt-1.5 space-y-1" aria-label="Pending messages">
          {queuedMessages.map((item, index) => (
            <div key={`${item.content}-${index}`} className="flex items-center gap-2 rounded-lg border bg-card/50 px-2.5 py-1 text-left text-[11px]">
              <span className="min-w-0 flex-1 truncate">{item.content}{item.attachmentNames.length ? ` · ${item.attachmentNames.join(', ')}` : ''}</span>
              {onRemoveQueued && (
                <Button type="button" variant="ghost" size="icon" className="size-5 shrink-0 text-muted-foreground hover:text-foreground" onClick={() => onRemoveQueued(index)} aria-label={`Remove queued message ${index + 1}`}>
                  <X size={11} />
                </Button>
              )}
            </div>
          ))}
        </div>
      ) : null}
    </div>
  )
}
