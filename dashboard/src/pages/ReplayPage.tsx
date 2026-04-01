import { useEffect, useState, useCallback } from 'react'
import { type Page } from '../App'

interface RoundRow {
  id: string
  simulation_id: string
  round_number: number
  agent_id: string
  agent_type: string
  action_text: string
  tension_level: number
  fracture_proposed: boolean
  fracture_accepted?: boolean
  new_rule_json?: string
  tokens_used: number
}

interface TensionPoint {
  round: number
  avg_tension: number
  fracture_count: number
}

interface RoundGroup {
  round: number
  actions: RoundRow[]
  avgTension: number
  fractureCount: number
}

const SVG_W = 640
const SVG_H = 80
const PAD = { top: 8, right: 12, bottom: 20, left: 36 }
const PLOT_W = SVG_W - PAD.left - PAD.right
const PLOT_H = SVG_H - PAD.top - PAD.bottom

export default function ReplayPage({ simId, onNavigate }: { simId: string; onNavigate: (p: Page, simId?: string) => void }) {
  const [rounds, setRounds] = useState<RoundGroup[]>([])
  const [tensionPoints, setTensionPoints] = useState<TensionPoint[]>([])
  const [currentRound, setCurrentRound] = useState(1)
  const [loading, setLoading] = useState(true)
  const [playing, setPlaying] = useState(false)

  useEffect(() => {
    Promise.all([
      fetch(`/api/v1/simulations/${simId}/rounds`).then(r => r.json()),
      fetch(`/api/v1/simulations/${simId}/events`).then(r => r.json()),
    ]).then(([rows, tp]: [RoundRow[], TensionPoint[]]) => {
      const grouped = groupByRound(rows ?? [])
      setRounds(grouped)
      setTensionPoints(tp ?? [])
      setCurrentRound(grouped[0]?.round ?? 1)
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [simId])

  // Auto-play
  useEffect(() => {
    if (!playing || rounds.length === 0) return
    const last = rounds[rounds.length - 1].round
    if (currentRound >= last) { setPlaying(false); return }
    const t = setTimeout(() => setCurrentRound(r => r + 1), 900)
    return () => clearTimeout(t)
  }, [playing, currentRound, rounds])

  const go = useCallback((delta: number) => {
    setCurrentRound(r => {
      const min = rounds[0]?.round ?? 1
      const max = rounds[rounds.length - 1]?.round ?? 1
      return Math.max(min, Math.min(max, r + delta))
    })
  }, [rounds])

  const group = rounds.find(g => g.round === currentRound)
  const maxRound = rounds[rounds.length - 1]?.round ?? 1
  const minRound = rounds[0]?.round ?? 1
  const disruptors = group?.actions.filter(a => a.agent_type === 'disruptor') ?? []
  const conformists = group?.actions.filter(a => a.agent_type === 'conformist') ?? []
  const fractures = group?.actions.filter(a => a.fracture_proposed) ?? []

  if (loading) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--color-text-muted)' }}>
      Carregando replay...
    </div>
  )

  if (rounds.length === 0) return (
    <div style={{ padding: '32px' }}>
      <button onClick={() => onNavigate('result', simId)} style={backBtn}>← Voltar ao relatório</button>
      <div style={{ marginTop: '24px', padding: '48px', textAlign: 'center', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', color: 'var(--color-text-muted)' }}>
        Dados de rodada não disponíveis para esta simulação
      </div>
    </div>
  )

  return (
    <div style={{ padding: '24px 32px', maxWidth: '1100px' }}>
      <button onClick={() => onNavigate('result', simId)} style={backBtn}>← Voltar ao relatório</button>

      <div style={{ margin: '16px 0 20px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '12px' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Replay da Simulação</h1>
          <p style={{ margin: '4px 0 0', fontSize: '13px', color: 'var(--color-text-muted)' }}>
            Rodada por rodada — {maxRound} rodadas · {rounds.reduce((s, g) => s + g.actions.length, 0)} ações de agentes
          </p>
        </div>
        {/* Playback controls */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <button onClick={() => go(-1)} disabled={currentRound <= minRound} style={ctrlBtn}>‹</button>
          <button onClick={() => setPlaying(p => !p)} style={{ ...ctrlBtn, minWidth: '80px', background: playing ? 'var(--color-accent)' : undefined, color: playing ? '#fff' : undefined }}>
            {playing ? '⏸ Pausar' : '▶ Play'}
          </button>
          <button onClick={() => go(1)} disabled={currentRound >= maxRound} style={ctrlBtn}>›</button>
          <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginLeft: '4px' }}>
            {currentRound} / {maxRound}
          </span>
        </div>
      </div>

      {/* Mini tension chart with clickable rounds */}
      <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '12px 16px', marginBottom: '16px' }}>
        <MiniChart
          points={tensionPoints}
          currentRound={currentRound}
          onSelectRound={setCurrentRound}
        />
        {/* Round slider */}
        <input
          type="range"
          min={minRound}
          max={maxRound}
          value={currentRound}
          onChange={e => setCurrentRound(Number(e.target.value))}
          style={{ width: '100%', marginTop: '6px', accentColor: 'var(--color-accent)' }}
        />
      </div>

      {/* Round header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '14px', flexWrap: 'wrap' }}>
        <div style={{ padding: '6px 16px', borderRadius: '20px', background: 'var(--color-accent)', color: '#fff', fontWeight: '700', fontSize: '14px' }}>
          Rodada {currentRound}
        </div>
        <Pill label="Tensão média" value={`${Math.round((group?.avgTension ?? 0) * 100)}%`} color={tensionColor(group?.avgTension ?? 0)} />
        <Pill label="Ações" value={String(group?.actions.length ?? 0)} color="var(--color-text-muted)" />
        {fractures.length > 0 && (
          <Pill label="⚡ FRACTURE POINTS" value={String(fractures.length)} color="var(--color-danger)" />
        )}
      </div>

      {/* Fracture events banner */}
      {fractures.length > 0 && (
        <div style={{ background: 'oklch(0.55 0.2 25 / 0.12)', border: '1px solid var(--color-danger)', borderRadius: '8px', padding: '12px 16px', marginBottom: '14px' }}>
          <div style={{ fontWeight: '700', fontSize: '13px', color: 'var(--color-danger)', marginBottom: '6px' }}>⚡ Fracture Points nesta rodada</div>
          {fractures.map((f, i) => {
            let ruleInfo = ''
            if (f.new_rule_json) {
              try { const p = JSON.parse(f.new_rule_json); ruleInfo = p.new_description ?? '' } catch { /* ignore */ }
            }
            return (
              <div key={i} style={{ fontSize: '12px', color: 'var(--color-text)', marginTop: '4px' }}>
                <span style={{ color: 'var(--color-danger)', fontWeight: '600' }}>{f.fracture_accepted ? '✓ Aceito' : '✗ Rejeitado'}</span>
                {ruleInfo && <span style={{ marginLeft: '8px', color: 'var(--color-text-muted)' }}>→ {ruleInfo}</span>}
              </div>
            )
          })}
        </div>
      )}

      {/* Agent actions grid */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
        <AgentList title={`Disruptores (${disruptors.length})`} actions={disruptors} accentColor="var(--color-danger)" />
        <AgentList title={`Conformistas (${conformists.length})`} actions={conformists} accentColor="var(--color-accent)" />
      </div>
    </div>
  )
}

function MiniChart({ points, currentRound, onSelectRound }: { points: TensionPoint[]; currentRound: number; onSelectRound: (r: number) => void }) {
  if (points.length === 0) return null
  const maxT = Math.max(1, ...points.map(p => p.avg_tension))
  const maxR = Math.max(...points.map(p => p.round))
  const x = (r: number) => PAD.left + ((r - 1) / Math.max(1, maxR - 1)) * PLOT_W
  const y = (t: number) => PAD.top + PLOT_H - (t / maxT) * PLOT_H
  const line = points.map(p => `${x(p.round).toFixed(1)},${y(p.avg_tension).toFixed(1)}`).join(' ')

  return (
    <svg width="100%" viewBox={`0 0 ${SVG_W} ${SVG_H}`} style={{ display: 'block', cursor: 'pointer' }}
      onClick={e => {
        const rect = (e.currentTarget as SVGSVGElement).getBoundingClientRect()
        const relX = (e.clientX - rect.left) / rect.width * SVG_W - PAD.left
        const r = Math.round(1 + (relX / PLOT_W) * (maxR - 1))
        onSelectRound(Math.max(1, Math.min(maxR, r)))
      }}>
      <polyline points={line} fill="none" stroke="var(--color-accent)" strokeWidth="1.5" opacity="0.6" />
      {points.filter(p => p.fracture_count > 0).map((p, i) => (
        <circle key={i} cx={x(p.round)} cy={y(p.avg_tension)} r="4" fill="var(--color-danger)" opacity="0.8" />
      ))}
      {/* Current round indicator */}
      <line x1={x(currentRound)} y1={PAD.top} x2={x(currentRound)} y2={PAD.top + PLOT_H}
        stroke="var(--color-accent)" strokeWidth="1.5" />
      <circle cx={x(currentRound)} cy={y(points.find(p => p.round === currentRound)?.avg_tension ?? 0)} r="5"
        fill="var(--color-accent)" stroke="var(--color-background)" strokeWidth="2" />
      {/* X axis labels */}
      {[1, Math.round(maxR / 2), maxR].map(r => (
        <text key={r} x={x(r)} y={SVG_H - 4} textAnchor="middle" fontSize="9" fill="var(--color-text-muted)">{r}</text>
      ))}
    </svg>
  )
}

function AgentList({ title, actions, accentColor }: { title: string; actions: RoundRow[]; accentColor: string }) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const toggle = (id: string) => setExpanded(prev => {
    const next = new Set(prev)
    if (next.has(id)) next.delete(id); else next.add(id)
    return next
  })

  return (
    <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', overflow: 'hidden' }}>
      <div style={{ padding: '10px 16px', borderBottom: '1px solid var(--color-border)', fontWeight: '600', fontSize: '13px', color: accentColor }}>
        {title}
      </div>
      <div style={{ maxHeight: '420px', overflowY: 'auto' }}>
        {actions.length === 0 ? (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--color-text-muted)', fontSize: '12px' }}>
            Nenhuma ação registrada
          </div>
        ) : actions.map(a => (
          <div key={a.id} style={{ borderBottom: '1px solid var(--color-border)', cursor: 'pointer' }}
            onClick={() => toggle(a.id)}>
            <div style={{ padding: '10px 14px', display: 'flex', alignItems: 'flex-start', gap: '10px' }}>
              {a.fracture_proposed && (
                <span style={{ fontSize: '11px', padding: '2px 6px', borderRadius: '4px', background: 'var(--color-danger)', color: '#fff', fontWeight: '700', flexShrink: 0, marginTop: '1px' }}>
                  {a.fracture_accepted ? '⚡ Aceito' : '⚡ Rejeitado'}
                </span>
              )}
              <div style={{ flex: 1, fontSize: '12px', color: 'var(--color-text)', lineHeight: '1.5' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontWeight: '500' }}>{a.agent_type === 'disruptor' ? '◈' : '◎'}</span>
                  <span style={{ fontSize: '10px', color: 'var(--color-text-muted)' }}>
                    tensão {Math.round(a.tension_level * 100)}%
                  </span>
                </div>
                <p style={{ margin: '4px 0 0', color: expanded.has(a.id) ? 'var(--color-text)' : 'var(--color-text-muted)' }}>
                  {expanded.has(a.id) ? a.action_text : truncate(a.action_text, 90)}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function Pill({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
      <span>{label}:</span>
      <span style={{ fontWeight: '700', color }}>{value}</span>
    </div>
  )
}

// ─── helpers ─────────────────────────────────────────────────────────────────

function groupByRound(rows: RoundRow[]): RoundGroup[] {
  const map = new Map<number, RoundRow[]>()
  for (const r of rows) {
    if (!map.has(r.round_number)) map.set(r.round_number, [])
    map.get(r.round_number)!.push(r)
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a - b)
    .map(([round, actions]) => ({
      round,
      actions,
      avgTension: actions.reduce((s, a) => s + a.tension_level, 0) / actions.length,
      fractureCount: actions.filter(a => a.fracture_proposed).length,
    }))
}

function truncate(s: string, n: number) {
  return s && s.length > n ? s.slice(0, n) + '…' : s
}

function tensionColor(t: number) {
  if (t >= 0.7) return 'var(--color-danger)'
  if (t >= 0.5) return 'var(--color-warning)'
  if (t >= 0.3) return '#f97316'
  return 'var(--color-success)'
}

const backBtn: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-text-muted)',
  fontSize: '13px', cursor: 'pointer', padding: '0',
}

const ctrlBtn: React.CSSProperties = {
  padding: '6px 14px', borderRadius: '7px', border: '1px solid var(--color-border)',
  background: 'var(--color-surface)', color: 'var(--color-text)',
  fontSize: '13px', cursor: 'pointer', fontWeight: '600',
}
