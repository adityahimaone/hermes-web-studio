import { Activity, BarChart3, Coins, RefreshCw, Server, Waypoints } from 'lucide-react'
import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { Button } from '../ui/button'
import { getInsights, type InsightsResponse } from '../../lib/insights-client'

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)
const formatDate = (value?: string | null) => value ? new Date(value).toLocaleString() : 'No recorded activity'

export function InsightsView() {
  const [data, setData] = useState<InsightsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const refresh = useCallback(async (signal?: AbortSignal) => {
    setLoading(true)
    try { setData(await getInsights(signal)); setError('') }
    catch (err) { if (err instanceof Error && err.name === 'AbortError') return; setError(err instanceof Error ? err.message : 'Insights are unavailable.') }
    finally { if (!signal?.aborted) setLoading(false) }
  }, [])
  useEffect(() => { const controller = new AbortController(); void refresh(controller.signal); return () => controller.abort() }, [refresh])

  return <section className="mx-auto w-full max-w-4xl p-5 sm:p-8" aria-labelledby="insights-title">
    <div className="mb-8 flex flex-wrap items-start justify-between gap-4"><div><p className="text-[11px] uppercase tracking-[0.18em] text-primary">Operational view</p><h1 id="insights-title" className="mt-2 text-2xl font-semibold">Insights</h1><p className="mt-1 max-w-xl text-sm text-muted-foreground">Recorded usage and provider history from the server-owned Hermes session state.</p></div><Button type="button" variant="outline" size="sm" onClick={() => void refresh()} disabled={loading}><RefreshCw size={14} aria-hidden="true" />{loading ? 'Reading state' : 'Refresh'}</Button></div>
    {error && <div className="mb-5 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-destructive/40 bg-destructive/5 p-4" role="alert"><p className="text-sm text-red-300">{error}</p><Button type="button" variant="outline" size="sm" onClick={() => void refresh()}>Try again</Button></div>}
    {loading && !data ? <p role="status" className="text-sm text-muted-foreground">Reading recorded Hermes state...</p> : data && <>
      <div className="mb-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4"><InsightMetric icon={<Activity size={16} />} label="Sessions" value={formatNumber(data.summary.sessions)} detail={`${formatNumber(data.summary.messages)} recorded messages`} /><InsightMetric icon={<Waypoints size={16} />} label="Conversation split" value={`${formatNumber(data.summary.user_messages)} / ${formatNumber(data.summary.assistant_messages)}`} detail="User / Hermes messages" /><InsightMetric icon={<BarChart3 size={16} />} label="Tokens" value={data.usage.available ? formatNumber(data.usage.total_tokens) : 'Not recorded'} detail={data.usage.available ? `${formatNumber(data.usage.input_tokens)} in · ${formatNumber(data.usage.output_tokens)} out` : 'No usage events in session state'} /><InsightMetric icon={<Server size={16} />} label="State sync" value={data.synchronization.status} detail={`${formatNumber(data.synchronization.sessions_read)}/${formatNumber(data.synchronization.sessions_scanned)} sessions read`} /></div>
      <div className="grid gap-4 lg:grid-cols-[1.25fr_0.75fr]"><section className="rounded-lg border bg-card p-4" aria-labelledby="provider-history-title"><div className="flex items-start justify-between gap-3"><div><h2 id="provider-history-title" className="text-sm font-medium">Provider history</h2><p className="mt-1 text-xs text-muted-foreground">Only provider and model values persisted with a session are shown.</p></div><Server size={16} className="text-muted-foreground" aria-hidden="true" /></div>{data.provider_history.length === 0 ? <p className="mt-6 rounded border border-dashed p-5 text-sm text-muted-foreground">No provider history is recorded yet.</p> : <div className="mt-4 space-y-2">{data.provider_history.map(item => <article key={`${item.provider}-${item.model || 'default'}`} className="flex flex-wrap items-center gap-3 rounded border px-3 py-3"><div className="min-w-0 flex-1"><p className="truncate text-sm">{item.provider}</p><p className="truncate text-xs text-muted-foreground">{item.model || 'Model not recorded'} · {formatNumber(item.sessions)} {item.sessions === 1 ? 'session' : 'sessions'}</p></div><p className="text-right text-xs text-muted-foreground">{formatNumber(item.messages)} messages{item.usage.available && <><br />{formatNumber(item.usage.total_tokens)} tokens</>}</p></article>)}</div>}</section><section className="rounded-lg border bg-card p-4" aria-labelledby="state-summary-title"><div className="flex items-start justify-between gap-3"><div><h2 id="state-summary-title" className="text-sm font-medium">State summary</h2><p className="mt-1 text-xs text-muted-foreground">Read time and billing availability.</p></div><Coins size={16} className="text-muted-foreground" aria-hidden="true" /></div><dl className="mt-5 space-y-4 text-sm"><div><dt className="text-xs text-muted-foreground">Last recorded activity</dt><dd className="mt-1">{formatDate(data.synchronization.last_activity_at)}</dd></div><div><dt className="text-xs text-muted-foreground">Cost</dt><dd className="mt-1">{data.cost.available ? 'Recorded' : 'Not available'}</dd>{!data.cost.available && <p className="mt-1 text-xs text-muted-foreground">{data.cost.reason}</p>}</div><div><dt className="text-xs text-muted-foreground">Source</dt><dd className="mt-1 break-words text-xs text-muted-foreground">{data.source}</dd></div></dl></section></div>
    </>}
  </section>
}

function InsightMetric({ icon, label, value, detail }: { icon: ReactNode; label: string; value: string; detail: string }) { return <article className="rounded-lg border bg-card p-4"><div className="flex items-center gap-2 text-xs text-muted-foreground">{icon}<span>{label}</span></div><p className="mt-4 truncate text-xl font-semibold" title={value}>{value}</p><p className="mt-1 text-xs text-muted-foreground">{detail}</p></article> }
