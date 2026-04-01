import { useEffect, useRef, useState } from 'react'
import { type Page } from '../App'

interface Sim { id: string; question: string; status: string; department: string; rounds: number; created_at: number; duration_ms?: number; current_round?: number; current_tension?: number; fracture_count?: number; last_agent_name?: string }
interface LiveProgress { current_round: number; current_tension: number; fracture_count: number; last_agent_name: string }

function useLiveProgress(sims: Sim[], onUpdate: (id: string, p: LiveProgress) => void) {
  const esRefs = useRef<Record<string, EventSource>>({})

  useEffect(() => {
    const running = sims.filter(s => s.status === 'running' || s.status === 'researching')
    const activeIds = new Set(running.map(s => s.id))

    // Close SSE for sims no longer running
    Object.keys(esRefs.current).forEach(id => {
      if (!activeIds.has(id)) {
        esRefs.current[id].close()
        delete esRefs.current[id]
      }
    })

    // Open SSE for new running sims
    running.forEach(sim => {
      if (esRefs.current[sim.id]) return
      const es = new EventSource(`/api/v1/simulations/${sim.id}/events`)
      es.onmessage = (e) => {
        try { onUpdate(sim.id, JSON.parse(e.data)) } catch {}
      }
      es.onerror = () => { es.close(); delete esRefs.current[sim.id] }
      esRefs.current[sim.id] = es
    })

    return () => {
      Object.values(esRefs.current).forEach(es => es.close())
      esRefs.current = {}
    }
  }, [sims.map(s => s.id + s.status).join(',')])
}

export default function SimulationsPage({ onNavigate }: { onNavigate: (p: Page, simId?: string, simIds?: string[]) => void }) {
  const [sims, setSims] = useState<Sim[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<Set<string>>(new Set())

  useEffect(() => {
    fetch('/api/v1/simulations').then(r => r.json()).then(d => { setSims(d ?? []); setLoading(false) }).catch(() => setLoading(false))
  }, [])

  useLiveProgress(sims, (id, p) => {
    setSims(prev => prev.map(s => s.id === id ? { ...s, status: 'running', current_round: p.current_round, current_tension: p.current_tension, fracture_count: p.fracture_count, last_agent_name: p.last_agent_name } : s))
  })

  const statusColor = (s: string) => s === 'done' ? 'var(--color-success)' : s === 'running' ? 'var(--color-accent)' : s === 'error' ? 'var(--color-danger)' : 'var(--color-warning)'

  const toggleSelect = (id: string) => setSelected(prev => {
    const next = new Set(prev)
    if (next.has(id)) { next.delete(id) } else { next.add(id) }
    return next
  })

  const selectedArr = Array.from(selected)
  const canCompare = selectedArr.length >= 2 && selectedArr.length <= 5

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Histórico de Simulações</h1>
          <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>Todas as simulações passadas e em andamento</p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          {canCompare && (
            <button
              onClick={() => onNavigate('comparison', undefined, selectedArr)}
              style={{ padding: '9px 18px', borderRadius: '8px', border: '1px solid var(--color-accent)', background: 'transparent', color: 'var(--color-accent)', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}
            >
              Comparar Selecionadas ({selectedArr.length})
            </button>
          )}
          <button onClick={() => onNavigate('new-simulation')} style={{ padding: '9px 18px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>+ Nova</button>
        </div>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px', color: 'var(--color-text-muted)' }}>Carregando...</div>
      ) : sims.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '60px', color: 'var(--color-text-muted)', background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)' }}>
          <div style={{ fontSize: '32px', marginBottom: '12px' }}>◎</div>
          <div style={{ fontWeight: '600', marginBottom: '6px' }}>Nenhuma simulação ainda</div>
          <div style={{ fontSize: '13px' }}>Inicie sua primeira simulação para ver os resultados aqui</div>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {sims.map(sim => (
            <div key={sim.id} style={{ background: 'var(--color-surface)', borderRadius: '10px', border: `1px solid ${selected.has(sim.id) ? 'var(--color-accent)' : 'var(--color-border)'}`, padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '14px' }}>
              {sim.status === 'done' && (
                <input
                  type="checkbox"
                  checked={selected.has(sim.id)}
                  onChange={() => toggleSelect(sim.id)}
                  style={{ width: '16px', height: '16px', cursor: 'pointer', accentColor: 'var(--color-accent)', flexShrink: 0 }}
                />
              )}
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: '14px', fontWeight: '500', color: 'var(--color-text)', marginBottom: '4px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{sim.question}</div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>
                  {sim.department} · {sim.rounds} rodadas
                  {sim.duration_ms ? ` · ${(sim.duration_ms / 1000).toFixed(1)}s` : ''}
                  · {new Date(sim.created_at * 1000).toLocaleDateString()}
                </div>
                {(sim.status === 'running' || sim.status === 'researching') && sim.current_round != null && (
                  <div style={{ marginTop: '6px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: 'var(--color-text-muted)', marginBottom: '3px' }}>
                      <span>Rodada {sim.current_round}/{sim.rounds}{sim.last_agent_name ? ` · ${sim.last_agent_name}` : ''}</span>
                      {sim.current_tension != null && <span style={{ color: sim.current_tension >= 0.7 ? 'var(--color-danger)' : sim.current_tension >= 0.5 ? '#f97316' : 'var(--color-text-muted)' }}>Tensão {Math.round(sim.current_tension * 100)}%{sim.fracture_count ? ` · ${sim.fracture_count} fratura(s)` : ''}</span>}
                    </div>
                    <div style={{ height: '4px', borderRadius: '2px', background: 'var(--color-background)', overflow: 'hidden' }}>
                      <div style={{ height: '100%', width: `${((sim.current_round ?? 0) / sim.rounds) * 100}%`, background: 'var(--color-accent)', borderRadius: '2px', transition: 'width 0.5s' }} />
                    </div>
                  </div>
                )}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexShrink: 0 }}>
                <div style={{ fontSize: '11px', padding: '4px 10px', borderRadius: '6px', background: `${statusColor(sim.status)}22`, color: statusColor(sim.status), fontWeight: '600' }}>{sim.status}</div>
                {(sim.status === 'running' || sim.status === 'researching') && (
                  <button
                    onClick={() => onNavigate('live-activity', sim.id)}
                    style={{ padding: '5px 12px', borderRadius: '6px', border: '1px solid var(--color-danger)', background: 'var(--color-danger)11', color: 'var(--color-danger)', fontSize: '12px', cursor: 'pointer', fontWeight: '600', display: 'flex', alignItems: 'center', gap: '5px' }}
                  >
                    <span className="animate-pulse-dot" style={{ display: 'inline-block', width: '6px', height: '6px', borderRadius: '50%', background: 'var(--color-danger)' }} />
                    Ao Vivo
                  </button>
                )}
                {sim.status === 'done' && (
                  <>
                    <button
                      onClick={() => onNavigate('result', sim.id)}
                      style={{ padding: '5px 12px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text)', fontSize: '12px', cursor: 'pointer', fontWeight: '500' }}
                    >
                      Ver Resultado
                    </button>
                    <button
                      onClick={() => onNavigate('convergence', sim.id)}
                      style={{ padding: '5px 12px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px', cursor: 'pointer', fontWeight: '500' }}
                    >
                      Ver Convergência
                    </button>
                    <button
                      onClick={() => onNavigate('replay', sim.id)}
                      style={{ padding: '5px 12px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px', cursor: 'pointer', fontWeight: '500' }}
                    >
                      ▶ Replay
                    </button>
                    <button
                      onClick={() => onNavigate('feedback', sim.id)}
                      style={{ padding: '5px 12px', borderRadius: '6px', border: '1px solid var(--color-accent)', background: 'transparent', color: 'var(--color-accent)', fontSize: '12px', cursor: 'pointer', fontWeight: '500' }}
                    >
                      Dar Feedback
                    </button>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
