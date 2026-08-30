import { useMemo } from 'react'

type Node = { id: string; label: string }
type Edge = { from: string; to: string }

function parseDiagram(source: string) {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const known = new Set<string>()
  const addNode = (id: string, label = id) => {
    const cleanId = id.replace(/[^a-zA-Z0-9_-]/g, '')
    if (!cleanId || known.has(cleanId)) return cleanId
    known.add(cleanId)
    nodes.push({ id: cleanId, label: label.replace(/^[\[({]|[\])}]$/g, '').slice(0, 80) })
    return cleanId
  }

  for (const line of source.split('\n')) {
    const match = line.match(/^\s*([\w-]+)(?:\s*\[[^\]]*\]|\s*\(([^)]*)\)|\s*\{([^}]*)\})?\s*[-=]+>\s*([\w-]+)(?:\s*\[([^\]]*)\]|\s*\(([^)]*)\)|\s*\{([^}]*)\})?/) 
    if (!match) continue
    const from = addNode(match[1])
    const to = addNode(match[4], match[5] || match[6] || match[7] || match[4])
    if (from && to) edges.push({ from, to })
  }
  return { nodes, edges }
}

export function MermaidDiagram({ source }: { source: string }) {
  const diagram = useMemo(() => parseDiagram(source), [source])
  if (diagram.nodes.length < 2 || diagram.edges.length === 0) {
    return <pre className="mermaid-fallback" role="img" aria-label="Mermaid diagram source"><code>{source}</code></pre>
  }
  const width = 320
  const height = Math.max(120, diagram.nodes.length * 64 + 24)
  const positions = new Map(diagram.nodes.map((node, index) => [node.id, { x: 16, y: index * 64 + 16 }]))
  return (
    <figure className="mermaid-diagram" aria-label="Mermaid diagram">
      <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Flow diagram">
        <defs><marker id="hermes-arrow" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto"><path d="M0,0 L8,4 L0,8 z" fill="currentColor" /></marker></defs>
        {diagram.edges.map((edge, index) => {
          const from = positions.get(edge.from)
          const to = positions.get(edge.to)
          if (!from || !to) return null
          return <line key={`${edge.from}-${edge.to}-${index}`} x1={from.x + 288} y1={from.y + 20} x2={to.x + 288} y2={to.y + 20} markerEnd="url(#hermes-arrow)" />
        })}
        {diagram.nodes.map((node) => {
          const position = positions.get(node.id)!
          return <g key={node.id}><rect x={position.x} y={position.y} width="288" height="40" rx="8" /><text x={position.x + 12} y={position.y + 25}>{node.label}</text></g>
        })}
      </svg>
      <figcaption>Flow diagram · safe native preview</figcaption>
    </figure>
  )
}

export function splitMermaidBlocks(content: string) {
  const parts: Array<{ kind: 'markdown' | 'mermaid'; content: string }> = []
  const pattern = /```mermaid\s*\n([\s\S]*?)```/gi
  let cursor = 0
  for (const match of content.matchAll(pattern)) {
    const index = match.index ?? 0
    if (index > cursor) parts.push({ kind: 'markdown', content: content.slice(cursor, index) })
    parts.push({ kind: 'mermaid', content: match[1].trim() })
    cursor = index + match[0].length
  }
  if (cursor < content.length) parts.push({ kind: 'markdown', content: content.slice(cursor) })
  return parts.length ? parts : [{ kind: 'markdown' as const, content }]
}
