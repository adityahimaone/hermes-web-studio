export type WorkspaceEntry = { name: string; path: string; type: 'file' | 'directory' | 'symlink'; size?: number; modified_at?: string; mime?: string }
export type WorkspacePreview = WorkspaceEntry & { content?: string; editable: boolean; binary: boolean }
export type GitStatus = { available: boolean; root?: string; branch?: string; entries?: { path: string; status: string }[]; error?: string }

export function breadcrumbs(path: string) {
  const clean = path.replaceAll('\\', '/').split('/').filter(Boolean)
  return [{ label: 'Workspace', path: '.' }, ...clean.map((part, index) => ({ label: part, path: clean.slice(0, index + 1).join('/') }))]
}
