import { forwardRef, type TextareaHTMLAttributes } from 'react'
import { cn } from '../../lib/cn'

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(function Textarea({ className, ...props }, ref) {
  return <textarea ref={ref} className={cn('min-h-20 w-full resize-y rounded-lg border bg-card px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50', className)} {...props} />
})
