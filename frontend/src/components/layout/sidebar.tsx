import { BrainCircuit, CheckSquare2, CircleStar, Clock3, FolderKanban, MessageSquareText, Settings2, Sparkles, UsersRound, Wrench, X } from 'lucide-react'
import { Button } from '../ui/button'
import { Tooltip } from '../ui/tooltip'
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
  return <aside data-testid="primary-rail" className={cn('h-screen w-[72px] shrink-0 flex-col items-center border-r bg-card/55 px-2 py-3 backdrop-blur-xl', mobileOpen ? 'fixed inset-y-0 left-0 z-50 flex w-[264px] items-stretch px-3 shadow-2xl' : 'hidden lg:flex')}>
    <div className={cn('flex h-12 items-center', mobileOpen ? 'gap-3 px-2' : 'justify-center')}><div className="grid size-11 shrink-0 place-items-center rounded-xl bg-primary text-primary-foreground shadow-[0_0_30px_rgb(139_92_246/0.25)]"><Sparkles size={20} /></div><div className={cn('min-w-0', mobileOpen ? 'block' : 'hidden')}><div className="truncate text-sm font-semibold tracking-tight">Hermes Studio</div><div className="text-[11px] text-muted-foreground">Agent workspace</div></div>{mobileOpen && onClose && <Button variant="ghost" size="icon" className="ml-auto lg:hidden" onClick={onClose} aria-label="Close navigation"><X size={16} /></Button>}</div>
    <Tooltip label="New conversation"><Button className={cn('mt-4', mobileOpen ? 'w-full justify-start' : 'size-11 px-0')} onClick={() => { onNewChat(); onClose?.() }} aria-label="New conversation"><Sparkles size={20} /><span className={cn(mobileOpen ? 'inline' : 'hidden')}>New conversation</span></Button></Tooltip>
    <nav data-testid="primary-navigation" className={cn('mt-5 space-y-1', mobileOpen ? 'w-full' : 'w-11')} aria-label="Main navigation">{navigation.map(({ label, icon: Icon }) => <Tooltip key={label} label={label}><Button variant="ghost" onClick={() => { onNavigate(label.toLowerCase()); onClose?.() }} aria-label={label} aria-current={currentView === label.toLowerCase() ? 'page' : undefined} className={cn('h-11', mobileOpen ? 'w-full justify-start' : 'w-11 justify-center px-0', currentView === label.toLowerCase() ? 'bg-accent text-accent-foreground' : 'text-muted-foreground')}><Icon aria-hidden="true" size={19} /><span className={cn(mobileOpen ? 'inline' : 'hidden')}>{label}</span></Button></Tooltip>)}</nav>
    <Tooltip label="Settings"><Button variant="ghost" className={cn('mt-auto', mobileOpen ? 'w-full justify-start' : 'size-11 px-0')} onClick={() => { onNavigate('settings'); onClose?.() }} aria-label="Settings"><Settings2 size={19} /><span className={cn(mobileOpen ? 'inline' : 'hidden')}>Settings</span></Button></Tooltip>
  </aside>
}
