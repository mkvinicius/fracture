import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface Sim { id: string; question: string; status: string; department: string; rounds: number; created_at: number; duration_ms?: number }

export default function SimulationsPage({ onNavigate }: { onNavigate: (p: Page, simId?: string, simIds?: string[]) => void }) {
  const [sims, setSims] = useState<Sim[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<Set<string>>(new Set())

  useEffect(() => {
    fetch('/api/v1/simulations').then(r => r.json()).then(d => { setSims(d ?? []); setLoading(false) }).catch(() => setLoading(false))
  }, [])

  const statusColor = (s: string) => s === 'complete' ? 'var(--color-success)' : s === 'running' ? 'var(--color-accent)' : s === 'failed' ? 'var(--color-danger)' : 'var(--color-warning)'

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
              {sim.status === 'complete' && (
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
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexShrink: 0 }}>
                <div style={{ fontSize: '11px', padding: '4px 10px', borderRadius: '6px', background: `${statusColor(sim.status)}22`, color: statusColor(sim.status), fontWeight: '600' }}>{sim.status}</div>
                {sim.status === 'complete' && (
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
