import { Archive, ArchiveX, Check, Copy, Download, Pencil, Pin, Plus, Search, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Dialog } from '../ui/dialog'
import { cn } from '../../lib/cn'
import { filterSessions, groupSessionsByDate, type SessionSummary } from '../../lib/chat-contract'
import { projectSessions } from '../../lib/conversation-runtime'
import { sessionExportUrl } from '../../lib/api-client'
import { DropdownMenu } from '../ui/dropdown-menu'
import { ContextRail } from '../layout/context-rail'

type Props = { sessions: SessionSummary[]; activeSessionId: string; onSelectSession: (id: string) => void; onSearch?: (query: string) => Promise<unknown>; onRename: (id: string, title: string) => Promise<void>; onPin: (id: string, pinned: boolean) => void; onArchive: (id: string, archived: boolean) => void; onDelete: (id: string) => Promise<void>; onDuplicate?: (id: string) => Promise<void>; onNewChat: () => void; loading: boolean; error?: string; onToggle: () => void }
type SourceFilter = 'all' | 'webui' | 'cli'

function field(session: SessionSummary, ...keys: string[]) { for (const key of keys) { const value = session[key]; if (typeof value === 'string' && value.trim()) return value.trim() } return '' }
function sourceOf(session: SessionSummary): SourceFilter { const source = field(session, 'source', 'session_source', 'origin', 'session_type').toLocaleLowerCase(); return source.includes('cli') || source.includes('cron') ? 'cli' : 'webui' }
function channelOf(session: SessionSummary) { return field(session, 'channel', 'channel_name', 'external_channel', 'transport') }
export function filterAriaPressed(selected: string, value: string) { return selected === value }
export function sessionRowAriaCurrent(sessionId: string, activeSessionId: string) { return sessionId === activeSessionId ? 'page' : undefined }
export function nextSessionId(ids: string[], currentId: string, direction: 'next' | 'previous') {
  if (!ids.length) return undefined
  const currentIndex = ids.indexOf(currentId)
  const offset = direction === 'next' ? 1 : -1
  return ids[(currentIndex < 0 ? 0 : currentIndex + offset + ids.length) % ids.length]
}
export function findSessionButton(root: ParentNode, sessionId: string) {
  return Array.from(root.querySelectorAll<HTMLButtonElement>('button[data-session-id]')).find(button => button.dataset.sessionId === sessionId)
}
export function focusSessionButton(root: ParentNode, sessionId: string) {
  const button = findSessionButton(root, sessionId)
  if (!button) return false
  button.focus()
  return true
}
export function focusSessionButtonAfterSelection(root: ParentNode, sessionId: string, schedule = window.requestAnimationFrame, maxFrames = 10) {
  let frames = 0
  const attempt = () => {
    const button = findSessionButton(root, sessionId)
    if (button?.getAttribute('aria-current') === 'page') {
      button.focus()
      return
    }
    if (frames++ < maxFrames) schedule(attempt)
  }
  schedule(attempt)
}
export function sessionActionVisibilityClass() { return 'absolute right-1 top-1/2 flex -translate-y-1/2 rounded-md bg-card/95 shadow-sm transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' }

function formatCompactDate(dateStr?: string) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ''
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  if (diffHours < 24 && diffHours >= 0) return diffHours === 0 ? 'now' : `${diffHours}h`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays}d`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)}w`
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

export function SessionRail({ sessions, activeSessionId, onSelectSession, onSearch, onRename, onPin, onArchive, onDelete, onDuplicate, onNewChat, loading, error, onToggle }: Props) {
  const [query, setQuery] = useState(''); const [sourceFilter, setSourceFilter] = useState<SourceFilter>('all'); const [tagFilter, setTagFilter] = useState('all'); const [selected, setSelected] = useState<string[]>([]); const [batchMode, setBatchMode] = useState(false); const [renameTarget, setRenameTarget] = useState<SessionSummary | null>(null); const [deleteTarget, setDeleteTarget] = useState<SessionSummary | null>(null); const [renameValue, setRenameValue] = useState('')
  const [showArchived, setShowArchived] = useState(false)
  
  const webCount = useMemo(() => sessions.filter(item => sourceOf(item) === 'webui').length, [sessions])
  const cliCount = useMemo(() => sessions.filter(item => sourceOf(item) === 'cli').length, [sessions])
  
  const allTags = useMemo(() => {
    const tags = new Set<string>()
    for (const session of sessions) {
      if (session.project) tags.add(session.project)
      if (session.tags) {
        for (const t of session.tags) if (t) tags.add(t)
      }
    }
    return Array.from(tags).sort()
  }, [sessions])

  const visible = filterSessions(sessions, query).filter(item => {
    if (!showArchived && item.archived) return false
    if (sourceFilter !== 'all' && sourceOf(item) !== sourceFilter) return false
    if (tagFilter === 'unassigned') return !item.project && (!item.tags || item.tags.length === 0)
    if (tagFilter !== 'all' && item.project !== tagFilter && !item.tags?.includes(tagFilter)) return false
    return true
  })
  
  const groups = groupSessionsByDate(visible)
  const visibleSessionIds = groups.flatMap(group => group.sessions.map(item => item.session_id))
  const activeProjection = useMemo(() => projectSessions(sessions).find(item => item.session_id === activeSessionId), [sessions, activeSessionId])
  useEffect(() => { if (!onSearch) return; const timer = window.setTimeout(() => { void onSearch(query) }, 250); return () => window.clearTimeout(timer) }, [onSearch, query])
  const toggleSelected = (id: string) => setSelected(current => current.includes(id) ? current.filter(item => item !== id) : [...current, id])
  const clearBatch = () => { setSelected([]); setBatchMode(false) }

  return <>
    <ContextRail title="Chat" subtitle="Recent sessions" open onToggle={onToggle} action={<Button type="button" variant="ghost" size="icon" className="size-7 rounded-lg text-muted-foreground hover:bg-accent/60 hover:text-foreground" onClick={onNewChat} aria-label="New chat" title="New chat"><Plus size={16} /></Button>}>
      {/* Sticky Top Filter Area */}
      <div className="sticky top-0 z-10 -mx-3 -mt-2 border-b border-border/40 bg-card/95 px-3 pb-2 pt-2 backdrop-blur-xl">
        {/* Search Input - Full Width */}
        <div className="relative w-full">
          <Search size={13} className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input id="session-search" value={query} onChange={event => setQuery(event.target.value)} placeholder="Filter conversations..." className="h-8 w-full rounded-full border-border/60 bg-muted/40 pl-8 text-[11px] placeholder:text-muted-foreground/60 focus:bg-background focus:ring-1 focus:ring-primary/40" aria-label="Search sessions" />
        </div>

        {/* Source Segmented Filters */}
        <div className="mt-2 flex w-full gap-1.5" role="group" aria-label="Session source filter">
          <button type="button" aria-pressed={filterAriaPressed(sourceFilter, 'webui')} onClick={() => setSourceFilter(sourceFilter === 'webui' ? 'all' : 'webui')} className={cn('filter-pill flex-1 justify-center border', sourceFilter === 'webui' ? 'border-primary/40 bg-primary/15 text-primary font-medium' : 'border-border/60 bg-card/40 text-muted-foreground hover:border-border hover:text-foreground')}>
            <span>WebUI sessions</span>
            <span className="opacity-70 tabular-nums">({webCount})</span>
          </button>
          <button type="button" aria-pressed={filterAriaPressed(sourceFilter, 'cli')} onClick={() => setSourceFilter(sourceFilter === 'cli' ? 'all' : 'cli')} className={cn('filter-pill flex-1 justify-center border', sourceFilter === 'cli' ? 'border-primary/40 bg-primary/15 text-primary font-medium' : 'border-border/60 bg-card/40 text-muted-foreground hover:border-border hover:text-foreground')}>
            <span>CLI sessions</span>
            <span className="opacity-70 tabular-nums">({cliCount})</span>
          </button>
        </div>

        {/* Tag Filter Pills */}
        <div className="mt-2 flex flex-wrap items-center gap-1 overflow-x-auto pb-0.5" role="group" aria-label="Session tags">
          <button type="button" aria-pressed={filterAriaPressed(tagFilter, 'all')} onClick={() => setTagFilter('all')} className={cn('filter-pill border', tagFilter === 'all' ? 'border-primary/40 bg-primary/15 text-primary' : 'border-border/50 bg-card/30 text-muted-foreground hover:text-foreground')}>All</button>
          <button type="button" aria-pressed={filterAriaPressed(tagFilter, 'unassigned')} onClick={() => setTagFilter('unassigned')} className={cn('filter-pill border', tagFilter === 'unassigned' ? 'border-primary/40 bg-primary/15 text-primary' : 'border-border/50 bg-card/30 text-muted-foreground hover:text-foreground')}>Unassigned</button>
          {allTags.map((tag, i) => {
            const colors = ['bg-blue-400', 'bg-purple-400', 'bg-amber-400', 'bg-emerald-400', 'bg-rose-400', 'bg-cyan-400']
            const dotColor = colors[i % colors.length]
            return (
              <button key={tag} type="button" aria-pressed={filterAriaPressed(tagFilter, tag)} onClick={() => setTagFilter(tagFilter === tag ? 'all' : tag)} className={cn('filter-pill border', tagFilter === tag ? 'border-primary/40 bg-primary/15 text-primary' : 'border-border/50 bg-card/30 text-muted-foreground hover:text-foreground')}>
                <span className={cn('size-1.5 rounded-full', dotColor)} />
                <span className="max-w-24 truncate">{tag}</span>
              </button>
            )
          })}
        </div>

        {/* Auxiliary quick links */}
        <div className="mt-1.5 flex items-center justify-between px-0.5 text-[10px] text-muted-foreground/70">
          <button type="button" onClick={() => setShowArchived(v => !v)} className="hover:text-muted-foreground underline-offset-2 hover:underline">
            {showArchived ? 'Hide archived' : 'Show archived'}
          </button>
          <button type="button" onClick={() => { setBatchMode(v => !v); setSelected([]) }} className="hover:text-muted-foreground">
            {batchMode ? 'Done' : 'Select'}
          </button>
        </div>
      </div>

      {activeProjection?.external && <section className="mt-2 rounded-lg border border-primary/25 bg-primary/5 p-2" aria-label="External session handoff"><div className="flex items-center gap-2"><span className="size-1.5 rounded-full bg-primary" aria-hidden="true" /><p className="min-w-0 flex-1 truncate text-xs font-medium">{channelOf(activeProjection) || activeProjection.source} session</p><span className="text-[9px] uppercase tracking-wide text-muted-foreground">External</span></div><p className="mt-0.5 truncate text-[10px] text-muted-foreground">{field(activeProjection, 'identity', 'sender', 'user_name') || 'Identity not provided'} · {field(activeProjection, 'routing', 'route', 'model') || 'Routing not provided'}</p><Button type="button" size="sm" variant="outline" className="mt-1.5 h-7 w-full text-[10px]" disabled title="External handoff requires an upstream channel contract">Handoff unavailable</Button></section>}

      {batchMode && selected.length > 0 && (
        <div className="mt-2 flex items-center justify-between rounded-lg border bg-muted/40 px-2 py-1">
          <span className="text-[11px] text-muted-foreground">{selected.length} selected</span>
          <div className="flex gap-1">
            <Button size="sm" variant="outline" className="h-7 px-2 text-[10px]" onClick={async () => { await Promise.all(selected.map(id => onArchive(id, true))); clearBatch() }}>Archive</Button>
            <Button size="sm" variant="outline" className="h-7 px-2 text-[10px] text-destructive" onClick={async () => { await Promise.all(selected.map(onDelete)); clearBatch() }}>Delete</Button>
          </div>
        </div>
      )}

      {/* Session list items */}
      <div className="mt-2 space-y-3">
        {loading && <p className="px-2 py-2 text-xs text-muted-foreground" role="status">Loading session history...</p>}
        {error && <p className="px-2 py-2 text-xs text-destructive" role="alert">{error}</p>}
        {!loading && !error && !sessions.length && <p className="px-2 py-2 text-xs text-muted-foreground">No saved sessions yet.</p>}
        {!loading && !error && groups.map(group => (
          <div key={group.label} className="space-y-0.5">
            <p className="px-1.5 pb-1 text-[9px] font-semibold uppercase tracking-[0.14em] text-muted-foreground/60">{group.label}</p>
            {group.sessions.map(item => {
              const channel = channelOf(item)
              const isSelected = selected.includes(item.session_id)
              const isActive = item.session_id === activeSessionId
              const dateBadge = formatCompactDate(item.updated_at)
              return (
                <div key={item.session_id} className={cn('group relative flex min-h-8 items-center rounded-lg px-1.5 py-1 transition-colors hover:bg-accent/50', isActive && 'bg-accent/70 font-medium text-foreground shadow-sm ring-1 ring-border/80', isSelected && 'ring-1 ring-primary/60')}>
                  {batchMode && <Button type="button" variant="ghost" size="icon" className="size-7 shrink-0" onClick={() => toggleSelected(item.session_id)} aria-label={`${isSelected ? 'Deselect' : 'Select'} ${item.title || 'session'}`}>{isSelected ? <Check size={13} /> : <span className="size-2.5 rounded-sm border" />}</Button>}
                  <Button type="button" variant="ghost" size="sm" onClick={() => onSelectSession(item.session_id)} onKeyDown={event => { if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return; event.preventDefault(); const nextId = nextSessionId(visibleSessionIds, item.session_id, event.key === 'ArrowDown' ? 'next' : 'previous'); if (nextId) {
  onSelectSession(nextId)
  focusSessionButtonAfterSelection(document, nextId)
} }} data-session-id={item.session_id} aria-current={sessionRowAriaCurrent(item.session_id, activeSessionId)} className="h-auto min-h-7 min-w-0 flex-1 justify-start gap-1.5 rounded-md px-1 py-0 text-left text-xs">
                    {item.pinned ? <Pin size={11} className="shrink-0 text-primary" /> : null}
                    <span className="min-w-0 flex-1 truncate text-[11px] leading-tight text-foreground/90">{item.title || 'Untitled session'}</span>
                    {channel && <span className="shrink-0 rounded bg-muted px-1 text-[8px] font-semibold uppercase tracking-wider text-muted-foreground">{channel}</span>}
                    {dateBadge && <span className="shrink-0 text-[10px] tabular-nums text-muted-foreground/60">{dateBadge}</span>}
                  </Button>
                  <div className={sessionActionVisibilityClass()}>
                    <DropdownMenu label={`Actions for ${item.title || 'session'}`} items={[
                      { label: 'Rename conversation', icon: Pencil, onSelect: () => { setRenameTarget(item); setRenameValue(item.title || '') } },
                      { label: 'Duplicate conversation', icon: Copy, onSelect: () => { if (onDuplicate) void onDuplicate(item.session_id) } },
                      { label: 'Export Markdown', icon: Download, onSelect: () => { window.location.assign(sessionExportUrl(item.session_id)) } },
                      { label: item.pinned ? 'Unpin conversation' : 'Pin conversation', icon: Pin, onSelect: () => onPin(item.session_id, !item.pinned) },
                      { label: item.archived ? 'Unarchive conversation' : 'Archive conversation', icon: item.archived ? Archive : ArchiveX, onSelect: () => onArchive(item.session_id, !item.archived) },
                      { label: 'Delete conversation', icon: Trash2, onSelect: () => setDeleteTarget(item), destructive: true }
                    ]} />
                  </div>
                </div>
              )
            })}
          </div>
        ))}
        {!loading && !error && sessions.length > 0 && groups.length === 0 && <p className="px-2 py-2 text-xs text-muted-foreground">No sessions match these filters.</p>}
      </div>
    </ContextRail>
    <Dialog open={Boolean(renameTarget)} title="Rename session" onClose={() => setRenameTarget(null)}><form className="grid gap-3" onSubmit={event => { event.preventDefault(); if (renameTarget && renameValue.trim()) { void onRename(renameTarget.session_id, renameValue.trim()); setRenameTarget(null) } }}><Input autoFocus value={renameValue} onChange={event => setRenameValue(event.target.value)} placeholder="Session name" aria-label="Session name" /><div className="flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => setRenameTarget(null)}>Cancel</Button><Button type="submit" disabled={!renameValue.trim()}>Rename</Button></div></form></Dialog>
    <Dialog open={Boolean(deleteTarget)} title="Delete session?" onClose={() => setDeleteTarget(null)}><p className="text-sm text-muted-foreground">This removes the saved session and its transcript.</p><div className="mt-4 flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => setDeleteTarget(null)}>Cancel</Button><Button type="button" onClick={() => { if (deleteTarget) void onDelete(deleteTarget.session_id); setDeleteTarget(null) }}>Delete session</Button></div></Dialog>
  </>
}
