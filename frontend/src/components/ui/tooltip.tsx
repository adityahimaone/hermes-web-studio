import type { ReactNode } from 'react'

type Props = { label: string; children: ReactNode }

export function Tooltip({ label, children }: Props) {
  return <span className="group/tooltip relative flex">
    {children}
    <span role="tooltip" className="pointer-events-none absolute left-full top-1/2 z-50 ml-2 -translate-y-1/2 whitespace-nowrap rounded-md border bg-popover px-2 py-1 text-xs font-medium text-popover-foreground opacity-0 shadow-lg transition-opacity duration-150 group-hover/tooltip:opacity-100 group-focus-within/tooltip:opacity-100">{label}</span>
  </span>
}
