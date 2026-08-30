import type { SelectHTMLAttributes } from 'react'
import { cn } from '../../lib/cn'

export function Select({ className, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={cn('h-10 rounded-lg border bg-card px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50', className)} {...props} />
}
