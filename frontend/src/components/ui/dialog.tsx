import type { ReactNode } from 'react'
import { X } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { Button } from './button'

type Props = { open: boolean; title: string; children: ReactNode; onClose: () => void }
export function Dialog({ open, title, children, onClose }: Props) {
  const closeRef = useRef<HTMLButtonElement>(null)
  const onCloseRef = useRef(onClose)
  onCloseRef.current = onClose
  useEffect(() => {
    if (!open) return
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null
    closeRef.current?.focus()
    const handleKeyDown = (event: KeyboardEvent) => { if (event.key === 'Escape') onCloseRef.current() }
    document.addEventListener('keydown', handleKeyDown)
    return () => { document.removeEventListener('keydown', handleKeyDown); requestAnimationFrame(() => opener?.focus()) }
  }, [open])
  if (!open) return null
  return createPortal(<div className="fixed inset-0 z-[80] grid place-items-center bg-black/65 p-3 sm:p-4" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) onClose() }}><section role="dialog" aria-modal="true" aria-labelledby="dialog-title" className="flex max-h-[calc(100dvh-1.5rem)] w-full max-w-md flex-col overflow-hidden rounded-xl border bg-card p-5 shadow-2xl shadow-black/50 sm:max-h-[calc(100dvh-2rem)]"><div className="flex shrink-0 items-center gap-3"><h2 id="dialog-title" className="min-w-0 flex-1 text-base font-semibold">{title}</h2><Button ref={closeRef} type="button" variant="ghost" size="icon" onClick={onClose} aria-label="Close dialog"><X size={16} /></Button></div><div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto overscroll-contain pr-1 pt-4">{children}</div></section></div>, document.body)
}
