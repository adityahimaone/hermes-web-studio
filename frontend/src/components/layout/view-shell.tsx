import type { ReactNode } from 'react'
import { ContextRail, ExpandRailButton } from './context-rail'

type ViewShellProps = {
  title: string
  subtitle: string
  sidebarOpen: boolean
  onToggleSidebar: () => void
  children: ReactNode
  railContent?: ReactNode
}

/** Shared contextual shell for non-chat views. It keeps navigation context visible without inventing controls. */
export function ViewShell({ title, subtitle, sidebarOpen, onToggleSidebar, children, railContent }: ViewShellProps) {
  return <div className="view-shell relative flex min-h-full">
    <ContextRail title={title} subtitle={subtitle} open={sidebarOpen} onToggle={onToggleSidebar}>
      {railContent || <div className="view-shell__rail-note rounded-lg border border-border/60 bg-muted/20 p-3 text-xs text-muted-foreground">Use the main pane to inspect and update {title.toLocaleLowerCase()}.</div>}
    </ContextRail>
    {!sidebarOpen && <ExpandRailButton label={title} onClick={onToggleSidebar} />}
    <div className="view-shell__content min-w-0 flex-1">{children}</div>
  </div>
}
