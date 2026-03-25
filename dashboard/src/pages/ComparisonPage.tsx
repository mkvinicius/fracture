import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface TensionDelta {
  rule_id: string
  description: string
  tensions: number[]
  delta: number
}

interface ComparisonReport {
  simulation_ids: string[]
  questions: string[]
  common_fractures: string[]
  divergent_fractures: Record<string, string[]>
  tension_delta: TensionDelta[]
  confidence_delta: number
  summary: string
}

const SIM_COLORS = [
  'var(--color-accent)',
  'var(--color-success)',
  '#f97316',
  'oklch(0.6 0.15 280)',
  'var(--color-warning)',
]

export default function ComparisonPage({ simIds, onNavigate }: { simIds: string[]; onNavigate: (p: Page, simId?: string) => void }) {
  const [report, setReport] = useState<ComparisonReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (simIds.length < 2) {
      setError('At least 2 simulations required for comparison')
      setLoading(false)
      return
    }
    fetch(`/api/simulations/compare?ids=${simIds.join(',')}`)
      .then(r => { if (!r.ok) throw new Error(`HTTP ${r.status}`); return r.json() })
      .then(d => { setReport(d); setLoading(false) })
      .catch(e => { setError(e.message); setLoading(false) })
  }, [simIds.join(',')])

  if (loading) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--color-text-muted)' }}>
      Comparing simulations...
    </div>
  )

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <button onClick={() => onNavigate('simulations')} style={backBtnStyle}>← Back</button>

      <div style={{ marginTop: '16px', marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Comparison Report</h1>
        <p style={{ margin: '6px 0 0', fontSize: '13px', color: 'var(--color-text-muted)' }}>
          {simIds.length} simulations compared
        </p>
      </div>

      {error && (
        <div style={{ padding: '16px', borderRadius: '8px', background: 'var(--color-danger)15', border: '1px solid var(--color-danger)', color: 'var(--color-danger)', marginBottom: '24px' }}>
          {error}
        </div>
      )}

      {report && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '28px' }}>
          {/* Summary */}
          <div style={{ padding: '20px', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)' }}>
            <p style={{ margin: 0, fontSize: '14px', lineHeight: '1.7', color: 'var(--color-text)' }}>{report.summary}</p>
          </div>

          {/* Simulation legend */}
          <div>
            <SectionTitle>Simulations</SectionTitle>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {report.simulation_ids.map((id, i) => (
                <div key={id} style={{ display: 'flex', gap: '10px', alignItems: 'center', padding: '10px 14px', background: 'var(--color-surface)', borderRadius: '8px', border: `1px solid ${SIM_COLORS[i % SIM_COLORS.length]}44` }}>
                  <div style={{ width: '10px', height: '10px', borderRadius: '50%', background: SIM_COLORS[i % SIM_COLORS.length], flexShrink: 0 }} />
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '2px' }}>{id.slice(0, 16)}…</div>
                    <div style={{ fontSize: '13px', color: 'var(--color-text)' }}>{report.questions[i] || '—'}</div>
                  </div>
                  <button onClick={() => onNavigate('result', id)} style={linkBtnStyle}>View</button>
                </div>
              ))}
            </div>
          </div>

          {/* Common fractures */}
          <div>
            <SectionTitle>Common Rupture Patterns</SectionTitle>
            {report.common_fractures?.length > 0 ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {report.common_fractures.map((f, i) => (
                  <div key={i} style={{ display: 'flex', gap: '10px', alignItems: 'center', padding: '10px 14px', background: 'var(--color-success)10', borderRadius: '8px', border: '1px solid var(--color-success)44' }}>
                    <span style={{ color: 'var(--color-success)', fontSize: '14px', fontWeight: '700' }}>✓</span>
                    <span style={{ fontSize: '13px', color: 'var(--color-text)' }}>{f}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div style={{ padding: '20px', textAlign: 'center', color: 'var(--color-text-muted)', fontSize: '13px', background: 'var(--color-surface)', borderRadius: '8px', border: '1px solid var(--color-border)' }}>
                No common rupture patterns — high divergence between runs
              </div>
            )}
          </div>

          {/* Divergent fractures */}
          {Object.keys(report.divergent_fractures || {}).length > 0 && (
            <div>
              <SectionTitle>Divergent Patterns (unique per simulation)</SectionTitle>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {Object.entries(report.divergent_fractures).map(([simId, fractures], idx) => (
                  <div key={simId} style={{ padding: '14px 16px', background: 'var(--color-surface)', borderRadius: '8px', border: `1px solid ${SIM_COLORS[idx % SIM_COLORS.length]}44` }}>
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center', marginBottom: '10px' }}>
                      <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: SIM_COLORS[idx % SIM_COLORS.length] }} />
                      <span style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>{simId.slice(0, 20)}</span>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                      {fractures.map((f, i) => (
                        <div key={i} style={{ fontSize: '13px', color: 'var(--color-text)', paddingLeft: '16px' }}>• {f}</div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Tension delta bars */}
          {report.tension_delta?.length > 0 && (
            <div>
              <SectionTitle>Tension Variance (top rules)</SectionTitle>
              <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '20px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
                {report.tension_delta.map((td, i) => (
                  <div key={i}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                      <span style={{ fontSize: '13px', color: 'var(--color-text)', fontWeight: '500' }}>{td.description}</span>
                      <span style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>Δ={td.delta.toFixed(2)}</span>
                    </div>
                    <div style={{ display: 'flex', gap: '4px' }}>
                      {td.tensions.map((t, j) => (
                        <div key={j} style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '2px', alignItems: 'center' }}>
                          <div style={{ width: '100%', height: '32px', background: 'var(--color-background)', borderRadius: '4px', overflow: 'hidden', display: 'flex', alignItems: 'flex-end' }}>
                            <div style={{ width: '100%', height: `${t * 100}%`, background: SIM_COLORS[j % SIM_COLORS.length], borderRadius: '2px', transition: 'height 0.3s' }} />
                          </div>
                          <span style={{ fontSize: '10px', color: 'var(--color-text-muted)' }}>{Math.round(t * 100)}%</span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Confidence delta */}
          <div style={{ padding: '20px', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', display: 'flex', gap: '24px', alignItems: 'center' }}>
            <div>
              <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '4px' }}>Confidence Spread</div>
              <div style={{ fontSize: '28px', fontWeight: '700', color: report.confidence_delta > 0.15 ? 'var(--color-warning)' : 'var(--color-success)' }}>
                {report.confidence_delta > 0 ? '↑' : '↓'}{Math.round(report.confidence_delta * 100)}%
              </div>
            </div>
            <div style={{ fontSize: '13px', color: 'var(--color-text-muted)', lineHeight: '1.5' }}>
              {report.confidence_delta > 0.15
                ? 'High variance — results are sensitive to initial conditions'
                : 'Low variance — results are stable across runs'}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 style={{ fontSize: '15px', fontWeight: '700', color: 'var(--color-text)', margin: '0 0 14px', paddingBottom: '8px', borderBottom: '1px solid var(--color-border)' }}>
      {children}
    </h2>
  )
}

const backBtnStyle: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-text-muted)', fontSize: '13px',
  cursor: 'pointer', padding: '0',
}

const linkBtnStyle: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-accent)', fontSize: '12px',
  cursor: 'pointer', padding: '0', fontWeight: '600',
}
