import type { ReactNode } from 'react'
import { X } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { Button } from './button'

type Props = { open: boolean; title: string; children: ReactNode; onClose: () => void }
export function Dialog({ open, title, children, onClose }: Props) {
  const closeRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    if (!open) return
    closeRef.current?.focus()
    const handleKeyDown = (event: KeyboardEvent) => { if (event.key === 'Escape') onClose() }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [open, onClose])
  if (!open) return null
  return createPortal(<div className="fixed inset-0 z-[80] grid place-items-center bg-black/65 p-4" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) onClose() }}><section role="dialog" aria-modal="true" aria-labelledby="dialog-title" className="w-full max-w-md rounded-xl border bg-card p-5 shadow-2xl shadow-black/50"><div className="flex items-center gap-3"><h2 id="dialog-title" className="min-w-0 flex-1 text-base font-semibold">{title}</h2><Button ref={closeRef} type="button" variant="ghost" size="icon" onClick={onClose} aria-label="Close dialog"><X size={16} /></Button></div><div className="mt-4">{children}</div></section></div>, document.body)
}
