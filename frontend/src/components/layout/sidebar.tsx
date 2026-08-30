import { Archive, ArchiveX, Bot, BrainCircuit, CheckSquare2, Clock3, FolderKanban, MessageSquareText, Pencil, Pin, Plus, Search, Settings2, Sparkles, Trash2, UsersRound, Wrench, X } from 'lucide-react'
import { Button } from '../ui/button'
import { Badge } from '../ui/badge'
import { cn } from '../../lib/cn'
import { filterSessions, groupSessionsByDate, type SessionSummary } from '../../lib/chat-contract'
import { useState } from 'react'

const navigation = [
  { label: 'Chat', icon: MessageSquareText, active: true },
  { label: 'Tasks', icon: Clock3 },
  { label: 'Skills', icon: Wrench },
  { label: 'Memory', icon: BrainCircuit },
  { label: 'Profiles', icon: UsersRound },
  { label: 'Todos', icon: CheckSquare2 },
  { label: 'Spaces', icon: FolderKanban },
]

export function Sidebar({ onNewChat, onNavigate, currentView, sessions, activeSessionId, onSelectSession, onRename, onPin, onArchive, onDelete, loading, error, mobileOpen, onClose }: { onNewChat: () => void; onNavigate: (view: string) => void; currentView: string; sessions: SessionSummary[]; activeSessionId: string; onSelectSession: (id: string) => void; onRename: (id: string) => void; onPin: (id: string, pinned: boolean) => void; onArchive: (id: string, archived: boolean) => void; onDelete: (id: string) => void; loading: boolean; error?: string; mobileOpen?: boolean; onClose?: () => void }) {
  const [query, setQuery] = useState('')
  const groups = groupSessionsByDate(filterSessions(sessions, query))
  return (
    <aside className={cn('h-screen w-[264px] shrink-0 flex-col border-r bg-card/55 p-3 backdrop-blur-xl', mobileOpen ? 'fixed inset-y-0 left-0 z-50 flex shadow-2xl' : 'hidden lg:flex')}>
      <div className="flex h-12 items-center gap-3 px-2">
        <div className="grid size-8 place-items-center rounded-xl bg-primary text-primary-foreground shadow-[0_0_30px_rgb(139_92_246/0.25)]"><Sparkles size={16} /></div>
        <div><div className="text-sm font-semibold tracking-tight">Hermes Studio</div><div className="text-[11px] text-muted-foreground">Agent workspace</div></div>
        <Badge className="ml-auto border-primary/20 bg-primary/10 text-primary">MVP</Badge>{onClose && <Button variant="ghost" size="icon" className="lg:hidden" onClick={onClose} aria-label="Close navigation"><X size={16} /></Button>}
      </div>

      <Button className="mt-4 w-full justify-start" onClick={() => { onNewChat(); onClose?.() }}><Plus size={16} /> New conversation</Button>

      <nav className="mt-5 space-y-1" aria-label="Main navigation">
        {navigation.map(({ label, icon: Icon }) => (
          <button key={label} onClick={() => { onNavigate(label.toLowerCase()); onClose?.() }} aria-current={currentView === label.toLowerCase() ? 'page' : undefined} className={cn('flex h-9 w-full items-center gap-3 rounded-lg px-3 text-sm transition-colors', currentView === label.toLowerCase() ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-accent/50')}>
            <Icon size={16} /><span>{label}</span>
          </button>
        ))}
      </nav>

      <div className="mt-6 flex items-center gap-2 px-2 text-[11px] font-medium uppercase tracking-[0.16em] text-muted-foreground"><label htmlFor="session-search">Recent sessions</label><Search size={13} className="ml-auto" /></div>
      <input id="session-search" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search sessions" className="mt-2 h-8 w-full rounded-lg border bg-background/50 px-3 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring" />
      <div className="mt-2 space-y-1">
        {loading && <p className="px-3 py-2 text-xs text-muted-foreground" role="status">Loading session history…</p>}
        {error && <p className="px-3 py-2 text-xs text-destructive" role="alert">{error}</p>}
        {!loading && !error && !sessions.length && <p className="px-3 py-2 text-xs text-muted-foreground">No saved sessions yet.</p>}
        {!loading && !error && groups.map((group) => <div key={group.label} className="pt-2"><p className="px-3 pb-1 text-[10px] uppercase tracking-[0.14em] text-muted-foreground">{group.label}</p>{group.sessions.map((item) => <div key={item.session_id} className={cn('group flex min-h-11 items-center gap-1 rounded-lg px-2 py-1 transition-colors hover:bg-accent/70', item.session_id === activeSessionId && 'bg-accent/60')}><button type="button" onClick={() => { onSelectSession(item.session_id); onClose?.() }} aria-current={item.session_id === activeSessionId ? 'page' : undefined} className="flex min-w-0 flex-1 items-center gap-2 rounded-md px-1 py-1 text-left text-sm focus-visible:ring-2 focus-visible:ring-ring"><Archive size={14} className="shrink-0 text-muted-foreground" /><span className="min-w-0 flex-1"><span className="block truncate">{item.title || 'Untitled session'}</span>{(item.project || item.project_id || item.tags?.length) && <span className="block truncate text-[10px] text-muted-foreground">{[item.project || item.project_id, ...(item.tags || [])].filter(Boolean).join(' · ')}</span>}</span>{item.pinned && <Pin size={12} className="shrink-0 text-primary" />}</button><div className="flex shrink-0 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"><button type="button" className="grid size-7 place-items-center rounded text-muted-foreground hover:bg-background hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring" onClick={() => onRename(item.session_id)} aria-label={`Rename ${item.title || 'session'}`} title="Rename"><Pencil size={12} /></button><button type="button" className="grid size-7 place-items-center rounded text-muted-foreground hover:bg-background hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring" onClick={() => onPin(item.session_id, !item.pinned)} aria-label={item.pinned ? `Unpin ${item.title || 'session'}` : `Pin ${item.title || 'session'}`} title={item.pinned ? 'Unpin' : 'Pin'}><Pin size={12} /></button><button type="button" className="grid size-7 place-items-center rounded text-muted-foreground hover:bg-background hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring" onClick={() => onArchive(item.session_id, !item.archived)} aria-label={item.archived ? `Unarchive ${item.title || 'session'}` : `Archive ${item.title || 'session'}`} title={item.archived ? 'Unarchive' : 'Archive'}>{item.archived ? <Archive size={12} /> : <ArchiveX size={12} />}</button><button type="button" className="grid size-7 place-items-center rounded text-muted-foreground hover:bg-destructive/20 hover:text-destructive focus-visible:ring-2 focus-visible:ring-ring" onClick={() => onDelete(item.session_id)} aria-label={`Delete ${item.title || 'session'}`} title="Delete"><Trash2 size={12} /></button></div></div>)}</div>)}
        {!loading && !error && sessions.length > 0 && groups.length === 0 && <p className="px-3 py-2 text-xs text-muted-foreground">No matching sessions.</p>}
      </div>

      <div className="mt-auto rounded-xl border bg-background/50 p-3">
        <div className="flex items-center gap-2 text-xs font-medium"><Bot size={14} className="text-primary" /> Gateway-first runtime</div>
        <p className="mt-1.5 text-[11px] leading-4 text-muted-foreground">Credentials stay in the Go service. Browser traffic uses normalized events.</p>
      </div>
      <Button variant="ghost" className="mt-2 w-full justify-start"><Settings2 size={16} /> Settings</Button>
    </aside>
  )
}
