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

  useEffect(() => {
    fetch('/api/config').then(r => r.json()).then(setConfig).catch(() => {})
  }, [])

  async function saveKey(provider: string) {
    const key = keys[provider]
    if (!key?.trim()) return
    setSaving(true)
    await fetch('/api/keys/validate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ provider, key: key.trim() }) })
    setSaving(false)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
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

      <div style={{ background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', padding: '24px' }}>
        <div style={{ fontWeight: '600', fontSize: '14px', color: 'var(--color-text)', marginBottom: '4px' }}>About FRACTURE</div>
        <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', lineHeight: '1.7' }}>
          FRACTURE v1.0.0-alpha · Built on Go + React<br/>
          Data stored locally in SQLite · No telemetry · No cloud dependency
        </div>
      </div>
    </div>
  )
}
