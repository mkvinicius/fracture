import { useState } from 'react'
import { type Page } from '../App'

const TEMPLATES = [
  { id: 'competitor-free', label: 'Concorrente vai de graça', question: 'Se nosso principal concorrente oferecesse seu produto principal gratuitamente amanhã, como o mercado reagiria em 12 meses?' },
  { id: 'new-entrant',     label: 'Big tech entra no mercado', question: 'Se uma grande empresa de tecnologia (Google, Apple ou Amazon) entrasse em nosso mercado com recursos ilimitados, quais regras quebrariam primeiro?' },
  { id: 'regulation',      label: 'Disrupção regulatória', question: 'Se uma nova regulação nos forçasse a mudar nosso modelo de negócios principal, como seria o mercado em 18 meses?' },
  { id: 'custom',          label: 'Pergunta personalizada', question: '' },
]

const DEPARTMENTS = ['Marketing', 'HR', 'Finance', 'Product', 'Sales', 'Operations', 'Strategy']
const ROUNDS = [10, 15, 20, 30, 40]

const URL_PLACEHOLDERS: Record<string, string> = {
  website:   'https://yourcompany.com',
  linkedin:  'https://linkedin.com/company/yourcompany',
  instagram: 'https://instagram.com/yourcompany',
  twitter:   'https://x.com/yourcompany',
  facebook:  'https://facebook.com/yourcompany',
  youtube:   'https://youtube.com/@yourcompany',
}

const URL_ICONS: Record<string, string> = {
  website:   '🌐',
  linkedin:  '💼',
  instagram: '📸',
  twitter:   '𝕏',
  facebook:  '📘',
  youtube:   '▶️',
}

type UrlEntry = { type: string; value: string }

export default function NewSimulationPage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const [template, setTemplate] = useState(TEMPLATES[0])
  const [question, setQuestion] = useState(TEMPLATES[0].question)
  const [department, setDepartment] = useState('Strategy')
  const [rounds, setRounds] = useState(40)
  const [context, setContext] = useState('')
  const [running, setRunning] = useState(false)
  const [simPhase, setSimPhase] = useState<'idle' | 'researching' | 'running'>('idle')
  const [researchSources, setResearchSources] = useState(0)
  const [error, setError] = useState('')
  const [showUrlSection, setShowUrlSection] = useState(false)
  const [extracting, setExtracting] = useState(false)
  const [extractedSummary, setExtractedSummary] = useState('')
  const [urlEntries, setUrlEntries] = useState<UrlEntry[]>([
    { type: 'website', value: '' },
    { type: 'linkedin', value: '' },
  ])

  function addUrlEntry() {
    if (urlEntries.length < 10) {
      const usedTypes = urlEntries.map(u => u.type)
      const nextType = Object.keys(URL_PLACEHOLDERS).find(t => !usedTypes.includes(t)) || 'website'
      setUrlEntries([...urlEntries, { type: nextType, value: '' }])
    }
  }

  function removeUrlEntry(idx: number) {
    setUrlEntries(urlEntries.filter((_, i) => i !== idx))
  }

  function updateUrlEntry(idx: number, field: 'type' | 'value', val: string) {
    setUrlEntries(urlEntries.map((u, i) => i === idx ? { ...u, [field]: val } : u))
  }

  async function handleExtractContext() {
    const urls = urlEntries.map(u => u.value).filter(v => v.trim() !== '')
    if (urls.length === 0) return
    setExtracting(true)
    setExtractedSummary('')
    try {
      const res = await fetch('/api/v1/extract-context', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ urls })
      })
      const data = await res.json()
      if (data.summary) {
        setExtractedSummary(data.summary)
      }
    } catch {
      // silent fail
    } finally {
      setExtracting(false)
    }
  }

  async function handleRun() {
    if (!question.trim()) return
    setRunning(true)
    setSimPhase('researching')
    setResearchSources(0)
    setError('')
    const urls = urlEntries.map(u => u.value).filter(v => v.trim() !== '')
    try {
      const res = await fetch('/api/v1/simulations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question, department, rounds, context, urls })
      })
      const data = await res.json()
      if (data.id) {
        // Poll for status to show research → running transition
        const poll = setInterval(async () => {
          try {
            const sr = await fetch(`/api/v1/simulations/${data.id}`)
            const sd = await sr.json()
            if (sd.research_sources) setResearchSources(sd.research_sources)
            if (sd.status === 'running') setSimPhase('running')
            if (sd.status === 'done' || sd.status === 'error') {
              clearInterval(poll)
              onNavigate('simulations')
            }
          } catch { clearInterval(poll); onNavigate('simulations') }
        }, 2000)
      } else {
        setError(data.error || 'Falha ao iniciar a simulação')
        setRunning(false)
        setSimPhase('idle')
      }
    } catch {
      setError('Não foi possível conectar ao motor FRACTURE. Verifique se o servidor está em execução.')
      setRunning(false)
      setSimPhase('idle')
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
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Nova Simulação</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>Defina sua pergunta estratégica. FRACTURE simulará como o mercado responde.</p>
      </div>

      {/* Templates */}
      <div style={{ marginBottom: '24px' }}>
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '10px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Modelos Rápidos</label>
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
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Pergunta Estratégica *</label>
        <textarea style={{ ...Input, height: '100px', resize: 'vertical' }} value={question} onChange={e => setQuestion(e.target.value)} placeholder="Qual pergunta estratégica você quer simular?" />
      </div>

      {/* Department + Rounds */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '20px' }}>
        <div>
          <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Departamento</label>
          <select style={{ ...Input }} value={department} onChange={e => setDepartment(e.target.value)}>
            {DEPARTMENTS.map(d => <option key={d} value={d}>{d}</option>)}
          </select>
        </div>
        <div>
          <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Rodadas da Simulação</label>
          <select style={{ ...Input }} value={rounds} onChange={e => setRounds(Number(e.target.value))}>
            {ROUNDS.map(r => <option key={r} value={r}>{r} rodadas (~{Math.round(r * 1.5)}s)</option>)}
          </select>
        </div>
      </div>

      {/* Company URLs Section */}
      <div style={{ marginBottom: '20px', borderRadius: '10px', border: '1px solid var(--color-border)', overflow: 'hidden' }}>
        <button
          onClick={() => setShowUrlSection(!showUrlSection)}
          style={{ width: '100%', padding: '12px 16px', background: 'var(--color-surface-2)', border: 'none', color: 'var(--color-text)', fontSize: '13px', fontWeight: '600', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'space-between', textAlign: 'left' as const }}
        >
          <span>🔗 Contexto da Empresa — Site & Redes Sociais <span style={{ color: 'var(--color-text-muted)', fontWeight: '400' }}>(opcional, melhora a precisão da simulação)</span></span>
          <span style={{ color: 'var(--color-text-muted)', fontSize: '11px' }}>{showUrlSection ? '▲ ocultar' : '▼ expandir'}</span>
        </button>

        {showUrlSection && (
          <div style={{ padding: '16px', borderTop: '1px solid var(--color-border)' }}>
            <p style={{ margin: '0 0 14px', fontSize: '12px', color: 'var(--color-text-muted)', lineHeight: '1.5' }}>
              Adicione o site e perfis de redes sociais da sua empresa. FRACTURE extrairá automaticamente informações públicas para tornar a simulação mais precisa e personalizada.
            </p>

            {urlEntries.map((entry, idx) => (
              <div key={idx} style={{ display: 'flex', gap: '8px', marginBottom: '10px', alignItems: 'center' }}>
                <select
                  value={entry.type}
                  onChange={e => updateUrlEntry(idx, 'type', e.target.value)}
                  style={{ ...Input, width: '130px', flexShrink: 0 }}
                >
                  {Object.keys(URL_PLACEHOLDERS).map(t => (
                    <option key={t} value={t}>{URL_ICONS[t]} {t.charAt(0).toUpperCase() + t.slice(1)}</option>
                  ))}
                </select>
                <input
                  type="url"
                  value={entry.value}
                  onChange={e => updateUrlEntry(idx, 'value', e.target.value)}
                  placeholder={URL_PLACEHOLDERS[entry.type]}
                  style={{ ...Input, flex: 1 }}
                />
                {urlEntries.length > 1 && (
                  <button onClick={() => removeUrlEntry(idx)} style={{ padding: '8px 10px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', cursor: 'pointer', fontSize: '14px', flexShrink: 0 }}>✕</button>
                )}
              </div>
            ))}

            <div style={{ display: 'flex', gap: '10px', marginTop: '12px' }}>
              {urlEntries.length < 10 && (
                <button onClick={addUrlEntry} style={{ padding: '8px 14px', borderRadius: '6px', border: '1px dashed var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px', cursor: 'pointer' }}>
                  + Adicionar outra URL
                </button>
              )}
              <button
                onClick={handleExtractContext}
                disabled={extracting || urlEntries.every(u => !u.value.trim())}
                style={{ padding: '8px 16px', borderRadius: '6px', border: '1px solid var(--color-accent)', background: 'oklch(0.65 0.22 30 / 0.1)', color: 'var(--color-accent)', fontSize: '12px', fontWeight: '600', cursor: extracting ? 'not-allowed' : 'pointer' }}
              >
                {extracting ? '⟳ Extraindo...' : '⚡ Visualizar contexto extraído'}
              </button>
            </div>

            {extractedSummary && (
              <div style={{ marginTop: '14px', padding: '12px', borderRadius: '8px', background: 'oklch(0.65 0.22 30 / 0.05)', border: '1px solid oklch(0.65 0.22 30 / 0.2)', fontSize: '12px', color: 'var(--color-text-muted)', maxHeight: '150px', overflowY: 'auto', whiteSpace: 'pre-wrap', fontFamily: 'monospace' }}>
                {extractedSummary}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Additional Context */}
      <div style={{ marginBottom: '28px' }}>
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Contexto Adicional (opcional)</label>
        <textarea style={{ ...Input, height: '80px', resize: 'vertical' }} value={context} onChange={e => setContext(e.target.value)} placeholder="Cole notícias relevantes, informações de concorrentes ou dados de mercado para enriquecer a simulação..." />
      </div>

      {error && <div style={{ color: 'var(--color-danger)', fontSize: '13px', marginBottom: '16px', padding: '12px', borderRadius: '8px', background: 'oklch(0.60 0.22 25 / 0.1)', border: '1px solid oklch(0.60 0.22 25 / 0.3)' }}>{error}</div>}

      {/* Progress indicator */}
      {running && (
        <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '10px', background: 'var(--color-surface-2)', border: '1px solid var(--color-border)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '10px' }}>
            <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: simPhase === 'researching' ? 'oklch(0.75 0.18 55)' : 'var(--color-accent)', animation: 'pulse 1.2s infinite' }} />
            <span style={{ fontSize: '13px', fontWeight: '600', color: 'var(--color-text)' }}>
              {simPhase === 'researching' ? '🔍 DeepSearch — Pesquisando contexto de mercado...' : '◈ FRACTURE — Executando simulação com 56 agentes...'}
            </span>
          </div>
          <div style={{ display: 'flex', gap: '20px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
            <span style={{ color: simPhase !== 'idle' ? 'oklch(0.75 0.18 55)' : 'var(--color-text-muted)' }}>✓ DeepSearch {simPhase === 'researching' ? 'executando...' : researchSources > 0 ? `— ${researchSources} fontes encontradas` : '— concluído'}</span>
            <span style={{ color: simPhase === 'running' ? 'var(--color-accent)' : 'var(--color-text-muted)' }}>{simPhase === 'running' ? '⟳ Simulação em andamento...' : '○ Simulação na fila'}</span>
          </div>
          <style>{`@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.3} }`}</style>
        </div>
      )}

      <div style={{ display: 'flex', gap: '12px' }}>
        <button onClick={() => onNavigate('home')} disabled={running} style={{ padding: '10px 20px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '14px', cursor: running ? 'not-allowed' : 'pointer' }}>Cancelar</button>
        <button onClick={handleRun} disabled={!question.trim() || running}
          style={{ padding: '10px 28px', borderRadius: '8px', border: 'none', background: running ? 'var(--color-border)' : 'var(--color-accent)', color: '#fff', fontSize: '14px', fontWeight: '600', cursor: running ? 'not-allowed' : 'pointer' }}>
          {running ? (simPhase === 'researching' ? '🔍 Pesquisando...' : '◈ Simulando...') : '◈ Iniciar Simulação'}
        </button>
      </div>
    </div>
  )
}
