import { CheckCircle2, CircleAlert, Clock3, GitBranch, ShieldAlert, Wrench } from 'lucide-react'
import type { ApprovalRequest, SubagentActivity, ToolActivity } from '../../lib/chat-contract'
import { Button } from '../ui/button'

function StatusIcon({ status }: { status: string }) { return status === 'running' ? <Clock3 size={15} className="text-primary" /> : status === 'error' ? <CircleAlert size={15} className="text-destructive" /> : <CheckCircle2 size={15} className="text-emerald-400" /> }

export function ActivityCards({ tools, subagents, approvals, onApproval }: { tools: ToolActivity[]; subagents: SubagentActivity[]; approvals: ApprovalRequest[]; onApproval?: (id: string, decision: 'approved' | 'denied') => void }) {
  if (!tools.length && !subagents.length && !approvals.length) return null
  return <div className="mb-4 space-y-2" aria-label="Hermes activity">
    {tools.map((tool) => <div key={tool.id} className="activity-card"><Wrench size={14} className="text-muted-foreground" /><StatusIcon status={tool.status} /><span className="font-medium">{tool.name}</span><span className="ml-auto text-muted-foreground">{tool.status}</span></div>)}
    {subagents.map((agent) => <div key={agent.id} className="activity-card"><GitBranch size={14} className="text-muted-foreground" /><StatusIcon status={agent.status} /><span className="font-medium">{agent.name}</span><span className="min-w-0 flex-1 truncate text-muted-foreground">{agent.task || 'Delegated task'}</span></div>)}
    {approvals.map((approval) => <div key={approval.id} className="approval-card"><ShieldAlert size={17} className="shrink-0 text-amber-300" /><div className="min-w-0 flex-1"><p className="font-medium">Approval required: {approval.name}</p>{approval.command && <code className="mt-1 block whitespace-pre-wrap break-words text-xs text-muted-foreground">{approval.command}</code>}{approval.reason && <p className="mt-1 text-xs text-muted-foreground">{approval.reason}</p>}</div>{approval.status !== 'pending' && <span className="text-xs capitalize text-muted-foreground">{approval.status}</span>}{approval.status === 'pending' && onApproval && <div className="flex shrink-0 gap-1"><Button type="button" size="sm" variant="outline" onClick={() => onApproval(approval.id, 'approved')} aria-label={`Approve ${approval.name}`}>Approve</Button><Button type="button" size="sm" variant="ghost" onClick={() => onApproval(approval.id, 'denied')} aria-label={`Deny ${approval.name}`}>Deny</Button></div>}{approval.status === 'pending' && !onApproval && <span className="text-xs text-muted-foreground">Waiting for Runs API</span>}</div>)}
  </div>
}
