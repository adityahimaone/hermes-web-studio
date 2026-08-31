import { MoreVertical } from 'lucide-react'
import { useEffect, useLayoutEffect, useRef, useState, type ComponentType } from 'react'
import { Button } from './button'

type MenuItem = { label: string; icon: ComponentType<{ size?: number }>; onSelect: () => void; destructive?: boolean }

export function DropdownMenu({ label, items }: { label: string; items: MenuItem[] }) {
  const [open, setOpen] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const itemRefs = useRef<Array<HTMLButtonElement | null>>([])

  useEffect(() => {
    if (!open) return
    const close = (event: MouseEvent) => { if (rootRef.current && !rootRef.current.contains(event.target as Node)) setOpen(false) }
    const escape = (event: KeyboardEvent) => { if (event.key === 'Escape') setOpen(false) }
    document.addEventListener('mousedown', close)
    document.addEventListener('keydown', escape)
    return () => { document.removeEventListener('mousedown', close); document.removeEventListener('keydown', escape) }
  }, [open])
  useLayoutEffect(() => { if (open) itemRefs.current[0]?.focus() }, [open])

  return <div ref={rootRef} className="relative shrink-0">
    <Button ref={triggerRef} type="button" variant="ghost" size="icon" className="size-11" aria-label={label} aria-expanded={open} aria-haspopup="menu" onClick={() => setOpen(value => !value)}><MoreVertical size={16} /></Button>
    {open && <div ref={menuRef} role="menu" aria-label={label} className="absolute right-0 top-9 z-50 min-w-52 rounded-lg border bg-popover p-1.5 text-popover-foreground shadow-xl">
      {items.map(({ label: itemLabel, icon: Icon, onSelect, destructive }, index) => <button ref={element => { itemRefs.current[index] = element }} key={itemLabel} type="button" role="menuitem" className={`flex min-h-10 w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent focus-visible:bg-accent focus-visible:outline-none ${destructive ? 'text-destructive hover:bg-destructive/10' : ''}`} onKeyDown={event => { const next = event.key === 'ArrowDown' ? index + 1 : event.key === 'ArrowUp' ? index - 1 : event.key === 'Home' ? 0 : event.key === 'End' ? items.length - 1 : -1; if (next >= 0) { event.preventDefault(); itemRefs.current[(next + items.length) % items.length]?.focus() } }} onClick={() => { triggerRef.current?.focus(); onSelect(); setOpen(false); requestAnimationFrame(() => triggerRef.current?.focus()) }}><Icon size={15} />{itemLabel}</button>)}
    </div>}
  </div>
}
