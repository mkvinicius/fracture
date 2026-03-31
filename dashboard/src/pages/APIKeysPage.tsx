import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface APIKey { id: string; name: string; key_prefix: string; sims_used: number; sims_limit: number; enabled: boolean; created_at: number; last_used_at?: number; raw_key?: string }

export default function APIKeysPage({ onNavigate: _onNavigate }: { onNavigate: (p: Page) => void }) {
  const [keys, setKeys] = useState<APIKey[]>([])
  const [loading, setLoading] = useState(true)
  const [name, setName] = useState('')
  const [simsLimit, setSimsLimit] = useState(0)
  const [creating, setCreating] = useState(false)
  const [newKey, setNewKey] = useState<APIKey | null>(null)

  const load = () => fetch('/api/v1/api-keys').then(r => r.json()).then(d => { setKeys(d ?? []); setLoading(false) }).catch(() => setLoading(false))

  useEffect(() => { load() }, [])

  async function create() {
    if (!name.trim()) return
    setCreating(true)
    const res = await fetch('/api/v1/api-keys', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, sims_limit: simsLimit })
    })
    const data = await res.json()
    setNewKey(data)
    setName(''); setCreating(false); load()
  }

  async function remove(id: string) {
    if (!window.confirm('Excluir esta chave? Integrações que a usam vão parar de funcionar.')) return
    await fetch(`/api/v1/api-keys/${id}`, { method: 'DELETE' })
    load()
  }

  const Input: React.CSSProperties = { padding: '8px 12px', borderRadius: '8px', border: '1px solid var(--color-border)', background: 'var(--color-surface-2)', color: 'var(--color-text)', fontSize: '13px', outline: 'none' }

  return (
    <div style={{ padding: '32px', maxWidth: '860px' }}>
      <div style={{ marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>API Pública</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>
          Integre o motor FRACTURE em suas ferramentas. Use a chave API para chamar <code style={{ background: 'var(--color-surface-2)', padding: '1px 6px', borderRadius: '4px' }}>POST /api/v1/simulations</code> com <code style={{ background: 'var(--color-surface-2)', padding: '1px 6px', borderRadius: '4px' }}>Authorization: Bearer frc_...</code>
        </p>
      </div>

      {/* Create key form */}
      <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '20px', marginBottom: '24px' }}>
        <div style={{ fontSize: '14px', fontWeight: '600', marginBottom: '14px' }}>Gerar Nova Chave API</div>
        <div style={{ display: 'flex', gap: '10px', alignItems: 'flex-end' }}>
          <div style={{ flex: 1 }}>
            <label style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Nome da integração</label>
            <input style={{ ...Input, width: '100%', boxSizing: 'border-box' as const }} value={name} onChange={e => setName(e.target.value)} placeholder='ex: "Painel de BI", "Slack bot"' />
          </div>
          <div>
            <label style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Limite de simulações (0 = ilimitado)</label>
            <input type="number" min={0} style={{ ...Input, width: '100px' }} value={simsLimit} onChange={e => setSimsLimit(Number(e.target.value))} />
          </div>
          <button onClick={create} disabled={creating || !name.trim()}
            style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer', whiteSpace: 'nowrap' }}>
            {creating ? 'Gerando...' : 'Gerar Chave'}
          </button>
        </div>
      </div>

      {/* New key reveal */}
      {newKey?.raw_key && (
        <div style={{ background: 'oklch(0.18 0.05 150)', border: '1px solid var(--color-success)', borderRadius: '10px', padding: '16px 20px', marginBottom: '24px' }}>
          <div style={{ fontSize: '13px', fontWeight: '700', color: 'var(--color-success)', marginBottom: '8px' }}>✓ Chave gerada — copie agora, não será exibida novamente</div>
          <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
            <code style={{ flex: 1, background: 'var(--color-surface)', padding: '8px 12px', borderRadius: '6px', fontSize: '12px', fontFamily: 'monospace', color: 'var(--color-text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {newKey.raw_key}
            </code>
            <button onClick={() => { navigator.clipboard?.writeText(newKey.raw_key!); }}
              style={{ padding: '7px 14px', borderRadius: '6px', border: '1px solid var(--color-success)', background: 'transparent', color: 'var(--color-success)', fontSize: '12px', cursor: 'pointer', whiteSpace: 'nowrap' }}>
              Copiar
            </button>
            <button onClick={() => setNewKey(null)}
              style={{ padding: '7px 10px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px', cursor: 'pointer' }}>
              ✕
            </button>
          </div>
        </div>
      )}

      {/* Keys list */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-muted)' }}>Carregando...</div>
      ) : keys.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '40px', background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', color: 'var(--color-text-muted)', fontSize: '13px' }}>
          Nenhuma chave API gerada ainda
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {keys.map(k => (
            <div key={k.id} style={{ background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '14px 20px', display: 'flex', alignItems: 'center', gap: '16px' }}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: '14px', fontWeight: '500', color: 'var(--color-text)' }}>{k.name}</div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '2px' }}>
                  <code style={{ background: 'var(--color-surface-2)', padding: '1px 6px', borderRadius: '4px' }}>{k.key_prefix}</code>
                  {' · '}{k.sims_used} simulações usadas{k.sims_limit > 0 ? ` / ${k.sims_limit}` : ' (ilimitado)'}
                  {k.last_used_at && ` · Último uso: ${new Date(k.last_used_at * 1000).toLocaleDateString()}`}
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexShrink: 0 }}>
                <span style={{ fontSize: '11px', padding: '3px 8px', borderRadius: '6px', background: k.enabled ? 'oklch(0.65 0.22 30 / 0.15)' : 'oklch(0.4 0 0 / 0.1)', color: k.enabled ? 'var(--color-accent)' : 'var(--color-text-muted)', fontWeight: '600' }}>
                  {k.enabled ? 'Ativa' : 'Inativa'}
                </span>
                {k.sims_limit > 0 && (
                  <div style={{ width: '80px', height: '6px', borderRadius: '3px', background: 'var(--color-border)', overflow: 'hidden' }}>
                    <div style={{ width: `${Math.min(100, (k.sims_used / k.sims_limit) * 100)}%`, height: '100%', background: k.sims_used >= k.sims_limit ? 'var(--color-danger)' : 'var(--color-accent)', transition: 'width 0.3s' }} />
                  </div>
                )}
                <button onClick={() => remove(k.id)}
                  style={{ padding: '5px 10px', borderRadius: '6px', border: '1px solid var(--color-danger)', background: 'transparent', color: 'var(--color-danger)', fontSize: '12px', cursor: 'pointer' }}>✕</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* API docs snippet */}
      <div style={{ marginTop: '32px', background: 'var(--color-surface)', borderRadius: '10px', border: '1px solid var(--color-border)', padding: '20px' }}>
        <div style={{ fontSize: '13px', fontWeight: '600', marginBottom: '12px' }}>Exemplo de uso</div>
        <pre style={{ margin: 0, fontSize: '12px', color: 'var(--color-text-muted)', background: 'var(--color-surface-2)', padding: '14px', borderRadius: '8px', overflow: 'auto', lineHeight: '1.6' }}>{`curl -X POST http://localhost:4000/api/v1/simulations \\
  -H "Authorization: Bearer frc_sua_chave_aqui" \\
  -H "Content-Type: application/json" \\
  -d '{
    "question": "Como o mercado reagirá à entrada de um novo player?",
    "department": "Strategy",
    "rounds": 20,
    "company_size": "media",
    "company_sector": "foodtech"
  }'`}</pre>
      </div>
    </div>
  )
}
