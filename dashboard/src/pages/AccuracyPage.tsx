import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface CalibrationRow {
  archetype_id: string
  domain: string
  accuracy_weight: number
  sample_count: number
}

interface AccuracyReport {
  feedback_count: number
  average_delta: number
  accurate_count: number
  partial_count: number
  inaccurate_count: number
  calibrations: CalibrationRow[] | null
}

interface ConfirmedRupture {
  id: string
  simulation_id: string
  rule_id: string
  rule_description: string
  notes: string
  confirmed_at: number
}

function Card({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '20px', ...style }}>
      {children}
    </div>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 style={{ margin: '0 0 16px', fontSize: '15px', fontWeight: '700', color: 'var(--color-text)', paddingBottom: '8px', borderBottom: '1px solid var(--color-border)' }}>
      {children}
    </h2>
  )
}

function DeltaBadge({ delta }: { delta: number }) {
  const color = delta >= 0.5 ? 'var(--color-success)' : delta >= 0 ? 'var(--color-warning)' : 'var(--color-danger)'
  return (
    <span style={{ padding: '2px 10px', borderRadius: '20px', fontSize: '12px', fontWeight: '700', background: `${color}22`, color }}>
      {delta >= 0 ? '+' : ''}{delta.toFixed(2)}
    </span>
  )
}

function WeightBar({ weight }: { weight: number }) {
  // weight range 0.3–2.0; neutral = 1.0
  const pct = ((weight - 0.3) / 1.7) * 100
  const color = weight >= 1.3 ? 'var(--color-success)' : weight >= 0.9 ? 'var(--color-accent)' : 'var(--color-danger)'
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
      <div style={{ flex: 1, height: '6px', borderRadius: '3px', background: 'var(--color-background)', overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${pct}%`, background: color, borderRadius: '3px', transition: 'width 0.3s' }} />
      </div>
      <span style={{ fontSize: '11px', fontWeight: '700', color, minWidth: '36px' }}>{weight.toFixed(2)}×</span>
    </div>
  )
}

export default function AccuracyPage({ onNavigate }: { onNavigate: (p: Page, simId?: string) => void }) {
  const [report, setReport] = useState<AccuracyReport | null>(null)
  const [ruptures, setRuptures] = useState<ConfirmedRupture[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([
      fetch('/api/company/accuracy').then(r => { if (!r.ok) throw new Error(`HTTP ${r.status}`); return r.json() }),
      fetch('/api/company/confirmations').then(r => r.ok ? r.json() : []).catch(() => []),
    ]).then(([rep, confs]) => {
      setReport(rep)
      setRuptures(confs || [])
      setLoading(false)
    }).catch(e => { setError(e.message); setLoading(false) })
  }, [])

  if (loading) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--color-text-muted)' }}>
      Loading accuracy data...
    </div>
  )

  if (error || !report) return (
    <div style={{ padding: '32px', color: 'var(--color-danger)' }}>
      {error || 'Failed to load accuracy report'}
    </div>
  )

  const total = report.feedback_count || 1
  const calRows = report.calibrations || []

  return (
    <div style={{ padding: '32px', maxWidth: '960px', display: 'flex', flexDirection: 'column', gap: '32px' }}>
      {/* Header */}
      <div>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Accuracy & Calibration</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>
          Real-world feedback loop — como as previsões do FRACTURE se confirmaram
        </p>
      </div>

      {/* Overview stats */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))', gap: '12px' }}>
        <Card>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '6px', fontWeight: '600' }}>FEEDBACKS</div>
          <div style={{ fontSize: '28px', fontWeight: '700', color: 'var(--color-text)' }}>{report.feedback_count}</div>
        </Card>
        <Card>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '6px', fontWeight: '600' }}>DELTA MÉDIO</div>
          <div style={{ fontSize: '28px', fontWeight: '700' }}>
            <DeltaBadge delta={report.average_delta} />
          </div>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginTop: '4px' }}>−1 (errado) a +1 (perfeito)</div>
        </Card>
        <Card>
          <div style={{ fontSize: '11px', color: 'var(--color-success)', marginBottom: '6px', fontWeight: '600' }}>PRECISOS</div>
          <div style={{ fontSize: '28px', fontWeight: '700', color: 'var(--color-success)' }}>{report.accurate_count}</div>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{Math.round((report.accurate_count / total) * 100)}%</div>
        </Card>
        <Card>
          <div style={{ fontSize: '11px', color: 'var(--color-warning)', marginBottom: '6px', fontWeight: '600' }}>PARCIAIS</div>
          <div style={{ fontSize: '28px', fontWeight: '700', color: 'var(--color-warning)' }}>{report.partial_count}</div>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{Math.round((report.partial_count / total) * 100)}%</div>
        </Card>
        <Card>
          <div style={{ fontSize: '11px', color: 'var(--color-danger)', marginBottom: '6px', fontWeight: '600' }}>IMPRECISOS</div>
          <div style={{ fontSize: '28px', fontWeight: '700', color: 'var(--color-danger)' }}>{report.inaccurate_count}</div>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{Math.round((report.inaccurate_count / total) * 100)}%</div>
        </Card>
        <Card>
          <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '6px', fontWeight: '600' }}>RUPTURAS CONFIRMADAS</div>
          <div style={{ fontSize: '28px', fontWeight: '700', color: 'var(--color-text)' }}>{ruptures.length}</div>
        </Card>
      </div>

      {/* Accuracy distribution bar */}
      {report.feedback_count > 0 && (
        <Card>
          <SectionTitle>Distribuição de Precisão</SectionTitle>
          <div style={{ display: 'flex', gap: '2px', height: '24px', borderRadius: '6px', overflow: 'hidden' }}>
            {report.accurate_count > 0 && (
              <div style={{ flex: report.accurate_count, background: 'var(--color-success)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <span style={{ fontSize: '11px', fontWeight: '700', color: '#fff' }}>{Math.round((report.accurate_count / total) * 100)}%</span>
              </div>
            )}
            {report.partial_count > 0 && (
              <div style={{ flex: report.partial_count, background: 'var(--color-warning)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <span style={{ fontSize: '11px', fontWeight: '700', color: '#fff' }}>{Math.round((report.partial_count / total) * 100)}%</span>
              </div>
            )}
            {report.inaccurate_count > 0 && (
              <div style={{ flex: report.inaccurate_count, background: 'var(--color-danger)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <span style={{ fontSize: '11px', fontWeight: '700', color: '#fff' }}>{Math.round((report.inaccurate_count / total) * 100)}%</span>
              </div>
            )}
          </div>
          <div style={{ display: 'flex', gap: '16px', marginTop: '10px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
            <span><span style={{ color: 'var(--color-success)' }}>■</span> Preciso</span>
            <span><span style={{ color: 'var(--color-warning)' }}>■</span> Parcial</span>
            <span><span style={{ color: 'var(--color-danger)' }}>■</span> Impreciso</span>
          </div>
        </Card>
      )}

      {/* Archetype calibration */}
      {calRows.length > 0 && (
        <div>
          <SectionTitle>Calibração de Arquétipos</SectionTitle>
          <Card style={{ padding: 0, overflow: 'hidden' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--color-border)', background: 'var(--color-background)' }}>
                  <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: '600', color: 'var(--color-text-muted)', fontSize: '11px' }}>ARQUÉTIPO</th>
                  <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: '600', color: 'var(--color-text-muted)', fontSize: '11px' }}>DOMÍNIO</th>
                  <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: '600', color: 'var(--color-text-muted)', fontSize: '11px', minWidth: '160px' }}>PESO DE PRECISÃO</th>
                  <th style={{ padding: '10px 16px', textAlign: 'right', fontWeight: '600', color: 'var(--color-text-muted)', fontSize: '11px' }}>AMOSTRAS</th>
                </tr>
              </thead>
              <tbody>
                {calRows.map((c, i) => (
                  <tr key={i} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td style={{ padding: '10px 16px', color: 'var(--color-text)', fontWeight: '500' }}>{c.archetype_id}</td>
                    <td style={{ padding: '10px 16px', color: 'var(--color-text-muted)' }}>{c.domain}</td>
                    <td style={{ padding: '10px 16px', minWidth: '160px' }}><WeightBar weight={c.accuracy_weight} /></td>
                    <td style={{ padding: '10px 16px', textAlign: 'right', color: 'var(--color-text-muted)' }}>{c.sample_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          <p style={{ marginTop: '8px', fontSize: '11px', color: 'var(--color-text-muted)' }}>
            Peso neutro = 1.0×. Arquétipos com &gt;1.0× recebem mais influência nas próximas simulações.
          </p>
        </div>
      )}

      {/* Confirmed ruptures */}
      {ruptures.length > 0 && (
        <div>
          <SectionTitle>Rupturas Confirmadas</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {ruptures.map((r, i) => (
              <Card key={i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '16px' }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: '14px', fontWeight: '600', color: 'var(--color-text)', marginBottom: '4px' }}>{r.rule_description}</div>
                  {r.notes && <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>{r.notes}</div>}
                </div>
                <div style={{ flexShrink: 0, textAlign: 'right' }}>
                  <button
                    onClick={() => onNavigate('result', r.simulation_id)}
                    style={{ padding: '4px 10px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '11px', cursor: 'pointer' }}
                  >
                    Ver simulação
                  </button>
                  <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginTop: '4px' }}>
                    {new Date(r.confirmed_at * 1000).toLocaleDateString()}
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {report.feedback_count === 0 && ruptures.length === 0 && (
        <Card style={{ textAlign: 'center', padding: '48px' }}>
          <div style={{ fontSize: '32px', marginBottom: '12px' }}>◎</div>
          <div style={{ fontWeight: '600', marginBottom: '6px', color: 'var(--color-text)' }}>Sem dados de precisão ainda</div>
          <div style={{ fontSize: '13px', color: 'var(--color-text-muted)' }}>
            Dê feedback nas simulações concluídas para começar a calibrar os arquétipos.
          </div>
        </Card>
      )}
    </div>
  )
}
