import { useState } from 'react'
import { type Page } from '../App'

interface FeedbackForm {
  outcome: string
  predicted_fracture: string
  actual_outcome: string
  delta_score: number
  notes: string
}

export default function FeedbackPage({ simId, onNavigate }: { simId: string; onNavigate: (p: Page, simId?: string) => void }) {
  const [form, setForm] = useState<FeedbackForm>({
    outcome: 'accurate',
    predicted_fracture: '',
    actual_outcome: '',
    delta_score: 0,
    notes: '',
  })
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const set = <K extends keyof FeedbackForm>(k: K, v: FeedbackForm[K]) =>
    setForm(prev => ({ ...prev, [k]: v }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const res = await fetch(`/api/v1/simulations/${simId}/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || `HTTP ${res.status}`)
      }
      setSubmitted(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Falha ao enviar')
    } finally {
      setSubmitting(false)
    }
  }

  if (submitted) {
    return (
      <div style={{ padding: '32px', maxWidth: '640px' }}>
        <button onClick={() => onNavigate('simulations')} style={backBtnStyle}>← Voltar para simulações</button>
        <div style={{ marginTop: '40px', textAlign: 'center', padding: '48px', background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)' }}>
          <div style={{ fontSize: '32px', marginBottom: '16px' }}>✓</div>
          <div style={{ fontSize: '16px', fontWeight: '700', color: 'var(--color-text)', marginBottom: '8px' }}>Feedback enviado</div>
          <div style={{ fontSize: '13px', color: 'var(--color-text-muted)', marginBottom: '24px' }}>
            O modelo de calibração usará seu feedback para melhorar simulações futuras.
          </div>
          <div style={{ display: 'flex', gap: '10px', justifyContent: 'center' }}>
            <button onClick={() => onNavigate('result', simId)} style={primaryBtnStyle}>Ver Relatório</button>
            <button onClick={() => onNavigate('simulations')} style={secondaryBtnStyle}>Todas as Simulações</button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div style={{ padding: '32px', maxWidth: '640px' }}>
      <button onClick={() => onNavigate('result', simId)} style={backBtnStyle}>← Voltar ao relatório</button>

      <div style={{ marginTop: '16px', marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Dar Feedback</h1>
        <p style={{ margin: '6px 0 0', fontSize: '13px', color: 'var(--color-text-muted)' }}>
          Seu feedback calibra a precisão dos arquétipos para simulações futuras.
        </p>
      </div>

      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
        {/* Outcome radio */}
        <div>
          <label style={labelStyle}>Precisão geral</label>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {[
              { value: 'accurate', label: 'Preciso' },
              { value: 'partially_accurate', label: 'Parcialmente preciso' },
              { value: 'inaccurate', label: 'Impreciso' },
              { value: 'too_early', label: 'Cedo demais para dizer' },
            ].map(opt => (
              <label key={opt.value} style={{
                display: 'flex', alignItems: 'center', gap: '6px', padding: '8px 14px',
                borderRadius: '8px', border: `1px solid ${form.outcome === opt.value ? 'var(--color-accent)' : 'var(--color-border)'}`,
                background: form.outcome === opt.value ? 'var(--color-accent)15' : 'var(--color-surface)',
                cursor: 'pointer', fontSize: '13px', color: 'var(--color-text)',
              }}>
                <input
                  type="radio"
                  name="outcome"
                  value={opt.value}
                  checked={form.outcome === opt.value}
                  onChange={() => set('outcome', opt.value)}
                  style={{ display: 'none' }}
                />
                {opt.label}
              </label>
            ))}
          </div>
        </div>

        {/* Delta score slider */}
        <div>
          <label style={labelStyle}>
            Pontuação delta
            <span style={{ marginLeft: '8px', fontWeight: '700', color: deltaColor(form.delta_score) }}>
              {form.delta_score > 0 ? '+' : ''}{form.delta_score.toFixed(2)}
            </span>
          </label>
          <input
            type="range"
            min={-1}
            max={1}
            step={0.05}
            value={form.delta_score}
            onChange={e => set('delta_score', parseFloat(e.target.value))}
            style={{ width: '100%', accentColor: 'var(--color-accent)' }}
          />
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: 'var(--color-text-muted)', marginTop: '4px' }}>
            <span>-1.0 (completamente errado)</span>
            <span>0 (neutro)</span>
            <span>+1.0 (perfeito)</span>
          </div>
        </div>

        {/* Predicted fracture */}
        <div>
          <label style={labelStyle}>Ruptura prevista (o que a simulação antecipou)</label>
          <input
            type="text"
            value={form.predicted_fracture}
            onChange={e => set('predicted_fracture', e.target.value)}
            placeholder="ex.: Intervenção regulatória em precificação de IA"
            style={inputStyle}
          />
        </div>

        {/* Actual outcome */}
        <div>
          <label style={labelStyle}>Resultado real (o que realmente aconteceu)</label>
          <textarea
            value={form.actual_outcome}
            onChange={e => set('actual_outcome', e.target.value)}
            placeholder="Descreva o que realmente ocorreu no mercado..."
            rows={4}
            style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
          />
        </div>

        {/* Notes */}
        <div>
          <label style={labelStyle}>Observações (opcional)</label>
          <textarea
            value={form.notes}
            onChange={e => set('notes', e.target.value)}
            placeholder="Contexto adicional ou observações..."
            rows={3}
            style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
          />
        </div>

        {error && (
          <div style={{ padding: '12px 16px', borderRadius: '8px', background: 'var(--color-danger)15', border: '1px solid var(--color-danger)', color: 'var(--color-danger)', fontSize: '13px' }}>
            {error}
          </div>
        )}

        <div style={{ display: 'flex', gap: '10px' }}>
          <button type="submit" disabled={submitting} style={{ ...primaryBtnStyle, opacity: submitting ? 0.6 : 1 }}>
            {submitting ? 'Enviando...' : 'Enviar Feedback'}
          </button>
          <button type="button" onClick={() => onNavigate('result', simId)} style={secondaryBtnStyle}>Cancelar</button>
        </div>
      </form>
    </div>
  )
}

function deltaColor(v: number): string {
  if (v >= 0.5) return 'var(--color-success)'
  if (v >= 0) return 'var(--color-warning)'
  if (v >= -0.5) return '#f97316'
  return 'var(--color-danger)'
}

const backBtnStyle: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-text-muted)', fontSize: '13px',
  cursor: 'pointer', padding: '0',
}

const labelStyle: React.CSSProperties = {
  display: 'block', fontSize: '13px', fontWeight: '600', color: 'var(--color-text)', marginBottom: '8px',
}

const inputStyle: React.CSSProperties = {
  width: '100%', padding: '10px 14px', borderRadius: '8px',
  border: '1px solid var(--color-border)', background: 'var(--color-surface)',
  color: 'var(--color-text)', fontSize: '13px', boxSizing: 'border-box',
  outline: 'none',
}

const primaryBtnStyle: React.CSSProperties = {
  padding: '10px 20px', borderRadius: '8px', border: 'none',
  background: 'var(--color-accent)', color: '#fff', fontSize: '13px',
  fontWeight: '600', cursor: 'pointer',
}

const secondaryBtnStyle: React.CSSProperties = {
  padding: '10px 20px', borderRadius: '8px',
  border: '1px solid var(--color-border)', background: 'transparent',
  color: 'var(--color-text-muted)', fontSize: '13px', cursor: 'pointer',
}
