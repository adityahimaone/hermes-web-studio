import { useCallback, useEffect, useState } from 'react'
import { createWorkspaceItem, deleteWorkspaceItem, getWorkspaceGit, getWorkspacePreview, getWorkspaceTree, renameWorkspaceItem, saveWorkspaceFile, uploadWorkspaceFile } from '../lib/api-client'
import type { GitStatus, WorkspaceEntry, WorkspacePreview } from '../lib/workspace-contract'

export function useWorkspace() {
  const [open, setOpen] = useState(false); const [path, setPath] = useState('.'); const [entries, setEntries] = useState<WorkspaceEntry[]>([]); const [preview, setPreview] = useState<WorkspacePreview | null>(null); const [git, setGit] = useState<GitStatus | null>(null); const [loading, setLoading] = useState(false); const [error, setError] = useState('')
  const refresh = useCallback(async (nextPath = path) => { setLoading(true); setError(''); try { const [tree, status] = await Promise.all([getWorkspaceTree(nextPath), getWorkspaceGit(nextPath)]); setPath(nextPath); setEntries(tree.entries); setGit(status) } catch (err) { setError(err instanceof Error ? err.message : 'Workspace could not be loaded') } finally { setLoading(false) } }, [path])
  useEffect(() => { if (open) void refresh(path) }, [open]) // eslint-disable-line react-hooks/exhaustive-deps
  const select = async (entry: WorkspaceEntry) => { if (entry.type === 'directory') { setPreview(null); await refresh(entry.path); return }; try { setError(''); setPreview(await getWorkspacePreview(entry.path)) } catch (err) { setError(err instanceof Error ? err.message : 'Preview could not be loaded') } }
  const mutate = async (fn: () => Promise<unknown>) => { await fn(); setPreview(null); await refresh(path) }
  return { open, setOpen, path, entries, preview, setPreview, git, loading, error, refresh, select, save: (content: string) => preview ? saveWorkspaceFile(preview.path, content).then(setPreview) : Promise.resolve(), create: (p: string, t: 'file' | 'directory') => mutate(() => createWorkspaceItem(p, t)), rename: (p: string, n: string) => mutate(() => renameWorkspaceItem(p, n)), remove: (p: string) => mutate(() => deleteWorkspaceItem(p)), upload: (f: File) => mutate(() => uploadWorkspaceFile(path, f)) }
}
