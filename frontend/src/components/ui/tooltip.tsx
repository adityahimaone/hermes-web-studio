import { Children, cloneElement, isValidElement, useId, type ReactNode } from 'react'

type Props = { label: string; children: ReactNode }

export function Tooltip({ label, children }: Props) {
  const tooltipId = useId()
  const trigger = isValidElement<{ 'aria-describedby'?: string }>(Children.only(children))
    ? cloneElement(Children.only(children) as React.ReactElement<{ 'aria-describedby'?: string }>, { 'aria-describedby': tooltipId })
    : children
  return <span className="group/tooltip relative flex">
    {trigger}
    <span id={tooltipId} role="tooltip" className="pointer-events-none absolute left-full top-1/2 z-50 ml-2 -translate-y-1/2 whitespace-nowrap rounded-md border bg-popover px-2 py-1 text-xs font-medium text-popover-foreground opacity-0 shadow-lg transition-opacity duration-150 group-hover/tooltip:opacity-100 group-focus-within/tooltip:opacity-100">{label}</span>
  </span>
}
