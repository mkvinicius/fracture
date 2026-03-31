import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface Schedule { id: string; question: string; department: string; rounds: number; interval_h: number; enabled: boolean; last_run_at?: number; next_run_at: number; created_at: number }

const INTERVAL_LABELS: Record<number, string> = { 24: 'Diário', 168: 'Semanal', 720: 'Mensal' }

export default function SchedulesPage({ onNavigate: _onNavigate }: { onNavigate: (p: Page) => void }) {
  const [schedules, setSchedules] = useState<Schedule[]>([])
  const [loading, setLoading] = useState(true)
  const [question, setQuestion] = useState('')
  const [department, setDepartment] = useState('Strategy')
  const [rounds, setRounds] = useState(20)
  const [intervalH, setIntervalH] = useState(168)
  const [creating, setCreating] = useState(false)
  const [showForm, setShowForm] = useState(false)

  const load = () => fetch('/api/v1/schedules').then(r => r.json()).then(d => { setSchedules(d ?? []); setLoading(false) }).catch(() => setLoading(false))

  useEffect(() => { load() }, [])

  async function create() {
    if (!question.trim()) return
    setCreating(true)
    await fetch('/api/v1/schedules', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, department, rounds, interval_h: intervalH })
    })
    setQuestion(''); setShowForm(false); setCreating(false); load()
  }

  async function toggle(id: string, enabled: boolean) {
    await fetch(`/api/v1/schedules/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: !enabled })
    })
    load()
  }

  async function remove(id: string) {
    if (!window.confirm('Excluir este agendamento?')) return
    await fetch(`/api/v1/schedules/${id}`, { method: 'DELETE' })
    load()
  }

  const Input: React.CSSProperties = { width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'var(--color-surface-2)', color: 'var(--color-text)', fontSize: '13px', outline: 'none', boxSizing: 'border-box' as const }

  return (
    <div style={{ padding: '32px', maxWidth: '860px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Simulações Agendadas</h1>
          <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>FRACTURE monitora seu mercado automaticamente e gera relatórios periódicos</p>
        </div>
        <button onClick={() => setShowForm(!showForm)} style={{ padding: '9px 18px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>
          + Agendar
        </button>
      </div>

      {showForm && (
        <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '20px', marginBottom: '24px' }}>
          <div style={{ fontSize: '14px', fontWeight: '600', marginBottom: '16px' }}>Nova Simulação Agendada</div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Pergunta estratégica *</label>
            <textarea style={{ ...Input, height: '80px', resize: 'vertical' }} value={question} onChange={e => setQuestion(e.target.value)} placeholder="Ex: Como o mercado de delivery está evoluindo?" />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '12px', marginBottom: '16px' }}>
            <div>
              <label style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Departamento</label>
              <select style={Input} value={department} onChange={e => setDepartment(e.target.value)}>
                {['Strategy', 'Marketing', 'Finance', 'Product', 'Operations'].map(d => <option key={d}>{d}</option>)}
              </select>
            </div>
            <div>
              <label style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Rodadas</label>
              <select style={Input} value={rounds} onChange={e => setRounds(Number(e.target.value))}>
                {[10, 20, 30].map(r => <option key={r} value={r}>{r} rodadas</option>)}
              </select>
            </div>
            <div>
              <label style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Frequência</label>
              <select style={Input} value={intervalH} onChange={e => setIntervalH(Number(e.target.value))}>
                <option value={24}>Diário</option>
                <option value={168}>Semanal</option>
                <option value={720}>Mensal</option>
              </select>
            </div>
          </div>
          <div style={{ display: 'flex', gap: '8px' }}>
            <button onClick={create} disabled={creating || !question.trim()} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>
              {creating ? 'Criando...' : 'Criar Agendamento'}
            </button>
            <button onClick={() => setShowForm(false)} style={{ padding: '8px 16px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '13px', cursor: 'pointer' }}>Cancelar</button>
          </div>
        </div>
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px', color: 'var(--color-text-muted)' }}>Carregando...</div>
      ) : schedules.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '60px', background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', color: 'var(--color-text-muted)' }}>
          <div style={{ fontSize: '28px', marginBottom: '12px' }}>⏱</div>
          <div style={{ fontWeight: '600', marginBottom: '6px' }}>Nenhum agendamento</div>
          <div style={{ fontSize: '13px' }}>Agende uma simulação recorrente para monitorar seu mercado automaticamente</div>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {schedules.map(s => (
            <div key={s.id} style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '16px', opacity: s.enabled ? 1 : 0.5 }}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: '14px', fontWeight: '500', color: 'var(--color-text)', marginBottom: '4px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{s.question}</div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>
                  {s.department} · {s.rounds} rodadas · {INTERVAL_LABELS[s.interval_h] ?? `${s.interval_h}h`}
                  {s.last_run_at && ` · Última: ${new Date(s.last_run_at * 1000).toLocaleDateString()}`}
                  {` · Próxima: ${new Date(s.next_run_at * 1000).toLocaleDateString()}`}
                </div>
              </div>
              <div style={{ display: 'flex', gap: '8px', flexShrink: 0 }}>
                <span style={{ fontSize: '11px', padding: '3px 8px', borderRadius: '6px', background: s.enabled ? 'oklch(0.65 0.22 30 / 0.15)' : 'oklch(0.4 0 0 / 0.1)', color: s.enabled ? 'var(--color-accent)' : 'var(--color-text-muted)', fontWeight: '600' }}>
                  {s.enabled ? 'Ativo' : 'Pausado'}
                </span>
                <button onClick={() => toggle(s.id, s.enabled)} style={{ padding: '5px 10px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px', cursor: 'pointer' }}>
                  {s.enabled ? 'Pausar' : 'Ativar'}
                </button>
                <button onClick={() => remove(s.id)} style={{ padding: '5px 10px', borderRadius: '6px', border: '1px solid var(--color-danger)', background: 'transparent', color: 'var(--color-danger)', fontSize: '12px', cursor: 'pointer' }}>✕</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
