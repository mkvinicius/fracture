import { useState, useEffect } from 'react'

const PROVIDERS = [
  { id: 'openai',    name: 'OpenAI',    desc: 'GPT-4o · Conformist agents + synthesis', placeholder: 'sk-...' },
  { id: 'anthropic', name: 'Anthropic', desc: 'Claude Sonnet · Disruptor agents',       placeholder: 'sk-ant-...' },
  { id: 'google',    name: 'Google',    desc: 'Gemini · Optional third model',          placeholder: 'AIza...' },
]

export default function SettingsPage() {
  const [keys, setKeys] = useState<Record<string, string>>({})
  const [config, setConfig] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [telemetryEnabled, setTelemetryEnabled] = useState<boolean | null>(null)
  const [telemetrySaving, setTelemetrySaving] = useState(false)

  useEffect(() => {
    fetch('/api/v1/config').then(r => r.json()).then(setConfig).catch(() => {})
    fetch('/api/v1/telemetry').then(r => r.json()).then((d: { enabled: boolean }) => setTelemetryEnabled(d.enabled)).catch(() => setTelemetryEnabled(true))
  }, [])

  async function saveKey(provider: string) {
    const key = keys[provider]
    if (!key?.trim()) return
    setSaving(true)
    await fetch('/api/v1/keys/validate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ provider, key: key.trim() }) })
    setSaving(false)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  async function toggleTelemetry() {
    if (telemetryEnabled === null) return
    const next = !telemetryEnabled
    setTelemetryEnabled(next)
    setTelemetrySaving(true)
    try {
      await fetch('/api/v1/telemetry', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ enabled: next }) })
    } catch { /* ignore */ }
    setTelemetrySaving(false)
  }

  const Input: React.CSSProperties = {
    flex: 1, padding: '9px 12px', borderRadius: '8px',
    border: '1px solid var(--color-border)', background: 'var(--color-surface-2)',
    color: 'var(--color-text)', fontSize: '13px', outline: 'none',
  }

  return (
    <div style={{ padding: '32px', maxWidth: '640px' }}>
      <div style={{ marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Settings</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>Configure AI providers and simulation defaults</p>
      </div>

      {/* AI Keys */}
      <div style={{ background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', padding: '24px', marginBottom: '20px' }}>
        <div style={{ fontWeight: '600', fontSize: '14px', color: 'var(--color-text)', marginBottom: '4px' }}>AI Provider Keys</div>
        <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '20px' }}>Keys are stored locally in your SQLite database. They never leave your machine.</div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {PROVIDERS.map(p => (
            <div key={p.id}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '6px' }}>
                <div>
                  <span style={{ fontWeight: '600', fontSize: '13px', color: 'var(--color-text)' }}>{p.name}</span>
                  <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginLeft: '8px' }}>{p.desc}</span>
                </div>
                {config[p.id + '_key_set'] === 'true' && <span style={{ fontSize: '11px', color: 'var(--color-success)' }}>✓ Connected</span>}
              </div>
              <div style={{ display: 'flex', gap: '8px' }}>
                <input style={Input} type="password" value={keys[p.id] || ''} onChange={e => setKeys(k => ({...k, [p.id]: e.target.value}))} placeholder={config[p.id + '_key_set'] === 'true' ? '••••••••••••••••' : p.placeholder} />
                <button onClick={() => saveKey(p.id)} disabled={!keys[p.id]?.trim() || saving}
                  style={{ padding: '9px 16px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', cursor: 'pointer', opacity: !keys[p.id]?.trim() ? 0.4 : 1 }}>
                  Save
                </button>
              </div>
            </div>
          ))}
        </div>
        {saved && <div style={{ marginTop: '12px', fontSize: '13px', color: 'var(--color-success)' }}>✓ Key saved successfully</div>}
      </div>

      {/* Telemetry */}
      <div style={{ background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', padding: '24px', marginBottom: '20px' }}>
        <div style={{ fontWeight: '600', fontSize: '14px', color: 'var(--color-text)', marginBottom: '4px' }}>Privacy &amp; Telemetry</div>
        <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '20px', lineHeight: '1.7' }}>
          FRACTURE collects <strong style={{ color: 'var(--color-text)' }}>anonymous usage data</strong> to improve the tool.
          No simulation content, no API keys, and no personal data are ever collected.
          Your IP is masked (last octet removed).
        </div>

        {/* What is collected */}
        <div style={{ padding: '12px 14px', borderRadius: '8px', background: 'var(--color-surface-2)', border: '1px solid var(--color-border)', marginBottom: '16px' }}>
          <div style={{ fontSize: '11px', fontWeight: '600', color: 'var(--color-text-muted)', marginBottom: '8px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Data collected
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '4px' }}>
            {[
              'Anonymous install ID (UUID)',
              'OS & architecture',
              'Country (from IP, masked)',
              'FRACTURE version',
            ].map(item => (
              <div key={item} style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'flex', alignItems: 'center', gap: '4px' }}>
                <span style={{ color: 'var(--color-success)', fontSize: '10px' }}>●</span> {item}
              </div>
            ))}
          </div>
        </div>

        {/* Toggle row */}
        <div
          style={{
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            padding: '12px 14px', borderRadius: '8px',
            border: `1px solid ${telemetryEnabled ? 'var(--color-accent)' : 'var(--color-border)'}`,
            background: telemetryEnabled ? 'oklch(0.65 0.22 30 / 0.06)' : 'var(--color-surface-2)',
            cursor: telemetrySaving || telemetryEnabled === null ? 'not-allowed' : 'pointer',
            opacity: telemetryEnabled === null ? 0.5 : 1,
            transition: 'all 0.2s',
          }}
          onClick={telemetrySaving || telemetryEnabled === null ? undefined : toggleTelemetry}
        >
          <div>
            <div style={{ fontSize: '13px', fontWeight: '600', color: 'var(--color-text)' }}>
              {telemetryEnabled ? 'Telemetry enabled' : 'Telemetry disabled'}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '2px' }}>
              {telemetryEnabled
                ? 'Helping improve FRACTURE — thank you!'
                : 'You can re-enable this anytime.'}
            </div>
          </div>
          {/* Toggle switch */}
          <div style={{
            width: '40px', height: '22px', borderRadius: '11px',
            background: telemetryEnabled ? 'var(--color-accent)' : 'var(--color-border)',
            position: 'relative', transition: 'background 0.2s', flexShrink: 0,
          }}>
            <div style={{
              position: 'absolute', top: '3px',
              left: telemetryEnabled ? '21px' : '3px',
              width: '16px', height: '16px', borderRadius: '50%',
              background: '#fff', transition: 'left 0.2s',
              boxShadow: '0 1px 3px rgba(0,0,0,0.2)',
            }} />
          </div>
        </div>
      </div>

      {/* About */}
      <div style={{ background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', padding: '24px' }}>
        <div style={{ fontWeight: '600', fontSize: '14px', color: 'var(--color-text)', marginBottom: '4px' }}>About FRACTURE</div>
        <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', lineHeight: '1.7' }}>
          FRACTURE v1.1.0 · Built on Go + React<br/>
          Data stored locally in SQLite · No cloud dependency
        </div>
      </div>
    </div>
  )
}
