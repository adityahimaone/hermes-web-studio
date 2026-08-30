import type { InputHTMLAttributes } from 'react'
import { cn } from '../../lib/cn'

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn('h-10 min-w-0 rounded-lg border bg-card px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50', className)} {...props} />
}
