import { ArrowDown, ArrowUp, BrainCircuit, CheckSquare2, CircleStar, Clock3, FolderKanban, MessageSquareText, Settings2, SlidersHorizontal, Sparkles, UsersRound, Wrench, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Button } from '../ui/button'
import { Tooltip } from '../ui/tooltip'
import { Dialog } from '../ui/dialog'
import { cn } from '../../lib/cn'

const defaultNavigation = [
  { label: 'Chat', icon: MessageSquareText },
  { label: 'Tasks', icon: Clock3 },
  { label: 'Skills', icon: Wrench },
  { label: 'Memory', icon: BrainCircuit },
  { label: 'Profiles', icon: UsersRound },
  { label: 'Todos', icon: CheckSquare2 },
  { label: 'Goals', icon: CircleStar },
  { label: 'Spaces', icon: FolderKanban },
]
const navigationKey = 'hermes-primary-navigation'
type NavItem = (typeof defaultNavigation)[number]

function loadNavigation() {
  try {
    const saved = JSON.parse(window.localStorage.getItem(navigationKey) || 'null') as string[] | null
    if (!Array.isArray(saved)) return defaultNavigation
    const byLabel = new Map(defaultNavigation.map(item => [item.label, item]))
    const ordered = saved.flatMap(label => { const item = byLabel.get(label); if (!item) return []; byLabel.delete(label); return [item] })
    return [...ordered, ...byLabel.values()]
  } catch { return defaultNavigation }
}

type Props = { onNewChat: () => void; onNavigate: (view: string) => void; currentView: string; mobileOpen?: boolean; onClose?: () => void }
export function Sidebar({ onNewChat, onNavigate, currentView, mobileOpen, onClose }: Props) {
  const [items, setItems] = useState<NavItem[]>(loadNavigation)
  const [visible, setVisible] = useState(() => { try { const saved = JSON.parse(window.localStorage.getItem(`${navigationKey}-visibility`) || 'null') as string[] | null; return new Set(saved || defaultNavigation.map(item => item.label)) } catch { return new Set(defaultNavigation.map(item => item.label)) } })
  const [customizeOpen, setCustomizeOpen] = useState(false)
  const persist = (next: NavItem[]) => { setItems(next); window.localStorage.setItem(navigationKey, JSON.stringify(next.map(item => item.label))) }
  const persistVisible = (next: Set<string>) => { setVisible(next); window.localStorage.setItem(`${navigationKey}-visibility`, JSON.stringify([...next])) }
  const move = (index: number, direction: -1 | 1) => { const next = [...items]; const target = index + direction; if (target < 0 || target >= next.length) return; [next[index], next[target]] = [next[target], next[index]]; persist(next) }
  useEffect(() => { if (!onClose) return; const media = window.matchMedia('(min-width: 1024px)'); const closeOnDesktop = () => { if (media.matches) onClose() }; closeOnDesktop(); media.addEventListener('change', closeOnDesktop); return () => media.removeEventListener('change', closeOnDesktop) }, [onClose])
  return <aside data-testid="primary-rail" className={cn('h-screen w-[72px] shrink-0 flex-col items-center border-r bg-card/55 px-2 py-3 backdrop-blur-xl', mobileOpen ? 'fixed inset-y-0 left-0 z-50 flex w-[264px] items-stretch px-3 shadow-2xl' : 'hidden lg:flex')}>
    <div className={cn('flex h-12 items-center', mobileOpen ? 'gap-3 px-2' : 'justify-center')}><div className="grid size-11 shrink-0 place-items-center rounded-xl bg-primary text-primary-foreground shadow-[0_0_30px_rgb(139_92_246/0.25)]"><Sparkles size={20} /></div><div className={cn('min-w-0', mobileOpen ? 'block' : 'hidden')}><div className="truncate text-sm font-semibold tracking-tight">Hermes Studio</div><div className="text-[11px] text-muted-foreground">Agent workspace</div></div>{mobileOpen && onClose && <Button variant="ghost" size="icon" className="ml-auto lg:hidden" onClick={onClose} aria-label="Close navigation"><X size={16} /></Button>}</div>
    <Tooltip label="New conversation"><Button className={cn('mt-4', mobileOpen ? 'w-full justify-start' : 'size-11 px-0')} onClick={() => { onNewChat(); onClose?.() }} aria-label="New conversation"><Sparkles size={20} /><span className={cn(mobileOpen ? 'inline' : 'hidden')}>New conversation</span></Button></Tooltip>
    <nav data-testid="primary-navigation" className={cn('mt-5 space-y-1', mobileOpen ? 'w-full' : 'w-11')} aria-label="Main navigation">{items.filter(item => visible.has(item.label)).map(({ label, icon: Icon }) => <Tooltip key={label} label={label}><Button variant="ghost" onClick={() => { onNavigate(label.toLowerCase()); onClose?.() }} aria-label={label} aria-current={currentView === label.toLowerCase() ? 'page' : undefined} className={cn('h-11', mobileOpen ? 'w-full justify-start' : 'w-11 justify-center px-0', currentView === label.toLowerCase() ? 'bg-accent text-accent-foreground' : 'text-muted-foreground')}><Icon aria-hidden="true" size={19} /><span className={cn(mobileOpen ? 'inline' : 'hidden')}>{label}</span></Button></Tooltip>)}</nav>
    <div className={cn('mt-auto flex gap-1', mobileOpen ? 'w-full' : 'w-11')}><Tooltip label="Customize navigation"><Button variant="ghost" size="icon" className={cn(mobileOpen ? 'flex-1 justify-start px-3' : 'size-11')} onClick={() => setCustomizeOpen(true)} aria-label="Customize navigation"><SlidersHorizontal size={19} /><span className={cn(mobileOpen ? 'inline' : 'hidden')}>Customize navigation</span></Button></Tooltip><Tooltip label="Settings"><Button variant="ghost" size="icon" className={cn(mobileOpen ? 'size-11' : 'hidden')} onClick={() => { onNavigate('settings'); onClose?.() }} aria-label="Settings"><Settings2 size={19} /></Button></Tooltip></div>
    <Tooltip label="Settings"><Button variant="ghost" size="icon" className={cn('mt-1', mobileOpen ? 'hidden' : 'size-11')} onClick={() => { onNavigate('settings'); onClose?.() }} aria-label="Settings"><Settings2 size={19} /></Button></Tooltip>
    <Dialog open={customizeOpen} title="Customize navigation" onClose={() => setCustomizeOpen(false)}><p className="text-xs text-muted-foreground">Choose visible sections and arrange their order. Chat stays available as the primary workspace.</p><div className="mt-4 space-y-2">{items.map((item, index) => { const active = visible.has(item.label); return <div key={item.label} className="flex items-center gap-2 rounded-lg border px-2 py-1.5"><Button type="button" variant={active ? 'default' : 'outline'} size="sm" role="checkbox" aria-checked={active} disabled={item.label === 'Chat'} onClick={() => { const next = new Set(visible); if (active) next.delete(item.label); else next.add(item.label); persistVisible(next) }} className="min-w-24 justify-start text-xs">{active ? 'Visible' : 'Hidden'} · {item.label}</Button><Button type="button" variant="ghost" size="icon" className="ml-auto size-8" disabled={index === 0} onClick={() => move(index, -1)} aria-label={`Move ${item.label} up`}><ArrowUp size={14} /></Button><Button type="button" variant="ghost" size="icon" className="size-8" disabled={index === items.length - 1} onClick={() => move(index, 1)} aria-label={`Move ${item.label} down`}><ArrowDown size={14} /></Button></div>})}</div><div className="mt-4 flex justify-end"><Button type="button" variant="outline" onClick={() => setCustomizeOpen(false)}>Done</Button></div></Dialog>
  </aside>
}
