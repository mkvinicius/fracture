import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface Sim { id: string; question: string; status: string; created_at: number }

export default function HomePage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const [sims, setSims] = useState<Sim[]>([])
  const [company, setCompany] = useState<{ name: string; sector: string } | null>(null)

  useEffect(() => {
    fetch('/api/simulations').then(r => r.json()).then(setSims).catch(() => {})
    fetch('/api/company').then(r => r.json()).then(setCompany).catch(() => {})
  }, [])

  const Card: React.CSSProperties = {
    background: 'var(--color-surface)', borderRadius: '12px',
    border: '1px solid var(--color-border)', padding: '24px',
  }

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      {/* Header */}
      <div style={{ marginBottom: '32px' }}>
        <h1 style={{ margin: 0, fontSize: '22px', fontWeight: '700', color: 'var(--color-text)' }}>
          {company ? `${company.name} — ` : ''}<span style={{ color: 'var(--color-accent)' }}>FRACTURE</span> Dashboard
        </h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>
          Simulate market ruptures. Discover what breaks before it breaks you.
        </p>
      </div>

      {/* Quick action */}
      <div style={{ ...Card, background: 'oklch(0.65 0.22 30 / 0.08)', border: '1px solid oklch(0.65 0.22 30 / 0.25)', marginBottom: '24px', cursor: 'pointer' }}
        onClick={() => onNavigate('new-simulation')}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <div style={{ fontWeight: '600', fontSize: '15px', color: 'var(--color-text)', marginBottom: '4px' }}>Run a New Simulation</div>
            <div style={{ fontSize: '13px', color: 'var(--color-text-muted)' }}>Ask a strategic question. Watch the market simulate itself.</div>
          </div>
          <div style={{ fontSize: '28px', color: 'var(--color-accent)' }}>◈</div>
        </div>
      </div>

      {/* Stats row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px', marginBottom: '24px' }}>
        {[
          { label: 'Total Simulations', value: sims.length },
          { label: 'Fracture Events', value: sims.filter(s => s.status === 'complete').length },
          { label: 'Archetypes Active', value: 12 },
        ].map(stat => (
          <div key={stat.label} style={Card}>
            <div style={{ fontSize: '28px', fontWeight: '700', color: 'var(--color-accent)', lineHeight: 1 }}>{stat.value}</div>
            <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '4px' }}>{stat.label}</div>
          </div>
        ))}
      </div>

      {/* Recent simulations */}
      <div style={Card}>
        <div style={{ fontWeight: '600', fontSize: '14px', color: 'var(--color-text)', marginBottom: '16px' }}>Recent Simulations</div>
        {sims.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '32px', color: 'var(--color-text-muted)', fontSize: '13px' }}>
            No simulations yet. Run your first one to see results here.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {sims.slice(0, 5).map(sim => (
              <div key={sim.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px', borderRadius: '8px', background: 'var(--color-surface-2)' }}>
                <div style={{ fontSize: '13px', color: 'var(--color-text)', flex: 1, marginRight: '16px' }}>{sim.question}</div>
                <div style={{ fontSize: '11px', padding: '3px 8px', borderRadius: '4px', background: sim.status === 'complete' ? 'oklch(0.65 0.18 145 / 0.15)' : 'oklch(0.75 0.18 85 / 0.15)', color: sim.status === 'complete' ? 'var(--color-success)' : 'var(--color-warning)' }}>{sim.status}</div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
