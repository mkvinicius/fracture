import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface Sim { id: string; question: string; status: string; department: string; rounds: number; created_at: number; duration_ms?: number }

export default function SimulationsPage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const [sims, setSims] = useState<Sim[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/simulations').then(r => r.json()).then(d => { setSims(d); setLoading(false) }).catch(() => setLoading(false))
  }, [])

  const statusColor = (s: string) => s === 'complete' ? 'var(--color-success)' : s === 'running' ? 'var(--color-accent)' : s === 'failed' ? 'var(--color-danger)' : 'var(--color-warning)'

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Simulation History</h1>
          <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>All past and running simulations</p>
        </div>
        <button onClick={() => onNavigate('new-simulation')} style={{ padding: '9px 18px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>+ New</button>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px', color: 'var(--color-text-muted)' }}>Loading...</div>
      ) : sims.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '60px', color: 'var(--color-text-muted)', background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)' }}>
          <div style={{ fontSize: '32px', marginBottom: '12px' }}>◎</div>
          <div style={{ fontWeight: '600', marginBottom: '6px' }}>No simulations yet</div>
          <div style={{ fontSize: '13px' }}>Run your first simulation to see results here</div>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {sims.map(sim => (
            <div key={sim.id} style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '16px' }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: '14px', fontWeight: '500', color: 'var(--color-text)', marginBottom: '4px' }}>{sim.question}</div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>
                  {sim.department} · {sim.rounds} rounds
                  {sim.duration_ms ? ` · ${(sim.duration_ms / 1000).toFixed(1)}s` : ''}
                  · {new Date(sim.created_at * 1000).toLocaleDateString()}
                </div>
              </div>
              <div style={{ fontSize: '11px', padding: '4px 10px', borderRadius: '6px', background: `${statusColor(sim.status)}22`, color: statusColor(sim.status), fontWeight: '600' }}>{sim.status}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
