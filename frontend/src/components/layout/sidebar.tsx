import { Archive, Bot, BrainCircuit, CheckSquare2, Clock3, FolderKanban, MessageSquareText, Plus, Search, Settings2, Sparkles, UsersRound, Wrench } from 'lucide-react'
import { Button } from '../ui/button'
import { Badge } from '../ui/badge'
import { cn } from '../../lib/cn'

const navigation = [
  { label: 'Chat', icon: MessageSquareText, active: true },
  { label: 'Tasks', icon: Clock3 },
  { label: 'Skills', icon: Wrench },
  { label: 'Memory', icon: BrainCircuit },
  { label: 'Profiles', icon: UsersRound },
  { label: 'Todos', icon: CheckSquare2 },
  { label: 'Spaces', icon: FolderKanban },
]

export function Sidebar({ onNewChat }: { onNewChat: () => void }) {
  return (
    <aside className="hidden h-screen w-[264px] shrink-0 flex-col border-r bg-card/55 p-3 backdrop-blur-xl lg:flex">
      <div className="flex h-12 items-center gap-3 px-2">
        <div className="grid size-8 place-items-center rounded-xl bg-primary text-primary-foreground shadow-[0_0_30px_rgb(139_92_246/0.25)]"><Sparkles size={16} /></div>
        <div><div className="text-sm font-semibold tracking-tight">Hermes Studio</div><div className="text-[11px] text-muted-foreground">Agent workspace</div></div>
        <Badge className="ml-auto border-primary/20 bg-primary/10 text-primary">MVP</Badge>
      </div>

      <Button className="mt-4 w-full justify-start" onClick={onNewChat}><Plus size={16} /> New conversation</Button>

      <nav className="mt-5 space-y-1" aria-label="Main navigation">
        {navigation.map(({ label, icon: Icon, active }) => (
          <button key={label} disabled={!active} className={cn('flex h-9 w-full items-center gap-3 rounded-lg px-3 text-sm transition-colors', active ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-accent/50 disabled:opacity-45')}>
            <Icon size={16} /><span>{label}</span>{!active && <span className="ml-auto text-[10px]">Soon</span>}
          </button>
        ))}
      </nav>

      <div className="mt-6 flex items-center px-2 text-[11px] font-medium uppercase tracking-[0.16em] text-muted-foreground"><span>Recent</span><Search size={13} className="ml-auto" /></div>
      <div className="mt-2 space-y-1">
        <button className="w-full rounded-lg bg-accent/60 px-3 py-2.5 text-left"><div className="truncate text-sm">New Hermes conversation</div><div className="mt-1 text-[11px] text-muted-foreground">Just now</div></button>
        <button disabled className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-muted-foreground opacity-50"><Archive size={14} /> Session history after M1</button>
      </div>

      <div className="mt-auto rounded-xl border bg-background/50 p-3">
        <div className="flex items-center gap-2 text-xs font-medium"><Bot size={14} className="text-primary" /> Gateway-first runtime</div>
        <p className="mt-1.5 text-[11px] leading-4 text-muted-foreground">Credentials stay in the Go service. Browser traffic uses normalized events.</p>
      </div>
      <Button variant="ghost" className="mt-2 w-full justify-start"><Settings2 size={16} /> Settings</Button>
    </aside>
  )
}

