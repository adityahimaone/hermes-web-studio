import { useQuery } from '@tanstack/react-query'
import { CircleAlert, LoaderCircle } from 'lucide-react'
import { getHermesHealth } from '../../lib/api-client'
import { Badge } from '../ui/badge'

export function ConnectionStatus() {
  const health = useQuery({
    queryKey: ['hermes-health'],
    queryFn: ({ signal }) => getHermesHealth(signal),
    refetchInterval: 30_000,
  })

  if (health.isPending) return <Badge className="h-6 gap-1 px-2 text-[10px] text-muted-foreground"><LoaderCircle size={10} className="animate-spin" /> Checking</Badge>
  if (health.data?.reachable) return <Badge className="h-6 gap-1.5 border-emerald-500/25 bg-emerald-500/10 px-2 text-[10px] font-medium text-emerald-300"><span className="size-1.5 rounded-full bg-emerald-400 shadow-[0_0_6px_#34d399]" /> Connected</Badge>
  return <Badge title={health.data?.message || health.error?.message} className="h-6 gap-1.5 border-amber-500/25 bg-amber-500/10 px-2 text-[10px] font-medium text-amber-300"><CircleAlert size={10} /> Offline</Badge>
}


