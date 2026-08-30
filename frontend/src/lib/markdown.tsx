import { useState, type ComponentPropsWithoutRef } from 'react'
import Markdown from 'react-markdown'
import { Check, Copy, ExternalLink } from 'lucide-react'
import { Button } from '../components/ui/button'

function Code({ children, className, ...props }: ComponentPropsWithoutRef<'code'>) {
  const [copied, setCopied] = useState(false)
  const value = String(children).replace(/\n$/, '')
  if (!className) return <code className="rounded bg-accent px-1.5 py-0.5 text-[0.9em]" {...props}>{children}</code>
  async function copy() {
    if (!navigator.clipboard) return
    await navigator.clipboard.writeText(value)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1400)
  }
  return <div className="code-block"><div className="code-block__toolbar"><span>{className.replace('language-', '') || 'code'}</span><Button type="button" variant="ghost" size="sm" onClick={copy} aria-label="Copy code" className="h-7 gap-1.5 px-2 text-xs">{copied ? <Check size={13} /> : <Copy size={13} />}{copied ? 'Copied' : 'Copy'}</Button></div><pre><code className={className} {...props}>{children}</code></pre></div>
}

export function SafeMarkdown({ children }: { children: string }) {
  return <Markdown components={{
    code: Code,
    a: ({ href, children: label, ...props }) => {
      const safe = href?.startsWith('https://') || href?.startsWith('http://') || href?.startsWith('mailto:')
      if (!safe) return <span className="text-muted-foreground" title="Link blocked">{label}</span>
      return <a href={href} target="_blank" rel="noreferrer noopener" {...props}>{label}<ExternalLink aria-hidden size={12} className="ml-1 inline" /></a>
    },
  }}>{children}</Markdown>
}
