import { useEffect, useRef, useState } from 'react'
import { type Page } from '../App'

interface CNode { id: string; description: string; type: string; company_id: string }
interface CEdge { from: string; to: string; strength: number; evidence: number }
interface GraphData { nodes: CNode[]; edges: CEdge[] }

// Simple force-directed layout: iterative spring simulation
interface Pos { x: number; y: number; vx: number; vy: number }

const W = 760
const H = 500
const RADIUS = 24

const nodeColor = (type: string) => {
  switch (type) {
    case 'decision': return 'var(--color-accent)'
    case 'outcome':  return '#22c55e'
    case 'cause':    return '#f97316'
    case 'effect':   return '#a855f7'
    default:         return 'var(--color-text-muted)'
  }
}

function runLayout(nodes: CNode[], edges: CEdge[], width: number, height: number): Map<string, Pos> {
  const pos = new Map<string, Pos>()
  nodes.forEach((n, i) => {
    const angle = (i / nodes.length) * 2 * Math.PI
    pos.set(n.id, {
      x: width / 2 + Math.cos(angle) * (Math.min(width, height) * 0.35),
      y: height / 2 + Math.sin(angle) * (Math.min(width, height) * 0.35),
      vx: 0, vy: 0,
    })
  })

  const ITER = 120
  const REPEL = 2200
  const ATTRACT = 0.04
  const IDEAL = 120
  const DAMP = 0.82

  for (let it = 0; it < ITER; it++) {
    // Repulsion
    nodes.forEach(a => nodes.forEach(b => {
      if (a.id === b.id) return
      const pa = pos.get(a.id)!; const pb = pos.get(b.id)!
      const dx = pa.x - pb.x; const dy = pa.y - pb.y
      const d2 = dx * dx + dy * dy + 0.01
      const f = REPEL / d2
      pa.vx += f * dx; pa.vy += f * dy
    }))
    // Attraction along edges
    edges.forEach(e => {
      const pa = pos.get(e.from); const pb = pos.get(e.to)
      if (!pa || !pb) return
      const dx = pb.x - pa.x; const dy = pb.y - pa.y
      const d = Math.sqrt(dx * dx + dy * dy) + 0.01
      const f = ATTRACT * (d - IDEAL)
      pa.vx += f * dx / d; pa.vy += f * dy / d
      pb.vx -= f * dx / d; pb.vy -= f * dy / d
    })
    // Integrate + damp + clamp
    nodes.forEach(n => {
      const p = pos.get(n.id)!
      p.vx *= DAMP; p.vy *= DAMP
      p.x = Math.max(RADIUS + 4, Math.min(width - RADIUS - 4, p.x + p.vx))
      p.y = Math.max(RADIUS + 4, Math.min(height - RADIUS - 4, p.y + p.vy))
    })
  }
  return pos
}

export default function CausalGraphPage({ onNavigate: _onNavigate }: { onNavigate: (p: Page) => void }) {
  const [graph, setGraph] = useState<GraphData | null>(null)
  const [loading, setLoading] = useState(true)
  const [company, setCompany] = useState('')
  const [tooltip, setTooltip] = useState<{ node: CNode; x: number; y: number } | null>(null)
  const svgRef = useRef<SVGSVGElement>(null)

  const load = (q: string) => {
    setLoading(true)
    const params = q ? `?company=${encodeURIComponent(q)}` : ''
    fetch(`/api/v1/causal-graph${params}`)
      .then(r => r.json())
      .then((d: GraphData) => { setGraph(d); setLoading(false) })
      .catch(() => setLoading(false))
  }

  useEffect(() => { load('') }, [])

  const positions = graph && graph.nodes.length > 0
    ? runLayout(graph.nodes, graph.edges ?? [], W, H)
    : null

  const topEdges = [...(graph?.edges ?? [])]
    .sort((a, b) => b.evidence - a.evidence)
    .slice(0, 5)

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <div style={{ marginBottom: '24px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Grafo Causal</h1>
        <p style={{ margin: '6px 0 0', fontSize: '13px', color: 'var(--color-text-muted)' }}>
          Relações causa→efeito aprendidas ao longo das simulações. Quanto mais evidências, mais espessa a aresta.
        </p>
      </div>

      {/* Filter */}
      <div style={{ display: 'flex', gap: '10px', marginBottom: '20px' }}>
        <input
          value={company}
          onChange={e => setCompany(e.target.value)}
          placeholder="Filtrar por empresa ou departamento..."
          style={{ flex: 1, padding: '8px 12px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'var(--color-surface)', color: 'var(--color-text)', fontSize: '13px' }}
          onKeyDown={e => e.key === 'Enter' && load(company)}
        />
        <button onClick={() => load(company)} style={{ padding: '8px 18px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>
          Filtrar
        </button>
        {company && <button onClick={() => { setCompany(''); load('') }} style={{ padding: '8px 14px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '13px', cursor: 'pointer' }}>Limpar</button>}
      </div>

      {/* Legend */}
      <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '16px' }}>
        {[['decision', 'Decisão'], ['outcome', 'Resultado'], ['cause', 'Causa'], ['effect', 'Efeito']].map(([type, label]) => (
          <div key={type} style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
            <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: nodeColor(type) }} />
            {label}
          </div>
        ))}
      </div>

      {loading ? (
        <div style={{ padding: '80px', textAlign: 'center', color: 'var(--color-text-muted)' }}>Carregando grafo...</div>
      ) : !graph || graph.nodes.length === 0 ? (
        <div style={{ padding: '60px', textAlign: 'center', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)' }}>
          <div style={{ fontSize: '32px', marginBottom: '12px' }}>◈</div>
          <div style={{ fontWeight: '600', color: 'var(--color-text)', marginBottom: '8px' }}>Grafo ainda vazio</div>
          <div style={{ fontSize: '13px', color: 'var(--color-text-muted)', maxWidth: '400px', margin: '0 auto' }}>
            O grafo causal é construído automaticamente ao longo das simulações. Execute algumas simulações e forneça feedback para ver as relações emergindo.
          </div>
        </div>
      ) : (
        <>
          <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', overflow: 'hidden', position: 'relative' }}>
            <svg ref={svgRef} width="100%" viewBox={`0 0 ${W} ${H}`} style={{ display: 'block' }}>
              <defs>
                <marker id="arrow" markerWidth="8" markerHeight="8" refX="20" refY="3" orient="auto">
                  <path d="M0,0 L0,6 L8,3 z" fill="var(--color-text-muted)" opacity="0.5" />
                </marker>
              </defs>

              {/* Edges */}
              {(graph.edges ?? []).map((e, i) => {
                const pa = positions!.get(e.from)
                const pb = positions!.get(e.to)
                if (!pa || !pb) return null
                const strokeW = Math.max(1, Math.min(5, e.evidence * 0.5))
                const opacity = 0.2 + Math.min(0.6, e.strength * 0.7)
                return (
                  <line key={i}
                    x1={pa.x} y1={pa.y} x2={pb.x} y2={pb.y}
                    stroke="var(--color-text-muted)"
                    strokeWidth={strokeW}
                    opacity={opacity}
                    markerEnd="url(#arrow)"
                  />
                )
              })}

              {/* Nodes */}
              {graph.nodes.map(n => {
                const p = positions!.get(n.id)!
                const color = nodeColor(n.type)
                return (
                  <g key={n.id} style={{ cursor: 'pointer' }}
                    onMouseEnter={ev => {
                      const rect = svgRef.current?.getBoundingClientRect()
                      if (!rect) return
                      setTooltip({ node: n, x: ev.clientX - rect.left, y: ev.clientY - rect.top })
                    }}
                    onMouseLeave={() => setTooltip(null)}>
                    <circle cx={p.x} cy={p.y} r={RADIUS}
                      fill={color} fillOpacity="0.15"
                      stroke={color} strokeWidth="2" />
                    <text x={p.x} y={p.y + 4} textAnchor="middle" fontSize="10"
                      fill={color} fontWeight="600">
                      {n.type === 'decision' ? '⬡' : n.type === 'outcome' ? '◉' : n.type === 'cause' ? '◈' : '◇'}
                    </text>
                  </g>
                )
              })}
            </svg>

            {/* Tooltip */}
            {tooltip && (
              <div style={{
                position: 'absolute', left: tooltip.x + 12, top: tooltip.y - 10,
                background: 'oklch(0.15 0.01 240)', border: '1px solid var(--color-border)',
                borderRadius: '8px', padding: '10px 14px', maxWidth: '260px',
                fontSize: '12px', color: 'var(--color-text)', pointerEvents: 'none',
                boxShadow: '0 4px 16px oklch(0 0 0 / 0.4)',
              }}>
                <div style={{ fontWeight: '700', marginBottom: '4px', color: nodeColor(tooltip.node.type) }}>
                  {tooltip.node.type}
                </div>
                <div style={{ lineHeight: '1.5' }}>{tooltip.node.description}</div>
              </div>
            )}
          </div>

          {/* Top relations */}
          {topEdges.length > 0 && (
            <div style={{ marginTop: '20px' }}>
              <h3 style={{ fontSize: '13px', fontWeight: '700', color: 'var(--color-text)', margin: '0 0 12px' }}>
                5 relações com mais evidências
              </h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {topEdges.map((e, i) => {
                  const from = graph.nodes.find(n => n.id === e.from)
                  const to = graph.nodes.find(n => n.id === e.to)
                  return (
                    <div key={i} style={{ background: 'var(--color-surface)', borderRadius: '8px', border: '1px solid var(--color-border)', padding: '10px 14px', fontSize: '12px' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                        <span style={{ color: nodeColor(from?.type ?? ''), fontWeight: '600' }}>{from?.description ?? e.from}</span>
                        <span style={{ color: 'var(--color-text-muted)' }}>→</span>
                        <span style={{ color: nodeColor(to?.type ?? ''), fontWeight: '600' }}>{to?.description ?? e.to}</span>
                        <div style={{ marginLeft: 'auto', display: 'flex', gap: '10px', flexShrink: 0 }}>
                          <span style={{ color: 'var(--color-text-muted)' }}>força: <b style={{ color: 'var(--color-text)' }}>{Math.round(e.strength * 100)}%</b></span>
                          <span style={{ color: 'var(--color-text-muted)' }}>evidências: <b style={{ color: 'var(--color-text)' }}>{e.evidence}</b></span>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
