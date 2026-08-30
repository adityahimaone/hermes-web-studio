import { BrainCircuit, CheckSquare2, CircleStar, Clock3, FolderKanban, MessageSquareText, Settings2, Sparkles, UsersRound, Wrench, X } from 'lucide-react'
import { Button } from '../ui/button'
import { Badge } from '../ui/badge'
import { cn } from '../../lib/cn'

const navigation = [
  { label: 'Chat', icon: MessageSquareText },
  { label: 'Tasks', icon: Clock3 },
  { label: 'Skills', icon: Wrench },
  { label: 'Memory', icon: BrainCircuit },
  { label: 'Profiles', icon: UsersRound },
  { label: 'Todos', icon: CheckSquare2 },
  { label: 'Goals', icon: CircleStar },
  { label: 'Spaces', icon: FolderKanban },
]

type Props = { onNewChat: () => void; onNavigate: (view: string) => void; currentView: string; mobileOpen?: boolean; onClose?: () => void }
export function Sidebar({ onNewChat, onNavigate, currentView, mobileOpen, onClose }: Props) {
  return <aside className={cn('h-screen w-[264px] shrink-0 flex-col border-r bg-card/55 p-3 backdrop-blur-xl', mobileOpen ? 'fixed inset-y-0 left-0 z-50 flex shadow-2xl' : 'hidden lg:flex')}>
    <div className="flex h-12 items-center gap-3 px-2"><div className="grid size-8 place-items-center rounded-xl bg-primary text-primary-foreground shadow-[0_0_30px_rgb(139_92_246/0.25)]"><Sparkles size={16} /></div><div><div className="text-sm font-semibold tracking-tight">Hermes Studio</div><div className="text-[11px] text-muted-foreground">Agent workspace</div></div><Badge className="ml-auto border-primary/20 bg-primary/10 text-primary">MVP</Badge>{onClose && <Button variant="ghost" size="icon" className="lg:hidden" onClick={onClose} aria-label="Close navigation"><X size={16} /></Button>}</div>
    <Button className="mt-4 w-full justify-start" onClick={() => { onNewChat(); onClose?.() }}><Sparkles size={16} /> New conversation</Button>
    <nav className="mt-5 space-y-1" aria-label="Main navigation">{navigation.map(({ label, icon: Icon }) => <Button key={label} variant="ghost" onClick={() => { onNavigate(label.toLowerCase()); onClose?.() }} aria-current={currentView === label.toLowerCase() ? 'page' : undefined} className={cn('h-9 w-full justify-start', currentView === label.toLowerCase() ? 'bg-accent text-accent-foreground' : 'text-muted-foreground')}><Icon size={16} /><span>{label}</span></Button>)}</nav>
    <div className="mt-auto rounded-xl border bg-background/50 p-3"><div className="flex items-center gap-2 text-xs font-medium"><MessageSquareText size={14} className="text-primary" /> Gateway-first runtime</div><p className="mt-1.5 text-[11px] leading-4 text-muted-foreground">Credentials stay in the Go service. Browser traffic uses normalized events.</p></div>
    <Button variant="ghost" className="mt-2 w-full justify-start" onClick={() => { onNavigate('settings'); onClose?.() }}><Settings2 size={16} /> Settings</Button>
  </aside>
}
