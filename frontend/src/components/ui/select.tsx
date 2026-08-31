import { Children, useEffect, useId, useRef, useState, type ChangeEventHandler, type ReactNode, type SelectHTMLAttributes } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '../../lib/cn'

type Option = { value: string; label: ReactNode; disabled: boolean }
type OptionElement = React.ReactElement<{ value?: string | number; disabled?: boolean; children?: ReactNode }>

function isOption(child: ReactNode): child is OptionElement {
  return typeof child === 'object' && child !== null && 'type' in child && child.type === 'option'
}

function readOptions(children: ReactNode): Option[] {
  return Children.toArray(children).flatMap(child => {
    if (!isOption(child)) return []
    const value = child.props.value === undefined ? String(child.props.children ?? '') : String(child.props.value)
    return [{ value, label: child.props.children, disabled: Boolean(child.props.disabled) }]
  })
}

type Props = Omit<SelectHTMLAttributes<HTMLSelectElement>, 'onChange'> & { onChange?: ChangeEventHandler<HTMLSelectElement> }

export function Select({ className, children, value, defaultValue, onChange, disabled, id, name, form, required, ...selectProps }: Props) {
  const options = readOptions(children)
  const listboxId = useId()
  const rootRef = useRef<HTMLDivElement>(null)
  const nativeRef = useRef<HTMLSelectElement>(null)
  const firstValue = options[0]?.value ?? ''
  const initialValue = defaultValue === undefined ? firstValue : String(Array.isArray(defaultValue) ? defaultValue[0] ?? '' : defaultValue)
  const [uncontrolledValue, setUncontrolledValue] = useState(initialValue)
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(() => Math.max(0, options.findIndex(option => option.value === String(value ?? uncontrolledValue))))
  const selectedValue = value === undefined ? uncontrolledValue : String(value)
  const selectedIndex = options.findIndex(option => option.value === selectedValue)
  const selectedOption = options[selectedIndex] ?? options[0]

  useEffect(() => {
    if (!open) return
    const close = (event: MouseEvent) => { if (!rootRef.current?.contains(event.target as Node)) setOpen(false) }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [open])

  useEffect(() => { if (selectedIndex >= 0) setActiveIndex(selectedIndex) }, [selectedIndex])

  function commit(index: number) {
    const option = options[index]
    if (!option || option.disabled) return
    setUncontrolledValue(option.value)
    setActiveIndex(index)
    setOpen(false)
    const native = nativeRef.current
    if (!native) return
    native.value = option.value
    native.dispatchEvent(new Event('change', { bubbles: true }))
  }

  function move(direction: 1 | -1, from = activeIndex) {
    if (options.length === 0) return
    let index = from
    do { index = (index + direction + options.length) % options.length } while (options[index]?.disabled && index !== from)
    setActiveIndex(index)
  }

  function onTriggerKeyDown(event: React.KeyboardEvent<HTMLButtonElement>) {
    selectProps.onKeyDown?.(event as unknown as React.KeyboardEvent<HTMLSelectElement>)
    if (event.defaultPrevented || disabled) return
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); setOpen(true); move(event.key === 'ArrowDown' ? 1 : -1) }
    else if (event.key === 'Home' || event.key === 'End') { event.preventDefault(); setOpen(true); setActiveIndex(event.key === 'Home' ? 0 : Math.max(0, options.length - 1)) }
    else if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); if (open) commit(activeIndex); else setOpen(true) }
    else if (event.key === 'Escape') { event.preventDefault(); setOpen(false) }
  }

  const { 'aria-label': ariaLabel, 'aria-labelledby': ariaLabelledBy, 'aria-describedby': ariaDescribedBy, ...nativeProps } = selectProps

  return <div ref={rootRef} className="relative min-w-0">
    <select ref={nativeRef} name={name} form={form} required={required} disabled={disabled} value={selectedValue} onChange={onChange} {...nativeProps} tabIndex={-1} aria-hidden="true" className="sr-only">{children}</select>
    <button id={id} type="button" disabled={disabled} aria-label={ariaLabel} aria-labelledby={ariaLabelledBy} aria-describedby={ariaDescribedBy} aria-required={required || undefined} aria-haspopup="listbox" aria-expanded={open} aria-controls={listboxId} aria-activedescendant={open && activeIndex >= 0 ? `${listboxId}-${activeIndex}` : undefined} className={cn('flex h-10 min-h-11 w-full min-w-0 items-center justify-between gap-2 rounded-lg border bg-card px-3 text-sm outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50', className)} onClick={() => setOpen(current => !current)} onKeyDown={onTriggerKeyDown}>
      <span className="min-w-0 truncate text-left">{selectedOption?.label}</span><ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform', open && 'rotate-180')} aria-hidden="true" />
    </button>
    {open && options.length > 0 && <div id={listboxId} role="listbox" aria-label={ariaLabel} className="absolute left-0 top-[calc(100%+0.35rem)] z-50 max-h-64 w-full min-w-40 overflow-auto rounded-lg border bg-popover p-1 text-popover-foreground shadow-xl">
      {options.map((option, index) => <button key={`${option.value}-${index}`} id={`${listboxId}-${index}`} type="button" role="option" aria-selected={index === selectedIndex} disabled={option.disabled} className={cn('flex min-h-9 w-full items-center rounded-md px-2.5 py-2 text-left text-sm outline-none', index === activeIndex && 'bg-accent text-accent-foreground', option.disabled && 'cursor-not-allowed opacity-45')} onMouseEnter={() => setActiveIndex(index)} onClick={() => commit(index)}>{option.label}</button>)}
    </div>}
  </div>
}
