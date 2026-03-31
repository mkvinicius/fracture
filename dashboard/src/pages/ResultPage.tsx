import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface TimelineEntry { horizon: string; description: string; confidence: number }
interface TensionEntry { rule_id: string; description: string; domain: string; tension: number; color: string }
interface RuptureScenario { rule_id: string; rule_description: string; probability: number; who_breaks: string; how_it_happens: string; impact_on_company: string; how_to_be_first: string }
interface Coalition { name: string; agent_names: string[]; shared_goal: string; strength: number; is_disruptive: boolean }
interface RuptureTimelineEvent { horizon: string; rule_id: string; description: string; trigger: string; probability: number }
interface ActionPlaybook { horizon_90_days: string[]; horizon_1_year: string[]; horizon_3_years: string[]; quick_wins: string[]; critical_risks: string[] }
interface FractureEvent { round: number; proposed_by: string; accepted: boolean; proposal: { original_rule_id: string; new_description: string } }
interface EnsembleResult { runs: number; consensus_probability: number; variance: number; consensus_scenarios: string[] }

interface FullReport {
  simulation_id: string
  question: string
  probable_future: { narrative: string; timeline: TimelineEntry[]; confidence: number; key_assumptions: string[] }
  tension_map: TensionEntry[]
  rupture_scenarios: RuptureScenario[]
  coalitions?: Coalition[]
  rupture_timeline?: RuptureTimelineEvent[]
  action_playbook?: ActionPlaybook
  fracture_events: FractureEvent[]
  ensemble_result?: EnsembleResult
  total_tokens: number
  duration_ms: number
  watermark: { tool: string; version: string; generated_at: string; notice: string }
}

const tensionColorMap: Record<string, string> = {
  green: 'var(--color-success)',
  yellow: 'var(--color-warning)',
  orange: '#f97316',
  red: 'var(--color-danger)',
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 style={{ fontSize: '15px', fontWeight: '700', color: 'var(--color-text)', margin: '0 0 16px', paddingBottom: '8px', borderBottom: '1px solid var(--color-border)' }}>
      {children}
    </h2>
  )
}

function Card({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '20px', ...style }}>
      {children}
    </div>
  )
}

export default function ResultPage({ simId, onNavigate }: { simId: string; onNavigate: (p: Page, simId?: string) => void }) {
  const [report, setReport] = useState<FullReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [shareUrl, setShareUrl] = useState<string | null>(null)
  const [sharing, setSharing] = useState(false)
  const [injectText, setInjectText] = useState('')
  const [injecting, setInjecting] = useState(false)
  const [injectStatus, setInjectStatus] = useState<string | null>(null)
  const [simStatus, setSimStatus] = useState<string>('done')

  useEffect(() => {
    fetch(`/api/v1/simulations/${simId}/report`)
      .then(r => { if (!r.ok) throw new Error(`HTTP ${r.status}`); return r.json() })
      .then(d => { setReport(d); setLoading(false) })
      .catch(e => { setError(e.message); setLoading(false) })
    // Check sim status for God View
    fetch(`/api/v1/simulations/${simId}`)
      .then(r => r.json()).then(d => setSimStatus(d.status || 'done')).catch(() => {})
  }, [simId])

  async function handleShare() {
    setSharing(true)
    try {
      const res = await fetch(`/api/v1/simulations/${simId}/share`, { method: 'POST' })
      const data = await res.json()
      if (data.token) {
        const url = `${window.location.origin}/share/${data.token}`
        setShareUrl(url)
        navigator.clipboard?.writeText(url)
      }
    } catch { /* ignore */ } finally { setSharing(false) }
  }

  async function handleInject() {
    if (!injectText.trim()) return
    setInjecting(true)
    setInjectStatus(null)
    try {
      const res = await fetch(`/api/v1/simulations/${simId}/inject`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ event: injectText })
      })
      if (res.ok) { setInjectStatus('Evento injetado — agentes reagirão na próxima rodada'); setInjectText('') }
      else { const d = await res.json(); setInjectStatus(`Erro: ${d.error}`) }
    } catch { setInjectStatus('Erro ao injetar evento') } finally { setInjecting(false) }
  }

  if (loading) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--color-text-muted)' }}>
      Carregando relatório...
    </div>
  )

  if (error || !report) return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <button onClick={() => onNavigate('simulations')} style={backBtnStyle}>← Voltar</button>
      <div style={{ marginTop: '24px', padding: '32px', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', color: 'var(--color-danger)', textAlign: 'center' }}>
        {error || 'Relatório não disponível'}
      </div>
    </div>
  )

  const conf = (v: number) => `${Math.round(v * 100)}%`

  return (
    <div style={{ padding: '32px', maxWidth: '960px', display: 'flex', flexDirection: 'column', gap: '32px' }}>
      {/* Header */}
      <div>
        <button onClick={() => onNavigate('simulations')} style={backBtnStyle}>← Voltar</button>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginTop: '16px', gap: '16px' }}>
          <div>
            <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Relatório da Simulação</h1>
            <p style={{ margin: '6px 0 0', fontSize: '14px', color: 'var(--color-text-muted)' }}>{report.question}</p>
            <div style={{ marginTop: '8px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
              {report.total_tokens.toLocaleString()} tokens · {(report.duration_ms / 1000).toFixed(1)}s · {report.watermark.version}
            </div>
          </div>
          <div style={{ display: 'flex', gap: '8px', flexShrink: 0, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
            <button onClick={() => downloadFile(`/api/v1/simulations/${simId}/export/markdown`, `fracture-${simId}.md`)} style={exportBtnStyle}>⬇ Markdown</button>
            <button onClick={() => downloadFile(`/api/v1/simulations/${simId}/export/json`, `fracture-${simId}.json`)} style={exportBtnStyle}>⬇ JSON</button>
            <button onClick={handleShare} disabled={sharing}
              style={{ padding: '9px 18px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '13px', fontWeight: '600', cursor: 'pointer', whiteSpace: 'nowrap' }}>
              {sharing ? '...' : '🔗 Compartilhar'}
            </button>
            <button
              onClick={() => onNavigate('feedback', simId)}
              style={{ padding: '9px 18px', borderRadius: '8px', border: '1px solid var(--color-accent)', background: 'transparent', color: 'var(--color-accent)', fontSize: '13px', fontWeight: '600', cursor: 'pointer', whiteSpace: 'nowrap' }}
            >
              Dar Feedback
            </button>
          </div>
        </div>
        {/* Share URL toast */}
        {shareUrl && (
          <div style={{ marginTop: '12px', padding: '10px 14px', borderRadius: '8px', background: 'oklch(0.65 0.22 30 / 0.1)', border: '1px solid oklch(0.65 0.22 30 / 0.3)', fontSize: '12px', display: 'flex', alignItems: 'center', gap: '10px' }}>
            <span style={{ color: 'var(--color-success)' }}>✓ Link copiado:</span>
            <code style={{ color: 'var(--color-text)', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{shareUrl}</code>
            <button onClick={() => setShareUrl(null)} style={{ background: 'none', border: 'none', color: 'var(--color-text-muted)', cursor: 'pointer' }}>✕</button>
          </div>
        )}
        {/* God View panel — only shown when sim is running */}
        {simStatus === 'running' && (
          <div style={{ marginTop: '16px', padding: '16px', borderRadius: '10px', border: '1px solid oklch(0.75 0.18 55 / 0.5)', background: 'oklch(0.75 0.18 55 / 0.05)' }}>
            <div style={{ fontSize: '13px', fontWeight: '700', color: 'oklch(0.75 0.18 55)', marginBottom: '10px' }}>⚡ God View — Injetar Evento</div>
            <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '10px' }}>Injete um evento externo agora. Os agentes reagirão na próxima rodada.</div>
            <div style={{ display: 'flex', gap: '8px' }}>
              <input
                value={injectText}
                onChange={e => setInjectText(e.target.value)}
                placeholder='ex: "O governo anunciou taxação de 15% em delivery"'
                style={{ flex: 1, padding: '8px 12px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'var(--color-surface-2)', color: 'var(--color-text)', fontSize: '13px', outline: 'none' }}
                onKeyDown={e => e.key === 'Enter' && handleInject()}
              />
              <button onClick={handleInject} disabled={injecting || !injectText.trim()}
                style={{ padding: '8px 16px', borderRadius: '6px', border: 'none', background: 'oklch(0.75 0.18 55)', color: '#000', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>
                {injecting ? '...' : 'Injetar'}
              </button>
            </div>
            {injectStatus && <div style={{ marginTop: '8px', fontSize: '12px', color: injectStatus.startsWith('Erro') ? 'var(--color-danger)' : 'var(--color-success)' }}>{injectStatus}</div>}
          </div>
        )}
      </div>

      {/* Ensemble badge (Premium) */}
      {report.ensemble_result && (
        <Card style={{ background: 'oklch(0.18 0.04 280)', borderColor: 'oklch(0.4 0.15 280)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '12px' }}>
            <span style={{ padding: '3px 10px', borderRadius: '20px', background: 'oklch(0.4 0.15 280)', color: '#fff', fontSize: '11px', fontWeight: '700' }}>PREMIUM · ENSEMBLE</span>
            <span style={{ fontSize: '13px', color: 'var(--color-text-muted)' }}>{report.ensemble_result.runs} execuções independentes</span>
          </div>
          <div style={{ display: 'flex', gap: '24px', flexWrap: 'wrap' }}>
            <div>
              <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '2px' }}>Probabilidade de Consenso</div>
              <div style={{ fontSize: '22px', fontWeight: '700', color: 'var(--color-text)' }}>{conf(report.ensemble_result.consensus_probability)}</div>
            </div>
            <div>
              <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '2px' }}>Variância</div>
              <div style={{ fontSize: '22px', fontWeight: '700', color: 'var(--color-text)' }}>{report.ensemble_result.variance.toFixed(3)}</div>
            </div>
          </div>
          {report.ensemble_result.consensus_scenarios?.length > 0 && (
            <div style={{ marginTop: '12px' }}>
              <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '6px' }}>Cenários de Consenso</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                {report.ensemble_result.consensus_scenarios.map((s, i) => (
                  <div key={i} style={{ fontSize: '13px', color: 'var(--color-text)', padding: '6px 10px', background: 'oklch(0.14 0.03 280)', borderRadius: '6px' }}>{s}</div>
                ))}
              </div>
            </div>
          )}
        </Card>
      )}

      {/* Probable Future */}
      <div>
        <SectionTitle>Futuro Provável · {conf(report.probable_future.confidence)} de confiança</SectionTitle>
        <Card>
          <p style={{ margin: '0 0 20px', fontSize: '14px', lineHeight: '1.7', color: 'var(--color-text)' }}>{report.probable_future.narrative}</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {report.probable_future.timeline?.map((t, i) => (
              <div key={i} style={{ display: 'flex', gap: '12px', alignItems: 'flex-start' }}>
                <div style={{ minWidth: '80px', fontSize: '11px', fontWeight: '700', color: 'var(--color-accent)', paddingTop: '2px' }}>{t.horizon}</div>
                <div style={{ flex: 1, fontSize: '13px', color: 'var(--color-text)' }}>{t.description}</div>
                <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', whiteSpace: 'nowrap' }}>{conf(t.confidence)}</div>
              </div>
            ))}
          </div>
          {report.probable_future.key_assumptions?.length > 0 && (
            <div style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid var(--color-border)' }}>
              <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '8px', fontWeight: '600' }}>Premissas-chave</div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px' }}>
                {report.probable_future.key_assumptions.map((a, i) => (
                  <span key={i} style={{ fontSize: '12px', padding: '4px 10px', borderRadius: '20px', background: 'var(--color-background)', border: '1px solid var(--color-border)', color: 'var(--color-text)' }}>{a}</span>
                ))}
              </div>
            </div>
          )}
        </Card>
      </div>

      {/* Tension Map */}
      {report.tension_map?.length > 0 && (
        <div>
          <SectionTitle>Mapa de Tensão</SectionTitle>
          <Card>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {report.tension_map.map((t, i) => (
                <div key={i}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                    <div style={{ fontSize: '13px', color: 'var(--color-text)', fontWeight: '500' }}>{t.description}</div>
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                      <span style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{t.domain}</span>
                      <span style={{ fontSize: '11px', fontWeight: '700', color: tensionColorMap[t.color] || 'var(--color-text)' }}>{Math.round(t.tension * 100)}%</span>
                    </div>
                  </div>
                  <div style={{ height: '6px', borderRadius: '3px', background: 'var(--color-background)', overflow: 'hidden' }}>
                    <div style={{ height: '100%', width: `${t.tension * 100}%`, background: tensionColorMap[t.color] || 'var(--color-accent)', borderRadius: '3px', transition: 'width 0.3s' }} />
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {/* Rupture Scenarios */}
      {report.rupture_scenarios?.length > 0 && (
        <div>
          <SectionTitle>Cenários de Ruptura</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {report.rupture_scenarios.map((s, i) => (
              <Card key={i}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '12px' }}>
                  <div style={{ fontSize: '14px', fontWeight: '600', color: 'var(--color-text)', flex: 1 }}>{s.rule_description}</div>
                  <span style={{ marginLeft: '12px', padding: '3px 10px', borderRadius: '20px', background: `${probColor(s.probability)}22`, color: probColor(s.probability), fontSize: '12px', fontWeight: '700', whiteSpace: 'nowrap' }}>
                    {conf(s.probability)}
                  </span>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px', color: 'var(--color-text-muted)' }}>
                  <div><strong style={{ color: 'var(--color-text)' }}>Quem rompe:</strong> {s.who_breaks}</div>
                  <div><strong style={{ color: 'var(--color-text)' }}>Como:</strong> {s.how_it_happens}</div>
                  <div><strong style={{ color: 'var(--color-text)' }}>Impacto:</strong> {s.impact_on_company}</div>
                </div>
                {s.how_to_be_first && (
                  <div style={{ marginTop: '12px', padding: '12px', borderRadius: '8px', background: 'oklch(0.18 0.04 145)', border: '1px solid oklch(0.35 0.1 145)', fontSize: '13px' }}>
                    <strong style={{ color: 'oklch(0.7 0.15 145)', display: 'block', marginBottom: '4px' }}>Como ser o primeiro:</strong>
                    <span style={{ color: 'var(--color-text)' }}>{s.how_to_be_first}</span>
                  </div>
                )}
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Coalitions */}
      {report.coalitions && report.coalitions.length > 0 && (
        <div>
          <SectionTitle>Coalizões</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '12px' }}>
            {report.coalitions.map((c, i) => (
              <Card key={i}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                  <div style={{ fontSize: '14px', fontWeight: '600', color: 'var(--color-text)' }}>{c.name}</div>
                  {c.is_disruptive && <span style={{ padding: '2px 8px', borderRadius: '20px', background: 'var(--color-danger)22', color: 'var(--color-danger)', fontSize: '11px', fontWeight: '700' }}>DISRUPTIVA</span>}
                </div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '8px' }}>{c.shared_goal}</div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px', marginBottom: '8px' }}>
                  {c.agent_names.map((n, j) => (
                    <span key={j} style={{ fontSize: '11px', padding: '2px 8px', borderRadius: '20px', background: 'var(--color-background)', border: '1px solid var(--color-border)', color: 'var(--color-text)' }}>{n}</span>
                  ))}
                </div>
                <div style={{ height: '4px', borderRadius: '2px', background: 'var(--color-background)' }}>
                  <div style={{ height: '100%', width: `${c.strength * 100}%`, background: 'var(--color-accent)', borderRadius: '2px' }} />
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Rupture Timeline */}
      {report.rupture_timeline && report.rupture_timeline.length > 0 && (
        <div>
          <SectionTitle>Linha do Tempo de Ruptura</SectionTitle>
          <Card>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              {report.rupture_timeline.map((e, i) => (
                <div key={i} style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                  <div style={{ minWidth: '90px', textAlign: 'right' }}>
                    <div style={{ fontSize: '12px', fontWeight: '700', color: 'var(--color-accent)' }}>{e.horizon}</div>
                    <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{conf(e.probability)}</div>
                  </div>
                  <div style={{ width: '1px', background: 'var(--color-border)', alignSelf: 'stretch', position: 'relative' }}>
                    <div style={{ position: 'absolute', top: '4px', left: '-4px', width: '8px', height: '8px', borderRadius: '50%', background: 'var(--color-accent)', border: '2px solid var(--color-background)' }} />
                  </div>
                  <div style={{ flex: 1, paddingTop: '0' }}>
                    <div style={{ fontSize: '13px', fontWeight: '500', color: 'var(--color-text)', marginBottom: '4px' }}>{e.description}</div>
                    <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>Gatilho: {e.trigger}</div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {/* Action Playbook */}
      {report.action_playbook && (
        <div>
          <SectionTitle>Plano de Ação</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '12px' }}>
            <PlaybookColumn title="90 Dias" items={report.action_playbook.horizon_90_days} accent="var(--color-success)" />
            <PlaybookColumn title="1 Ano" items={report.action_playbook.horizon_1_year} accent="var(--color-accent)" />
            <PlaybookColumn title="3 Anos" items={report.action_playbook.horizon_3_years} accent="oklch(0.6 0.15 280)" />
            <PlaybookColumn title="Vitórias Rápidas" items={report.action_playbook.quick_wins} accent="var(--color-warning)" />
            <PlaybookColumn title="Riscos Críticos" items={report.action_playbook.critical_risks} accent="var(--color-danger)" />
          </div>
        </div>
      )}

      {/* Fracture Events */}
      {report.fracture_events?.length > 0 && (
        <div>
          <SectionTitle>Eventos de Ruptura</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {report.fracture_events.map((fe, i) => (
              <div key={i} style={{ display: 'flex', gap: '12px', alignItems: 'center', padding: '12px 16px', background: 'var(--color-surface)', borderRadius: '8px', border: `1px solid ${fe.accepted ? 'var(--color-success)' : 'var(--color-border)'}` }}>
                <div style={{ fontSize: '12px', fontWeight: '700', color: 'var(--color-text-muted)', minWidth: '60px' }}>Rodada {fe.round}</div>
                <div style={{ flex: 1, fontSize: '13px', color: 'var(--color-text)' }}>
                  <span style={{ color: 'var(--color-text-muted)' }}>{fe.proposed_by.slice(0, 8)} → </span>
                  {fe.accepted ? fe.proposal.new_description : `Alteração rejeitada em ${fe.proposal.original_rule_id}`}
                </div>
                <span style={{ padding: '2px 8px', borderRadius: '20px', fontSize: '11px', fontWeight: '700', background: fe.accepted ? 'var(--color-success)22' : 'var(--color-border)', color: fe.accepted ? 'var(--color-success)' : 'var(--color-text-muted)' }}>
                  {fe.accepted ? 'ACEITO' : 'REJEITADO'}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Watermark */}
      <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', textAlign: 'center', paddingBottom: '16px' }}>
        {report.watermark.notice}
      </div>
    </div>
  )
}

function PlaybookColumn({ title, items, accent }: { title: string; items?: string[]; accent: string }) {
  if (!items?.length) return null
  return (
    <Card>
      <div style={{ fontSize: '12px', fontWeight: '700', color: accent, marginBottom: '10px' }}>{title}</div>
      <ul style={{ margin: 0, padding: '0 0 0 16px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
        {items.map((item, i) => (
          <li key={i} style={{ fontSize: '13px', color: 'var(--color-text)' }}>{item}</li>
        ))}
      </ul>
    </Card>
  )
}

function probColor(p: number): string {
  if (p >= 0.7) return 'var(--color-danger)'
  if (p >= 0.5) return '#f97316'
  if (p >= 0.3) return 'var(--color-warning)'
  return 'var(--color-success)'
}

async function downloadFile(url: string, filename: string) {
  const res = await fetch(url)
  if (!res.ok) return
  const blob = await res.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = filename
  a.click()
  URL.revokeObjectURL(a.href)
}

const backBtnStyle: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-text-muted)', fontSize: '13px',
  cursor: 'pointer', padding: '0', display: 'flex', alignItems: 'center', gap: '4px',
}

const exportBtnStyle: React.CSSProperties = {
  padding: '9px 14px', borderRadius: '8px', border: '1px solid var(--color-border)',
  background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px',
  fontWeight: '600', cursor: 'pointer', whiteSpace: 'nowrap',
}
