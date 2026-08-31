import { Clipboard, Download, ExternalLink, File, Folder, FolderOpen, FolderPlus, GitBranch, Pencil, Plus, RefreshCw, Trash2, Upload, X } from 'lucide-react'
import { useEffect, useRef, useState, type CSSProperties } from 'react'
import { Button } from '../ui/button'
import { Textarea } from '../ui/textarea'
import { Input } from '../ui/input'
import { Dialog } from '../ui/dialog'
import { SafeMarkdown } from '../../lib/markdown'
import { breadcrumbs, type WorkspacePreview } from '../../lib/workspace-contract'
import { workspaceDownloadUrl, workspaceOpenUrl } from '../../lib/api-client'
import { cn } from '../../lib/cn'

type Props = ReturnType<typeof import('../../hooks/use-workspace').useWorkspace> & { width: number; onWidthChange: (width: number) => void }

function formatFileSize(size?: number) {
  if (size === undefined || size === null) return ''
  if (size === 0) return '(empty)'
  if (size < 1024) return `${size}b`
  return `${(size / 1024).toFixed(1)}k`
}

export function WorkspacePanel({ open, setOpen, path, entries, preview, setPreview, git, loading, error, refresh, select, save, create, rename, remove, upload, width, onWidthChange }: Props) {
  const input = useRef<HTMLInputElement>(null); const [editing, setEditing] = useState(false); const [draft, setDraft] = useState(''); const [nameDialog, setNameDialog] = useState<{ title: string; action: (value: string) => void } | null>(null); const [nameValue, setNameValue] = useState(''); const [deleteOpen, setDeleteOpen] = useState(false); const [tab, setTab] = useState<'files' | 'artifacts' | 'todos'>('files'); const [todos, setTodos] = useState<{ id: string; title: string; status: string }[]>([]); const [todosLoading, setTodosLoading] = useState(false); const [pathMessage, setPathMessage] = useState('')
  useEffect(() => { if (tab !== 'todos') return; setTodosLoading(true); void fetch('/api/control/todos').then(async response => { const data = await response.json(); if (!response.ok) throw new Error(data.message || 'Todos unavailable'); setTodos(Array.isArray(data.items) ? data.items : []) }).catch(() => setTodos([])).finally(() => setTodosLoading(false)) }, [tab])
  if (!open) return null
  const act = (label: string, fn: (value: string) => void) => { setNameValue(''); setNameDialog({ title: label, action: fn }) }
  const copyPath = async (value: string) => { try { await navigator.clipboard.writeText(value); setPathMessage(`Copied ${value}`) } catch { setPathMessage('Copy is unavailable in this browser') } }
  const openFile = (value: string) => { window.open(workspaceOpenUrl(value), '_blank', 'noopener,noreferrer') }
  const show = (item: WorkspacePreview) => item.binary ? <div className="flex flex-col gap-3 text-xs text-muted-foreground"><span>Binary file · {item.mime || 'unknown type'}</span><a className="text-primary underline" href={workspaceDownloadUrl(item.path)} download>Download file</a></div> : item.mime?.startsWith('image/') ? <img className="max-h-80 max-w-full rounded border object-contain" src={workspaceDownloadUrl(item.path)} alt={item.name} /> : item.name.endsWith('.md') ? <div className="message-markdown text-sm"><SafeMarkdown>{item.content || ''}</SafeMarkdown></div> : editing ? <Textarea autoFocus className="min-h-72 font-mono text-xs" value={draft} onChange={e => setDraft(e.target.value)} /> : <pre className="max-h-[32rem] overflow-auto whitespace-pre-wrap break-words rounded border bg-background p-3 font-mono text-xs">{item.content}</pre>

  return <aside className="workspace-panel relative flex h-full shrink-0 flex-col border-l bg-card/75 backdrop-blur-xl" style={{ '--workspace-width': `${width}px` } as CSSProperties} aria-label="Workspace">
    <div className="workspace-resizer" role="separator" aria-label="Resize workspace" tabIndex={0} aria-valuenow={width} onKeyDown={e => { if (e.key === 'ArrowLeft') onWidthChange(width - 20); if (e.key === 'ArrowRight') onWidthChange(width + 20) }} onPointerDown={e => { const start = e.clientX; const initial = width; const move = (event: PointerEvent) => onWidthChange(initial - (event.clientX - start)); const up = () => { window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', up) }; window.addEventListener('pointermove', move); window.addEventListener('pointerup', up) }} />
    
    {/* Header matching Image 3 */}
    <header className="flex h-11 items-center justify-between border-b px-3">
      <div className="flex items-center gap-2">
        <h2 className="text-[11px] font-semibold uppercase tracking-[0.14em] text-foreground">WORKSPACE</h2>
        <span className="rounded bg-muted/60 px-1.5 py-0.2 text-[9px] font-semibold uppercase text-muted-foreground">{git?.branch || 'MAIN'}</span>
      </div>
      <div className="flex items-center gap-0.5">
        <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => act('New file name', n => void create(`${path === '.' ? '' : path + '/'}${n}`, 'file'))} title="New file" aria-label="New file"><Plus size={14} /></button>
        <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => act('New folder name', n => void create(`${path === '.' ? '' : path + '/'}${n}`, 'directory'))} title="New folder" aria-label="New folder"><FolderPlus size={14} /></button>
        <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => void refresh()} title="Refresh workspace" aria-label="Refresh workspace"><RefreshCw size={13} /></button>
        <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => input.current?.click()} title="Upload file" aria-label="Upload file"><Upload size={13} /></button>
        <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => setOpen(false)} title="Close workspace" aria-label="Close workspace"><X size={15} /></button>
        <input ref={input} hidden type="file" onChange={e => { const file = e.target.files?.[0]; if (file) void upload(file); e.currentTarget.value = '' }} />
      </div>
    </header>

    {/* Compact Segmented Tabs */}
    <div className="flex gap-1 border-b px-3 py-1.5" role="tablist" aria-label="Workspace panel sections">
      {(['files', 'artifacts', 'todos'] as const).map(value => (
        <button
          key={value}
          role="tab"
          aria-selected={tab === value}
          className={cn('flex-1 rounded-md px-2 py-1 text-[11px] font-medium transition-all capitalize', tab === value ? 'bg-primary/15 font-semibold text-primary' : 'text-muted-foreground hover:bg-accent/40 hover:text-foreground')}
          onClick={() => setTab(value)}
        >
          {value === 'artifacts' ? 'Artifacts 0' : value}
        </button>
      ))}
    </div>

    {tab === 'files' && <>
      {breadcrumbs(path).length > 1 && (
        <div className="flex gap-1 overflow-x-auto border-b px-3 py-1 text-[11px]">
          {breadcrumbs(path).map(item => <button key={item.path} className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => void refresh(item.path)}>{item.label} /</button>)}
        </div>
      )}

      {error && <p className="border-b border-destructive/40 p-2 text-xs text-red-300">{error}</p>}

      {/* File Tree Rows */}
      <div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto p-1.5">
        {loading ? (
          <p className="p-2 text-xs text-muted-foreground">Loading workspace...</p>
        ) : entries.length === 0 ? (
          <p className="p-2 text-xs text-muted-foreground">This folder is empty.</p>
        ) : (
          entries.map(entry => (
            <div
              key={entry.path}
              role="button"
              tabIndex={0}
              onClick={() => void select(entry)}
              onKeyDown={e => { if (e.key === 'Enter') void select(entry) }}
              className="flex min-h-7 cursor-pointer items-center justify-between gap-1.5 rounded px-2 py-1 text-xs text-foreground/90 transition-colors hover:bg-accent/50"
            >
              <div className="flex min-w-0 items-center gap-2">
                {entry.type === 'directory' ? <Folder size={13} className="shrink-0 text-primary" /> : <File size={13} className="shrink-0 text-muted-foreground" />}
                <span className="truncate text-[11px]">{entry.name}</span>
              </div>
              <span className="shrink-0 font-mono text-[10px] text-muted-foreground/60">{formatFileSize(entry.size)}</span>
            </div>
          ))
        )}
      </div>

      {preview && (
        <section className="max-h-[50%] overflow-auto border-t bg-card/90 p-2.5">
          <div className="mb-2 flex items-center gap-1">
            <strong className="min-w-0 flex-1 truncate text-xs">{preview.name}</strong>
            <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent" onClick={() => void copyPath(preview.path)} aria-label="Copy relative path" title="Copy path"><Clipboard size={13} /></button>
            <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent" onClick={() => openFile(preview.path)} aria-label="Open file in browser" title="Open in browser"><ExternalLink size={13} /></button>
            {preview.editable && !preview.binary && (
              <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent" onClick={() => { setDraft(preview.content || ''); setEditing(!editing) }} aria-label="Edit file" title="Edit"><Pencil size={13} /></button>
            )}
            <button type="button" className="rounded p-1 text-muted-foreground hover:bg-accent" onClick={() => act('New name', n => void rename(preview.path, n))} aria-label="Rename file" title="Rename"><Pencil size={13} /></button>
            <a className="rounded p-1 text-muted-foreground hover:bg-accent" href={workspaceDownloadUrl(preview.path)} download aria-label="Download file" title="Download"><Download size={13} /></a>
            <button type="button" className="rounded p-1 text-destructive hover:bg-accent" onClick={() => setDeleteOpen(true)} aria-label="Delete file" title="Delete"><Trash2 size={13} /></button>
          </div>
          {pathMessage && <p className="mb-2 text-[10px] text-muted-foreground" role="status">{pathMessage}</p>}
          {show(preview)}
          {editing && (
            <div className="mt-2 flex gap-2">
              <Button size="sm" className="h-7 text-xs" onClick={() => { void save(draft); setEditing(false) }}>Save</Button>
              <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => setEditing(false)}>Cancel</Button>
            </div>
          )}
        </section>
      )}
    </>}

    {tab === 'artifacts' && <div className="p-4 text-xs text-muted-foreground" role="status">No artifacts are available for this workspace yet.</div>}
    {tab === 'todos' && <div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto p-3">{todosLoading ? <p role="status" className="text-xs text-muted-foreground">Loading todos...</p> : todos.length === 0 ? <p className="text-xs text-muted-foreground">No todos in the active workspace.</p> : todos.map(todo => <div key={todo.id} className="border-b py-2 text-xs"><p>{todo.title}</p><p className="text-muted-foreground">{todo.status}</p></div>)}</div>}

    <Dialog open={Boolean(nameDialog)} title={nameDialog?.title || ''} onClose={() => setNameDialog(null)}><form className="grid gap-3" onSubmit={event => { event.preventDefault(); if (nameDialog && nameValue.trim()) { nameDialog.action(nameValue.trim()); setNameDialog(null) } }}><Input autoFocus value={nameValue} onChange={event => setNameValue(event.target.value)} placeholder="Enter a name" aria-label="Name" /><div className="flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => setNameDialog(null)}>Cancel</Button><Button type="submit" disabled={!nameValue.trim()}>Continue</Button></div></form></Dialog>
    <Dialog open={deleteOpen} title="Delete item?" onClose={() => setDeleteOpen(false)}><p className="text-sm text-muted-foreground">This removes the selected workspace item.</p><div className="mt-4 flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => setDeleteOpen(false)}>Cancel</Button><Button type="button" onClick={() => { if (preview) void remove(preview.path); setDeleteOpen(false) }}>Delete item</Button></div></Dialog>
  </aside>
}
