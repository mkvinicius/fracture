import { useState } from 'react'
import { type Page } from '../App'

const TEMPLATES = [
  { id: 'competitor-free', label: 'Competitor goes free', question: 'If our main competitor offered their core product for free tomorrow, how would the market react in 12 months?' },
  { id: 'new-entrant',     label: 'Big tech enters market', question: 'If a major tech company (Google, Apple, or Amazon) entered our market with unlimited resources, which rules would break first?' },
  { id: 'regulation',      label: 'Regulatory disruption', question: 'If a new regulation forced us to change our core business model, what would the market look like in 18 months?' },
  { id: 'custom',          label: 'Custom question', question: '' },
]

const DEPARTMENTS = ['Marketing', 'HR', 'Finance', 'Product', 'Sales', 'Operations', 'Strategy']
const ROUNDS = [10, 15, 20, 30, 40]

const INDUSTRIES = [
  { id: '', label: '— Geral (sem skill específica)', description: '' },
  { id: 'healthcare', label: '🏥 Healthcare & Life Sciences', description: 'Agentes especializados em regulação ANVISA, reembolso SUS/ANS, healthtechs, hospitais e pharma. Inclui Paul Farmer, Atul Gawande, Eric Topol e outros.' },
  { id: 'fintech', label: '💳 Fintech & Financial Services', description: 'Agentes especializados em Open Finance, PIX, regulação BACEN/CVM, neobanks e criptomoedas. Inclui David Vélez, Sébastian Mejía, Vitalik Buterin e outros.' },
  { id: 'retail', label: '🛒 Retail & Consumer', description: 'Agentes especializados em e-commerce, marketplaces, omnichannel, PROCON e consumer behavior. Inclui Marcos Galperin, Luiza Trajano, Abilio Diniz e outros.' },
  { id: 'legal', label: '⚖️ Legal & LegalTech', description: 'Agentes especializados em regulação OAB, automação jurídica, acesso à justiça e ética profissional. Inclui Daniel Kessler, Modesto Carvalhosa e outros.' },
  { id: 'education', label: '🎓 Education & EdTech', description: 'Agentes especializados em regulação MEC, EAD, microlearning, credencialismo e futuro do trabalho. Inclui Salman Khan, Daphne Koller e outros.' },
  { id: 'agro', label: '🌱 Agro & AgriTech', description: 'Agentes especializados em commodities, MAPA, ESG/desmatamento, precision agriculture e cadeia do agro. Inclui Marcos Jank, Blairo Maggi, Rachel Maia e outros.' },
  { id: 'construction', label: '🏗️ Construção Civil & PropTech', description: 'Agentes especializados em CAIXA/FGTS, MCMV, CREA, proptech e mercado imobiliário. Inclui Elie Horn, Eduardo Fischer, Patrícia Pereira e outros.' },
  { id: 'logistics', label: '🚚 Logística & Supply Chain', description: 'Agentes especializados em ANTT, last-mile, frete marketplace, cold chain e porto de Santos. Inclui Fábio Schvartsman, Paulo Guimarães, Cristina Palmaka e outros.' },
  { id: 'saas', label: '💻 SaaS & Tech B2B', description: 'Agentes especializados em ARR/NRR, vertical SaaS, AI coding tools, LGPD e enterprise GTM. Inclui Jason Lemkin, Henrique Dubugras, David Sacks e outros.' },
  { id: 'energy', label: '⚡ Energia & Utilities', description: 'Agentes especializados em ANEEL, solar distribuída, Petrobras, hidrogênio verde e transição energética. Inclui Jean-Paul Prates, Roberto Wajsman, Rodrigo Limp e outros.' },
  { id: 'manufacturing', label: '🏭 Indústria & Manufatura', description: 'Agentes especializados em Custo Brasil, Indústria 4.0, reshoring, BNDES e competição chinesa. Inclui Paulo Skaf, Klaus Schwab, Sergio Rial e outros.' },
  { id: 'media', label: '📺 Mídia & Entretenimento', description: 'Agentes especializados em streaming, creator economy, CONAR, AI content e publicidade digital. Inclui João Roberto Marinho, Reed Hastings, MrBeast e outros.' },
  { id: 'tourism', label: '✈️ Turismo & Hospitalidade', description: 'Agentes especializados em ANAC, OTAs, Airbnb, LATAM/Gol e experience economy. Inclui Brian Chesky, Guilherme Paulus, Chip Conley e outros.' },
]

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
  const [industry, setIndustry] = useState('')
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
      const res = await fetch('/api/extract-context', {
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
      const res = await fetch('/api/simulations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question, department, rounds, context, urls, industry })
      })
      const data = await res.json()
      if (data.id) {
        // Poll for status to show research → running transition
        const poll = setInterval(async () => {
          try {
            const sr = await fetch(`/api/simulations/${data.id}`)
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
        setError(data.error || 'Failed to start simulation')
        setRunning(false)
        setSimPhase('idle')
      }
    } catch {
      setError('Could not connect to FRACTURE engine. Make sure the server is running.')
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

      {/* Industry / Vertical */}
      <div style={{ marginBottom: '20px' }}>
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Setor / Vertical <span style={{ fontWeight: '400', textTransform: 'none', letterSpacing: 0 }}>(opcional)</span></label>
        <select style={{ ...Input }} value={industry} onChange={e => setIndustry(e.target.value)}>
          {INDUSTRIES.map(ind => <option key={ind.id} value={ind.id}>{ind.label}</option>)}
        </select>
        {industry && INDUSTRIES.find(i => i.id === industry)?.description && (
          <div style={{ marginTop: '8px', padding: '10px 14px', borderRadius: '8px', background: 'oklch(0.65 0.22 30 / 0.06)', border: '1px solid oklch(0.65 0.22 30 / 0.2)', fontSize: '12px', color: 'var(--color-text-muted)', lineHeight: '1.6' }}>
            {INDUSTRIES.find(i => i.id === industry)?.description}
          </div>
        )}
      </div>

      {/* Company URLs Section */}
      <div style={{ marginBottom: '20px', borderRadius: '10px', border: '1px solid var(--color-border)', overflow: 'hidden' }}>
        <button
          onClick={() => setShowUrlSection(!showUrlSection)}
          style={{ width: '100%', padding: '12px 16px', background: 'var(--color-surface-2)', border: 'none', color: 'var(--color-text)', fontSize: '13px', fontWeight: '600', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'space-between', textAlign: 'left' as const }}
        >
          <span>🔗 Company Context — Website & Social Media <span style={{ color: 'var(--color-text-muted)', fontWeight: '400' }}>(optional, improves simulation accuracy)</span></span>
          <span style={{ color: 'var(--color-text-muted)', fontSize: '11px' }}>{showUrlSection ? '▲ hide' : '▼ expand'}</span>
        </button>

        {showUrlSection && (
          <div style={{ padding: '16px', borderTop: '1px solid var(--color-border)' }}>
            <p style={{ margin: '0 0 14px', fontSize: '12px', color: 'var(--color-text-muted)', lineHeight: '1.5' }}>
              Add your company's website and social media profiles. FRACTURE will automatically extract public information to make the simulation more accurate and personalized.
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
                  + Add another URL
                </button>
              )}
              <button
                onClick={handleExtractContext}
                disabled={extracting || urlEntries.every(u => !u.value.trim())}
                style={{ padding: '8px 16px', borderRadius: '6px', border: '1px solid var(--color-accent)', background: 'oklch(0.65 0.22 30 / 0.1)', color: 'var(--color-accent)', fontSize: '12px', fontWeight: '600', cursor: extracting ? 'not-allowed' : 'pointer' }}
              >
                {extracting ? '⟳ Extracting...' : '⚡ Preview extracted context'}
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
        <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '8px', fontWeight: '600', letterSpacing: '0.5px', textTransform: 'uppercase' }}>Additional Context (optional)</label>
        <textarea style={{ ...Input, height: '80px', resize: 'vertical' }} value={context} onChange={e => setContext(e.target.value)} placeholder="Paste relevant news, competitor info, or market data to enrich the simulation..." />
      </div>

      {error && <div style={{ color: 'var(--color-danger)', fontSize: '13px', marginBottom: '16px', padding: '12px', borderRadius: '8px', background: 'oklch(0.60 0.22 25 / 0.1)', border: '1px solid oklch(0.60 0.22 25 / 0.3)' }}>{error}</div>}

      {/* Progress indicator */}
      {running && (
        <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '10px', background: 'var(--color-surface-2)', border: '1px solid var(--color-border)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '10px' }}>
            <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: simPhase === 'researching' ? 'oklch(0.75 0.18 55)' : 'var(--color-accent)', animation: 'pulse 1.2s infinite' }} />
            <span style={{ fontSize: '13px', fontWeight: '600', color: 'var(--color-text)' }}>
              {simPhase === 'researching' ? '🔍 DeepSearch — Researching market context...' : '◈ FRACTURE — Running simulation with 32 agents...'}
            </span>
          </div>
          <div style={{ display: 'flex', gap: '20px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
            <span style={{ color: simPhase !== 'idle' ? 'oklch(0.75 0.18 55)' : 'var(--color-text-muted)' }}>✓ DeepSearch {simPhase === 'researching' ? 'running...' : researchSources > 0 ? `— ${researchSources} sources found` : '— complete'}</span>
            <span style={{ color: simPhase === 'running' ? 'var(--color-accent)' : 'var(--color-text-muted)' }}>{simPhase === 'running' ? '⟳ Simulation running...' : '○ Simulation queued'}</span>
          </div>
          <style>{`@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.3} }`}</style>
        </div>
      )}

      <div style={{ display: 'flex', gap: '12px' }}>
        <button onClick={() => onNavigate('home')} disabled={running} style={{ padding: '10px 20px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '14px', cursor: running ? 'not-allowed' : 'pointer' }}>Cancel</button>
        <button onClick={handleRun} disabled={!question.trim() || running}
          style={{ padding: '10px 28px', borderRadius: '8px', border: 'none', background: running ? 'var(--color-border)' : 'var(--color-accent)', color: '#fff', fontSize: '14px', fontWeight: '600', cursor: running ? 'not-allowed' : 'pointer' }}>
          {running ? (simPhase === 'researching' ? '🔍 Researching...' : '◈ Simulating...') : '◈ Run Simulation'}
        </button>
      </div>
    </div>
  )
}
