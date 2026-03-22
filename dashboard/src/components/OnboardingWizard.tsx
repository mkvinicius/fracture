import { useState } from 'react'

const PROVIDERS = [
  { id: 'openai',    name: 'OpenAI',    desc: 'GPT-4o — recommended for synthesis',    placeholder: 'sk-...' },
  { id: 'anthropic', name: 'Anthropic', desc: 'Claude Sonnet — best for disruption',   placeholder: 'sk-ant-...' },
  { id: 'google',    name: 'Google',    desc: 'Gemini — optional third model',         placeholder: 'AIza...' },
]

export default function OnboardingWizard({ onComplete }: { onComplete: () => void }) {
  const [step, setStep] = useState(0)
  const [company, setCompany] = useState({ name: '', sector: '', description: '' })
  const [keys, setKeys] = useState<Record<string, string>>({})
  const [telemetryEnabled, setTelemetryEnabled] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // steps: 0=Welcome, 1=Company, 2=AI Keys, 3=Telemetry, 4=Ready
  const steps = ['Welcome', 'Company Profile', 'AI Keys', 'Privacy', 'Ready']

  async function handleFinish() {
    setSaving(true)
    setError('')
    try {
      // Save company
      await fetch('/api/company', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(company) })
      // Save keys
      for (const [provider, key] of Object.entries(keys)) {
        if (key.trim()) {
          await fetch('/api/keys/validate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ provider, key: key.trim() }) })
        }
      }
      // Save telemetry preference
      await fetch('/api/telemetry', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ enabled: telemetryEnabled }) })
      // Mark onboarding complete
      await fetch('/api/onboarding/complete', { method: 'POST' })
      onComplete()
    } catch {
      setError('Failed to save. Make sure FRACTURE is running.')
      setSaving(false)
    }
  }

  const S: React.CSSProperties = {
    minHeight: '100%', background: 'var(--color-background)',
    display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '40px 20px'
  }
  const Card: React.CSSProperties = {
    width: '100%', maxWidth: '520px',
    background: 'var(--color-surface)', borderRadius: '16px',
    border: '1px solid var(--color-border)', padding: '40px',
  }
  const Input: React.CSSProperties = {
    width: '100%', padding: '10px 14px', borderRadius: '8px',
    border: '1px solid var(--color-border)', background: 'var(--color-surface-2)',
    color: 'var(--color-text)', fontSize: '14px', outline: 'none',
    boxSizing: 'border-box' as const,
  }
  const Btn: React.CSSProperties = {
    padding: '10px 24px', borderRadius: '8px', border: 'none', cursor: 'pointer',
    background: 'var(--color-accent)', color: '#fff', fontSize: '14px', fontWeight: '600',
  }

  return (
    <div style={S}>
      <div style={Card}>
        {/* Progress */}
        <div style={{ display: 'flex', gap: '8px', marginBottom: '32px' }}>
          {steps.map((s, i) => (
            <div key={s} style={{ flex: 1, height: '3px', borderRadius: '2px',
              background: i <= step ? 'var(--color-accent)' : 'var(--color-border)' }} />
          ))}
        </div>

        {step === 0 && (
          <div>
            <div style={{ fontSize: '28px', fontWeight: '700', marginBottom: '12px', color: 'var(--color-text)' }}>
              Welcome to <span style={{ color: 'var(--color-accent)' }}>FRACTURE</span>
            </div>
            <p style={{ color: 'var(--color-text-muted)', lineHeight: '1.7', marginBottom: '32px' }}>
              FRACTURE simulates how market rules break — and how you can be the one to break them first.
              Let's set up your workspace in 4 quick steps.
            </p>
            <button style={Btn} onClick={() => setStep(1)}>Get Started →</button>
          </div>
        )}

        {step === 1 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px', color: 'var(--color-text)' }}>Company Profile</div>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '24px', fontSize: '13px' }}>
              This context is injected into every simulation so agents understand your market.
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
              <div>
                <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Company Name</label>
                <input style={Input} value={company.name} onChange={e => setCompany(c => ({...c, name: e.target.value}))} placeholder="Acme Corp" />
              </div>
              <div>
                <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Sector / Industry</label>
                <input style={Input} value={company.sector} onChange={e => setCompany(c => ({...c, sector: e.target.value}))} placeholder="e.g. SaaS, Retail, Healthcare" />
              </div>
              <div>
                <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Brief Description</label>
                <textarea style={{...Input, height: '80px', resize: 'vertical'}} value={company.description} onChange={e => setCompany(c => ({...c, description: e.target.value}))} placeholder="What does your company do? Who are your main competitors?" />
              </div>
            </div>
            <div style={{ display: 'flex', gap: '12px', marginTop: '24px' }}>
              <button style={{...Btn, background: 'var(--color-surface-2)', color: 'var(--color-text)'}} onClick={() => setStep(0)}>← Back</button>
              <button style={Btn} onClick={() => setStep(2)} disabled={!company.name}>Next →</button>
            </div>
          </div>
        )}

        {step === 2 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px', color: 'var(--color-text)' }}>AI Keys</div>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '24px', fontSize: '13px' }}>
              Keys are stored locally on your machine. They never leave your computer.
              Add at least one to run simulations.
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              {PROVIDERS.map(p => (
                <div key={p.id} style={{ padding: '16px', borderRadius: '10px', border: '1px solid var(--color-border)', background: 'var(--color-surface-2)' }}>
                  <div style={{ fontWeight: '600', fontSize: '13px', color: 'var(--color-text)', marginBottom: '2px' }}>{p.name}</div>
                  <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '10px' }}>{p.desc}</div>
                  <input style={Input} type="password" value={keys[p.id] || ''} onChange={e => setKeys(k => ({...k, [p.id]: e.target.value}))} placeholder={p.placeholder} />
                </div>
              ))}
            </div>
            {error && <div style={{ color: 'var(--color-danger)', fontSize: '13px', marginTop: '12px' }}>{error}</div>}
            <div style={{ display: 'flex', gap: '12px', marginTop: '24px' }}>
              <button style={{...Btn, background: 'var(--color-surface-2)', color: 'var(--color-text)'}} onClick={() => setStep(1)}>← Back</button>
              <button style={Btn} onClick={() => setStep(3)} disabled={Object.values(keys).every(k => !k.trim())}>Next →</button>
            </div>
          </div>
        )}

        {step === 3 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px', color: 'var(--color-text)' }}>
              Privacy &amp; Telemetry
            </div>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '20px', fontSize: '13px', lineHeight: '1.7' }}>
              FRACTURE collects <strong style={{ color: 'var(--color-text)' }}>anonymous usage data</strong> to understand
              how the tool is being used and improve future versions. No personal data, no simulation content,
              no API keys are ever collected.
            </p>

            {/* What is collected */}
            <div style={{ padding: '14px 16px', borderRadius: '10px', background: 'var(--color-surface-2)', border: '1px solid var(--color-border)', marginBottom: '20px' }}>
              <div style={{ fontSize: '12px', fontWeight: '600', color: 'var(--color-text-muted)', marginBottom: '10px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                What is collected
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px' }}>
                {[
                  ['🔑', 'Anonymous install ID'],
                  ['🖥️', 'OS &amp; architecture'],
                  ['🌍', 'Country (from IP)'],
                  ['📦', 'FRACTURE version'],
                ].map(([icon, label]) => (
                  <div key={label} style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
                    <span>{icon}</span>
                    <span dangerouslySetInnerHTML={{ __html: label }} />
                  </div>
                ))}
              </div>
              <div style={{ marginTop: '10px', fontSize: '11px', color: 'var(--color-text-muted)', opacity: 0.7 }}>
                Your IP is masked (last octet removed). No simulation data is ever sent.
              </div>
            </div>

            {/* Toggle */}
            <div
              style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: '14px 16px', borderRadius: '10px',
                border: `1px solid ${telemetryEnabled ? 'var(--color-accent)' : 'var(--color-border)'}`,
                background: telemetryEnabled ? 'oklch(0.65 0.22 30 / 0.08)' : 'var(--color-surface-2)',
                cursor: 'pointer', transition: 'all 0.2s',
              }}
              onClick={() => setTelemetryEnabled(v => !v)}
            >
              <div>
                <div style={{ fontSize: '14px', fontWeight: '600', color: 'var(--color-text)' }}>
                  {telemetryEnabled ? 'Telemetry enabled' : 'Telemetry disabled'}
                </div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '2px' }}>
                  {telemetryEnabled
                    ? 'Helping improve FRACTURE — thank you!'
                    : 'You can re-enable this anytime in Settings.'}
                </div>
              </div>
              {/* Toggle switch */}
              <div style={{
                width: '44px', height: '24px', borderRadius: '12px',
                background: telemetryEnabled ? 'var(--color-accent)' : 'var(--color-border)',
                position: 'relative', transition: 'background 0.2s', flexShrink: 0,
              }}>
                <div style={{
                  position: 'absolute', top: '3px',
                  left: telemetryEnabled ? '23px' : '3px',
                  width: '18px', height: '18px', borderRadius: '50%',
                  background: '#fff', transition: 'left 0.2s',
                }} />
              </div>
            </div>

            <div style={{ display: 'flex', gap: '12px', marginTop: '24px' }}>
              <button style={{...Btn, background: 'var(--color-surface-2)', color: 'var(--color-text)'}} onClick={() => setStep(2)}>← Back</button>
              <button style={Btn} onClick={() => setStep(4)}>Next →</button>
            </div>
          </div>
        )}

        {step === 4 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '12px', color: 'var(--color-text)' }}>
              You're ready 🎯
            </div>
            <p style={{ color: 'var(--color-text-muted)', lineHeight: '1.7', marginBottom: '32px' }}>
              FRACTURE is configured. You can now run your first simulation — ask any strategic question and watch the market simulate itself.
            </p>
            <div style={{ padding: '16px', borderRadius: '10px', background: 'oklch(0.65 0.22 30 / 0.1)', border: '1px solid oklch(0.65 0.22 30 / 0.3)', marginBottom: '28px' }}>
              <div style={{ fontSize: '13px', color: 'var(--color-accent)', fontWeight: '600', marginBottom: '4px' }}>Tip: Start with a question like</div>
              <div style={{ fontSize: '13px', color: 'var(--color-text)', fontStyle: 'italic' }}>
                "If a competitor offered our core product for free, how would the market react in 12 months?"
              </div>
            </div>
            <button style={{...Btn, opacity: saving ? 0.6 : 1}} onClick={handleFinish} disabled={saving}>
              {saving ? 'Saving...' : 'Launch FRACTURE →'}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
