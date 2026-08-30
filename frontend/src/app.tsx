import { BrainCircuit, Clock3, Menu, MessageSquareText, PanelRight, RotateCcw, Settings2, Wrench } from 'lucide-react'
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
import { SessionRail } from './components/chat/session-rail'

export function App() {
  const chat = useChat()
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const workspace = useWorkspace()
  const [workspaceWidth, setWorkspaceWidth] = useState(() => Number(window.localStorage.getItem('hermes-workspace-width')) || 360)
  const [view, setView] = useState('chat')
  const [chatSessionsOpen, setChatSessionsOpen] = useState(true)
  const mobileNavigation = [{ label: 'Chat', view: 'chat', icon: MessageSquareText }, { label: 'Tasks', view: 'tasks', icon: Clock3 }, { label: 'Skills', view: 'skills', icon: Wrench }, { label: 'Memory', view: 'memory', icon: BrainCircuit }, { label: 'Settings', view: 'settings', icon: Settings2 }]
  const handleNavigate = (nextView: string) => {
    if (nextView === 'chat' && view === 'chat') {
      setChatSessionsOpen(open => !open)
      return
    }
    setView(nextView)
    if (nextView === 'chat') setChatSessionsOpen(true)
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar onNewChat={() => { chat.reset(); setView('chat'); setChatSessionsOpen(true) }} onNavigate={handleNavigate} currentView={view} mobileOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <main id="main-content" className="flex min-w-0 flex-1 flex-col" aria-label="Hermes Studio workspace">
        <header data-testid="titlebar" className="flex h-14 shrink-0 items-center gap-3 border-b bg-background/70 px-3 backdrop-blur-xl sm:px-5">
          <Button variant="ghost" size="icon" className="lg:hidden" onClick={() => setMobileNavOpen(true)} aria-label="Open navigation"><Menu size={18} /></Button>
          <div className="min-w-0"><h2 className="truncate text-sm font-medium">{view === 'chat' ? (chat.sessions.find((session) => session.session_id === chat.activeSessionId)?.title || 'New Hermes conversation') : view}</h2><p className="text-[11px] text-muted-foreground">{view === 'chat' ? 'Default profile · Gateway runtime' : 'Hermes Studio control center'}</p></div>
          <div className="ml-auto flex items-center gap-2">
            <ConnectionStatus />
            <IdentityControls />
            <Button variant="ghost" size="icon" onClick={chat.reset} aria-label="Reset chat"><RotateCcw size={16} /></Button>
            <Button variant="ghost" size="icon" onClick={() => workspace.setOpen(!workspace.open)} aria-label={workspace.open ? 'Close workspace' : 'Open workspace'}><PanelRight size={17} /></Button>
          </div>
        </header>
        {view === 'chat' ? <div className="flex min-h-0 flex-1">{chatSessionsOpen && <SessionRail sessions={chat.sessions} activeSessionId={chat.activeSessionId} onSelectSession={chat.selectSession} onSearch={chat.searchSessions} onRename={chat.rename} onPin={chat.pin} onArchive={chat.archive} onDelete={chat.remove} onDuplicate={chat.duplicate} loading={chat.sessionLoading} error={chat.sessionError} onToggle={() => setChatSessionsOpen(false)} />}<div className="flex min-w-0 flex-1 flex-col"><div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto"><MessageList messages={chat.messages} stream={chat.streamState} onEdit={chat.edit} onRetry={chat.retry} onApproval={chat.approve} /></div><Composer onSend={chat.send} onCommand={command => { if (command === '/clear') chat.reset(); else if (command === '/help') chat.setDraft('Try /clear to reset this conversation, or write a message for Hermes.') }} onCancel={chat.cancel} onRemoveQueued={chat.removeQueued} isStreaming={chat.isStreaming} draft={chat.draft} onDraftChange={chat.setDraft} queuedMessages={chat.queuedMessages} contextUsage={chat.streamState.usage} workspacePath={workspace.path} onWorkspaceOpen={() => workspace.setOpen(true)} /></div></div> : <div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto"><ControlCenter view={view} /></div>}
        <nav data-testid="mobile-bottom-nav" className="mobile-bottom-nav lg:hidden" aria-label="Mobile navigation">{mobileNavigation.map(({ label, view: target, icon: Icon }) => <Button key={target} type="button" variant={view === target ? 'default' : 'ghost'} className="mobile-bottom-nav__item" onClick={() => handleNavigate(target)} aria-label={label} aria-current={view === target ? 'page' : undefined}><Icon aria-hidden="true" size={17} /><span>{label}</span></Button>)}</nav>
      </main>
      <WorkspacePanel {...workspace} width={workspaceWidth} onWidthChange={value => { const next = Math.max(280, Math.min(520, value)); setWorkspaceWidth(next); window.localStorage.setItem('hermes-workspace-width', String(next)) }} />
    </div>
  )
}
