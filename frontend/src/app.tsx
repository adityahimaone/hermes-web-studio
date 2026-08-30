import { Menu, PanelRight, RotateCcw } from 'lucide-react'
import { useState } from 'react'
import { Sidebar } from './components/layout/sidebar'
import { Button } from './components/ui/button'
import { ConnectionStatus } from './components/chat/connection-status'
import { MessageList } from './components/chat/message-list'
import { Composer } from './components/chat/composer'
import { useChat } from './hooks/use-chat'
import { useWorkspace } from './hooks/use-workspace'
import { WorkspacePanel } from './components/workspace/workspace-panel'
import { IdentityControls } from './components/auth/identity-controls'
import { ControlCenter } from './components/control/control-center'

export function App() {
  const chat = useChat()
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const workspace = useWorkspace()
  const [workspaceWidth, setWorkspaceWidth] = useState(360)
  const [view, setView] = useState('chat')

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar onNewChat={() => { chat.reset(); setView('chat') }} onNavigate={setView} currentView={view} sessions={chat.sessions} activeSessionId={chat.activeSessionId} onSelectSession={chat.selectSession} onRename={chat.rename} onPin={chat.pin} onArchive={chat.archive} onDelete={chat.remove} loading={chat.sessionLoading} error={chat.sessionError} mobileOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <main className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center gap-3 border-b bg-background/70 px-3 backdrop-blur-xl sm:px-5">
          <Button variant="ghost" size="icon" className="lg:hidden" onClick={() => setMobileNavOpen(true)} aria-label="Open navigation"><Menu size={18} /></Button>
          <div className="min-w-0"><h2 className="truncate text-sm font-medium">{view === 'chat' ? (chat.sessions.find((session) => session.session_id === chat.activeSessionId)?.title || 'New Hermes conversation') : view}</h2><p className="text-[11px] text-muted-foreground">{view === 'chat' ? 'Default profile · Gateway runtime' : 'Hermes Studio control center'}</p></div>
          <div className="ml-auto flex items-center gap-2">
            <ConnectionStatus />
            <IdentityControls />
            <Button variant="ghost" size="icon" onClick={chat.reset} aria-label="Reset chat"><RotateCcw size={16} /></Button>
            <Button variant="ghost" size="icon" onClick={() => workspace.setOpen(!workspace.open)} aria-label={workspace.open ? 'Close workspace' : 'Open workspace'}><PanelRight size={17} /></Button>
          </div>
        </header>
        {view === 'chat' ? <><div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto"><MessageList messages={chat.messages} stream={chat.streamState} onEdit={chat.edit} onRetry={chat.retry} onApproval={chat.approve} /></div><Composer onSend={chat.send} onCancel={chat.cancel} isStreaming={chat.isStreaming} draft={chat.draft} onDraftChange={chat.setDraft} queuedMessages={chat.queuedMessages} /></> : <div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto"><ControlCenter view={view} /></div>}
      </main>
      <WorkspacePanel {...workspace} width={workspaceWidth} onWidthChange={value => setWorkspaceWidth(Math.max(280, Math.min(520, value)))} />
    </div>
  )
}
