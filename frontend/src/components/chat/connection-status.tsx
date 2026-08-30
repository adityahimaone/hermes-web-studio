import { useQuery } from '@tanstack/react-query'
import { CircleAlert, LoaderCircle, Radio } from 'lucide-react'
import { getHermesHealth } from '../../lib/api-client'
import { Badge } from '../ui/badge'

export function ConnectionStatus() {
  const health = useQuery({
    queryKey: ['hermes-health'],
    queryFn: ({ signal }) => getHermesHealth(signal),
    refetchInterval: 30_000,
  })

  if (health.isPending) return <Badge className="gap-1.5 text-muted-foreground"><LoaderCircle size={11} className="animate-spin" /> Checking Hermes</Badge>
  if (health.data?.reachable) return <Badge className="gap-1.5 border-emerald-500/25 bg-emerald-500/10 text-emerald-300"><Radio size={11} /> Hermes connected</Badge>
  return <Badge title={health.data?.message || health.error?.message} className="gap-1.5 border-amber-500/25 bg-amber-500/10 text-amber-300"><CircleAlert size={11} /> Hermes offline</Badge>
}

