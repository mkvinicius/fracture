import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface TensionPoint {
  round: number
  avg_tension: number
  fracture_count: number
}

const THRESHOLD = 0.7   // default fracture threshold
const SVG_W = 720
const SVG_H = 260
const PAD = { top: 20, right: 20, bottom: 40, left: 52 }
const PLOT_W = SVG_W - PAD.left - PAD.right
const PLOT_H = SVG_H - PAD.top - PAD.bottom

export default function ConvergencePage({ simId, onNavigate }: { simId: string; onNavigate: (p: Page, simId?: string) => void }) {
  const [points, setPoints] = useState<TensionPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch(`/api/v1/simulations/${simId}/events`)
      .then(r => { if (!r.ok) throw new Error(`HTTP ${r.status}`); return r.json() })
      .then((d: TensionPoint[]) => { setPoints(d ?? []); setLoading(false) })
      .catch(e => { setError(e.message); setLoading(false) })
  }, [simId])

  if (loading) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--color-text-muted)' }}>
      Carregando dados de convergência...
    </div>
  )

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <button onClick={() => onNavigate('result', simId)} style={backBtnStyle}>← Voltar ao relatório</button>

      <div style={{ marginTop: '16px', marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Gráfico de Convergência</h1>
        <p style={{ margin: '6px 0 0', fontSize: '13px', color: 'var(--color-text-muted)' }}>
          Tensão média de mercado por rodada — velocidade de convergência da simulação
        </p>
      </div>

      {error && (
        <div style={{ padding: '16px', borderRadius: '8px', background: 'var(--color-danger)15', border: '1px solid var(--color-danger)', color: 'var(--color-danger)', marginBottom: '24px' }}>
          {error}
        </div>
      )}

      {points.length === 0 && !error ? (
        <div style={{ padding: '48px', textAlign: 'center', color: 'var(--color-text-muted)', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)' }}>
          Nenhum dado de rodada disponível para esta simulação
        </div>
      ) : (
        <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '24px', overflowX: 'auto' }}>
          <TensionChart points={points} />
          <Legend />
        </div>
      )}

      {points.length > 0 && (
        <div style={{ marginTop: '20px', display: 'flex', gap: '20px', flexWrap: 'wrap' }}>
          <StatCard label="Rodadas" value={String(points[points.length - 1]?.round ?? 0)} />
          <StatCard label="Tensão Máxima" value={`${Math.round(Math.max(...points.map(p => p.avg_tension)) * 100)}%`} />
          <StatCard label="Tensão Final" value={`${Math.round((points[points.length - 1]?.avg_tension ?? 0) * 100)}%`} />
          <StatCard label="Pontos de Ruptura" value={String(points.reduce((s, p) => s + p.fracture_count, 0))} />
        </div>
      )}
    </div>
  )
}

function TensionChart({ points }: { points: TensionPoint[] }) {
  if (points.length === 0) return null

  const maxRound = Math.max(...points.map(p => p.round))
  const maxTension = Math.max(1, Math.max(...points.map(p => p.avg_tension)))

  const xScale = (round: number) => PAD.left + ((round - 1) / Math.max(1, maxRound - 1)) * PLOT_W
  const yScale = (t: number) => PAD.top + PLOT_H - (t / maxTension) * PLOT_H
  const thresholdY = yScale(THRESHOLD / maxTension * maxTension)

  // Build polyline points string
  const linePoints = points.map(p => `${xScale(p.round).toFixed(1)},${yScale(p.avg_tension).toFixed(1)}`).join(' ')

  // Area fill path
  const areaPath = [
    `M ${xScale(points[0].round).toFixed(1)} ${(PAD.top + PLOT_H).toFixed(1)}`,
    ...points.map(p => `L ${xScale(p.round).toFixed(1)} ${yScale(p.avg_tension).toFixed(1)}`),
    `L ${xScale(points[points.length - 1].round).toFixed(1)} ${(PAD.top + PLOT_H).toFixed(1)}`,
    'Z',
  ].join(' ')

  // Y-axis labels
  const yLabels = [0, 0.25, 0.5, 0.75, 1.0].map(v => ({
    y: yScale(v * maxTension),
    label: `${Math.round(v * 100)}%`,
  }))

  // X-axis labels (up to 8)
  const step = Math.max(1, Math.ceil(maxRound / 8))
  const xLabels = points.filter(p => p.round === 1 || p.round % step === 0 || p.round === maxRound)

  // Fracture point markers
  const fracturePoints = points.filter(p => p.fracture_count > 0)

  return (
    <svg width="100%" viewBox={`0 0 ${SVG_W} ${SVG_H}`} style={{ fontFamily: 'inherit', display: 'block' }}>
      <defs>
        <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="var(--color-accent)" stopOpacity="0.25" />
          <stop offset="100%" stopColor="var(--color-accent)" stopOpacity="0.03" />
        </linearGradient>
      </defs>

      {/* Grid lines */}
      {yLabels.map((yl, i) => (
        <line key={i} x1={PAD.left} y1={yl.y} x2={PAD.left + PLOT_W} y2={yl.y}
          stroke="var(--color-border)" strokeWidth="1" strokeDasharray={i === 0 ? 'none' : '3,3'} />
      ))}

      {/* Threshold line */}
      <line
        x1={PAD.left} y1={thresholdY} x2={PAD.left + PLOT_W} y2={thresholdY}
        stroke="var(--color-danger)" strokeWidth="1.5" strokeDasharray="6,4"
        opacity="0.7"
      />
      <text x={PAD.left + PLOT_W - 4} y={thresholdY - 5} textAnchor="end"
        fill="var(--color-danger)" fontSize="10" opacity="0.8">limiar</text>

      {/* Area fill */}
      <path d={areaPath} fill="url(#areaGrad)" />

      {/* Main line */}
      <polyline
        points={linePoints}
        fill="none"
        stroke="var(--color-accent)"
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
      />

      {/* Fracture point markers */}
      {fracturePoints.map((p, i) => (
        <g key={i}>
          <circle
            cx={xScale(p.round)} cy={yScale(p.avg_tension)}
            r="5" fill="var(--color-danger)" stroke="var(--color-background)" strokeWidth="2"
          />
        </g>
      ))}

      {/* Y-axis labels */}
      {yLabels.map((yl, i) => (
        <text key={i} x={PAD.left - 6} y={yl.y + 4} textAnchor="end"
          fill="var(--color-text-muted)" fontSize="11">{yl.label}</text>
      ))}

      {/* X-axis labels */}
      {xLabels.map(p => (
        <text key={p.round} x={xScale(p.round)} y={PAD.top + PLOT_H + 16}
          textAnchor="middle" fill="var(--color-text-muted)" fontSize="11">
          {p.round}
        </text>
      ))}

      {/* Axis labels */}
      <text x={PAD.left + PLOT_W / 2} y={SVG_H - 2} textAnchor="middle"
        fill="var(--color-text-muted)" fontSize="11">Rodada</text>
      <text x={14} y={PAD.top + PLOT_H / 2} textAnchor="middle"
        fill="var(--color-text-muted)" fontSize="11"
        transform={`rotate(-90, 14, ${PAD.top + PLOT_H / 2})`}>Tensão</text>
    </svg>
  )
}

function Legend() {
  return (
    <div style={{ display: 'flex', gap: '20px', marginTop: '12px', flexWrap: 'wrap' }}>
      <LegendItem color="var(--color-accent)" label="Tensão média" />
      <LegendItem color="var(--color-danger)" label="Ponto de ruptura" dot />
      <LegendItem color="var(--color-danger)" dashed label="Limiar (0.7)" />
    </div>
  )
}

function LegendItem({ color, label, dot, dashed }: { color: string; label: string; dot?: boolean; dashed?: boolean }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
      {dot ? (
        <div style={{ width: '10px', height: '10px', borderRadius: '50%', background: color, border: '2px solid var(--color-background)' }} />
      ) : (
        <div style={{ width: '24px', height: '2px', background: color, borderRadius: '1px', borderTop: dashed ? `2px dashed ${color}` : undefined }} />
      )}
      {label}
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ padding: '14px 20px', background: 'var(--color-surface)', borderRadius: '8px', border: '1px solid var(--color-border)', minWidth: '100px' }}>
      <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '4px' }}>{label}</div>
      <div style={{ fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>{value}</div>
    </div>
  )
}

const backBtnStyle: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-text-muted)', fontSize: '13px',
  cursor: 'pointer', padding: '0',
}
