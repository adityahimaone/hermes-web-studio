import { BarChart3, BrainCircuit, Clock3, KanbanSquare, Menu, MessageSquareText, PanelRight, RotateCcw, Settings2, Wrench } from 'lucide-react'
import { useRef, useState } from 'react'
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
import { ExpandRailButton } from './components/layout/context-rail'

export function App() {
  const chat = useChat()
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const [customizeNavigationOpen, setCustomizeNavigationOpen] = useState(false)
  const mobileNavTrigger = useRef<HTMLButtonElement>(null)
  const workspace = useWorkspace()
  const [workspaceWidth, setWorkspaceWidth] = useState(() => Number(window.localStorage.getItem('hermes-workspace-width')) || 360)
  const [view, setView] = useState('chat')
  const [chatSessionsOpen, setChatSessionsOpen] = useState(true)
  const [controlSidebarOpen, setControlSidebarOpen] = useState(true)
  const mobileNavigation = [{ label: 'Chat', view: 'chat', icon: MessageSquareText }, { label: 'Tasks', view: 'tasks', icon: Clock3 }, { label: 'Kanban', view: 'kanban', icon: KanbanSquare }, { label: 'Insights', view: 'insights', icon: BarChart3 }, { label: 'Skills', view: 'skills', icon: Wrench }, { label: 'Memory', view: 'memory', icon: BrainCircuit }, { label: 'Settings', view: 'settings', icon: Settings2 }]
  const handleNavigate = (nextView: string) => {
    if (nextView === 'chat' && view === 'chat') {
      setChatSessionsOpen(open => !open)
      return
    }
    setView(nextView)
    if (nextView === 'chat') setChatSessionsOpen(true)
    else setControlSidebarOpen(true)
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar onNavigate={handleNavigate} currentView={view} mobileOpen={mobileNavOpen} onClose={() => { setMobileNavOpen(false); requestAnimationFrame(() => mobileNavTrigger.current?.focus()) }} customizeOpen={customizeNavigationOpen} onCustomizeOpenChange={setCustomizeNavigationOpen} />
      <main id="main-content" className="flex min-w-0 flex-1 flex-col" aria-label="Hermes Studio workspace">
        <header data-testid="titlebar" className="flex h-16 shrink-0 items-center justify-between gap-3 border-b bg-background/80 px-3 backdrop-blur-xl sm:px-4">
          <div className="flex items-center gap-2">
            <Button ref={mobileNavTrigger} variant="ghost" size="icon" className="size-8 lg:hidden" onClick={() => setMobileNavOpen(true)} aria-label="Open navigation"><Menu size={17} /></Button>
            {view !== 'chat' && <h2 className="truncate text-sm font-semibold capitalize">{view}</h2>}
          </div>
          {view === 'chat' && (
            <div className="flex min-w-0 max-w-[60%] items-center justify-center gap-2 rounded-full border border-border/70 bg-card/60 px-3 py-1 shadow-sm">
              <span className="flex size-2 rounded-full bg-cyan-400 shadow-[0_0_8px_#22d3ee]" aria-hidden="true" />
              <h2 className="truncate text-xs font-semibold tracking-tight text-foreground">{chat.sessions.find((session) => session.session_id === chat.activeSessionId)?.title || 'New Hermes conversation'}</h2>
              {chat.messages.length > 0 && <span className="rounded-full bg-muted px-1.5 py-0.2 text-[10px] font-semibold text-muted-foreground tabular-nums">{chat.messages.length}</span>}
              <button type="button" onClick={chat.reset} className="rounded p-0.5 text-muted-foreground hover:bg-accent hover:text-foreground" aria-label="Reset chat" title="Reset conversation"><RotateCcw size={12} /></button>
            </div>
          )}
          <div className="flex items-center gap-1.5">
            <ConnectionStatus />
            <IdentityControls showAuthForm={false} />
            <Button variant="ghost" size="icon" className="size-8 rounded-lg text-muted-foreground hover:bg-accent/60 hover:text-foreground" onClick={() => workspace.setOpen(!workspace.open)} aria-label={workspace.open ? 'Close workspace' : 'Open workspace'} title={workspace.open ? 'Close workspace' : 'Open workspace'}><PanelRight size={16} /></Button>
          </div>
        </header>
        {view === 'chat' ? <div className="relative flex min-h-0 flex-1">{chatSessionsOpen && <SessionRail sessions={chat.sessions} activeSessionId={chat.activeSessionId} onSelectSession={chat.selectSession} onSearch={chat.searchSessions} onRename={chat.rename} onPin={chat.pin} onArchive={chat.archive} onDelete={chat.remove} onDuplicate={chat.duplicate} onNewChat={() => { chat.reset(); setView('chat'); setChatSessionsOpen(true) }} loading={chat.sessionLoading} error={chat.sessionError} onToggle={() => setChatSessionsOpen(false)} />}{!chatSessionsOpen && <ExpandRailButton label="Chat" onClick={() => setChatSessionsOpen(true)} />}<div className="flex min-w-0 flex-1 flex-col"><div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto"><MessageList messages={chat.messages} stream={chat.streamState} onEdit={chat.edit} onRetry={chat.retry} onApproval={chat.approve} /></div><Composer onSend={chat.send} onCommand={command => { if (command === '/clear') chat.reset(); else if (command === '/help') chat.setDraft('Try /clear to reset this conversation, or write a message for Hermes.'); else chat.send(command) }} onCancel={chat.cancel} onRemoveQueued={chat.removeQueued} isStreaming={chat.isStreaming} draft={chat.draft} onDraftChange={chat.setDraft} queuedMessages={chat.queuedMessages} contextUsage={chat.streamState.usage} workspacePath={workspace.path} onWorkspaceOpen={() => workspace.setOpen(true)} /></div></div> : <div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto"><ControlCenter view={view} sidebarOpen={controlSidebarOpen} onToggleSidebar={() => setControlSidebarOpen(open => !open)} onCustomizeNavigation={() => setCustomizeNavigationOpen(true)} /></div>}
        <nav data-testid="mobile-bottom-nav" className="mobile-bottom-nav lg:hidden" aria-label="Mobile navigation">{mobileNavigation.map(({ label, view: target, icon: Icon }) => <Button key={target} type="button" variant={view === target ? 'default' : 'ghost'} className="mobile-bottom-nav__item" onClick={() => handleNavigate(target)} aria-label={label} aria-current={view === target ? 'page' : undefined}><Icon aria-hidden="true" size={17} /><span>{label}</span></Button>)}</nav>
      </main>
      <WorkspacePanel {...workspace} width={workspaceWidth} onWidthChange={value => { const next = Math.max(280, Math.min(520, value)); setWorkspaceWidth(next); window.localStorage.setItem('hermes-workspace-width', String(next)) }} />
    </div>
  )
}
