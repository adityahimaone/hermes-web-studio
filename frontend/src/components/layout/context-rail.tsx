import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { ReactNode } from 'react'
import { Button } from '../ui/button'
import { cn } from '../../lib/cn'

type Props = { title: string; subtitle?: string; open: boolean; onToggle: () => void; action?: ReactNode; children: ReactNode; className?: string }

export function ContextRail({ title, subtitle, open, onToggle, action, children, className }: Props) {
  if (!open) return null
  return <aside data-testid="context-rail" className={cn('context-rail hidden shrink-0 flex-col border-r bg-card/25 p-3 lg:flex', className)} aria-label={`Context ${title} navigation`}>
    <div className="context-rail__header grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 border-b pb-3">
      <div className="min-w-0"><p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-foreground">{title}</p>{subtitle && <p className="mt-0.5 truncate text-[11px] text-muted-foreground">{subtitle}</p>}</div>
      {action}
      <Button type="button" variant="ghost" size="icon" className="size-10 shrink-0" onClick={onToggle} aria-label={`Collapse ${title} sidebar`}><ChevronLeft size={18} /></Button>
    </div>
    <div className="context-rail__body thin-scrollbar mt-3 min-h-0 flex-1 overflow-y-auto">{children}</div>
  </aside>
}

export function ExpandRailButton({ label, onClick }: { label: string; onClick: () => void }) {
  return <Button type="button" variant="ghost" size="icon" className="context-rail__expand absolute left-0 top-1/2 z-10 size-10 -translate-x-1/2 -translate-y-1/2 rounded-full border bg-card shadow-lg" onClick={onClick} aria-label={`Expand ${label} sidebar`}><ChevronRight size={18} /></Button>
}
