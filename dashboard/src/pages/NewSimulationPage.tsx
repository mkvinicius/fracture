import { useState } from 'react'
import { type Page } from '../App'

const TEMPLATES = [
  { id: 'competitor-free', label: 'Competitor goes free', question: 'If our main competitor offered their core product for free tomorrow, how would the market react in 12 months?' },
  { id: 'new-entrant',     label: 'Big tech enters market', question: 'If a major tech company (Google, Apple, or Amazon) entered our market with unlimited resources, which rules would break first?' },
  { id: 'regulation',      label: 'Regulatory disruption', question: 'If a new regulation forced us to change our core business model, what would the market look like in 18 months?' },
  { id: 'custom',          label: 'Custom question', question: '' },
]

const DEPARTMENTS = ['Marketing', 'HR', 'Finance', 'Product', 'Sales', 'Operations', 'Strategy']
const ROUNDS = [10, 15, 20, 30]

export default function NewSimulationPage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const [template, setTemplate] = useState(TEMPLATES[0])
  const [question, setQuestion] = useState(TEMPLATES[0].question)
  const [department, setDepartment] = useState('Strategy')
  const [rounds, setRounds] = useState(20)
  const [context, setContext] = useState('')
  const [running, setRunning] = useState(false)
  const [error, setError] = useState('')

  async function handleRun() {
    if (!question.trim()) return
    setRunning(true)
    setError('')
    try {
      const res = await fetch('/api/simulations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question, department, rounds, context })
      })
      const data = await res.json()
      if (data.id) {
        onNavigate('simulations')
      } else {
        setError(data.error || 'Failed to start simulation')
        setRunning(false)
      }
    } catch {
      setError('Could not connect to FRACTURE engine. Make sure the server is running.')
      setRunning(false)
    }
  }

  const Input: React.CSSProperties = {
    width: '100%', padding: '10px 14px', borderRadius: '8px',
    border: '1px solid var(--color-border)', background: 'var(--color-surface-2)',
    color: 'var(--color-text)', fontSize: '14px', outline: 'none', boxSizing: 'border-box' as const,
  }

  return (
    <div style={{ padding: '32px', maxWidth: '720px' }}>
      <div style={{ marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>New Simulation</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>Define your strategic question. FRACTURE will simulate how the market responds.</p>
      </div>

      {/* Templates */}
      <div style={{ marginBottom: '24px' }}>
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '10px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Quick Templates</label>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
          {TEMPLATES.map(t => (
            <button key={t.id} onClick={() => { setTemplate(t); if (t.question) setQuestion(t.question) }}
              style={{ padding: '10px 14px', borderRadius: '8px', border: `1px solid ${template.id === t.id ? 'var(--color-accent)' : 'var(--color-border)'}`, background: template.id === t.id ? 'oklch(0.65 0.22 30 / 0.1)' : 'var(--color-surface-2)', color: template.id === t.id ? 'var(--color-accent)' : 'var(--color-text-muted)', fontSize: '13px', cursor: 'pointer', textAlign: 'left' as const }}>
              {t.label}
            </button>
          ))}
        </div>
      </div>

      {/* Question */}
      <div style={{ marginBottom: '20px' }}>
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Strategic Question *</label>
        <textarea style={{ ...Input, height: '100px', resize: 'vertical' }} value={question} onChange={e => setQuestion(e.target.value)} placeholder="What strategic question do you want to simulate?" />
      </div>

      {/* Department + Rounds */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '20px' }}>
        <div>
          <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Department</label>
          <select style={{ ...Input }} value={department} onChange={e => setDepartment(e.target.value)}>
            {DEPARTMENTS.map(d => <option key={d} value={d}>{d}</option>)}
          </select>
        </div>
        <div>
          <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Simulation Rounds</label>
          <select style={{ ...Input }} value={rounds} onChange={e => setRounds(Number(e.target.value))}>
            {ROUNDS.map(r => <option key={r} value={r}>{r} rounds (~{Math.round(r * 1.5)}s)</option>)}
          </select>
        </div>
      </div>

      {/* Context */}
      <div style={{ marginBottom: '28px' }}>
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Additional Context (optional)</label>
        <textarea style={{ ...Input, height: '80px', resize: 'vertical' }} value={context} onChange={e => setContext(e.target.value)} placeholder="Paste relevant news, competitor info, or market data to enrich the simulation..." />
      </div>

      {error && <div style={{ color: 'var(--color-danger)', fontSize: '13px', marginBottom: '16px', padding: '12px', borderRadius: '8px', background: 'oklch(0.60 0.22 25 / 0.1)', border: '1px solid oklch(0.60 0.22 25 / 0.3)' }}>{error}</div>}

      <div style={{ display: 'flex', gap: '12px' }}>
        <button onClick={() => onNavigate('home')} style={{ padding: '10px 20px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '14px', cursor: 'pointer' }}>Cancel</button>
        <button onClick={handleRun} disabled={!question.trim() || running}
          style={{ padding: '10px 28px', borderRadius: '8px', border: 'none', background: running ? 'var(--color-border)' : 'var(--color-accent)', color: '#fff', fontSize: '14px', fontWeight: '600', cursor: running ? 'not-allowed' : 'pointer' }}>
          {running ? '⟳ Starting...' : '◈ Run Simulation'}
        </button>
      </div>
    </div>
  )
}
