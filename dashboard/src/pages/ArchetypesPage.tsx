import { useState, useEffect } from 'react'

type Archetype = {
  id: string
  name: string
  agent_type: string
  description: string
  memory_weight: number
  is_active: boolean
}

export default function ArchetypesPage() {
  const [archetypes, setArchetypes] = useState<Archetype[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/v1/archetypes')
      .then(r => r.json())
      .then((d: Archetype[]) => { setArchetypes(d ?? []); setLoading(false) })
      .catch(() => setLoading(false))
  }, [])

  const conformists = archetypes.filter(a => a.agent_type === 'conformist')
  const disruptors  = archetypes.filter(a => a.agent_type === 'disruptor')

  const Card = ({ a }: { a: Archetype }) => (
    <div style={{
      background: 'var(--color-surface)', borderRadius: '10px',
      border: `1px solid ${a.agent_type === 'disruptor' ? 'oklch(0.55 0.18 300 / 0.4)' : 'var(--color-border)'}`,
      padding: '16px'
    }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
        <div style={{ fontWeight: '600', fontSize: '13px', color: 'var(--color-text)' }}>{a.name}</div>
        <div style={{
          fontSize: '11px', padding: '2px 8px', borderRadius: '4px',
          background: a.agent_type === 'disruptor' ? 'oklch(0.55 0.18 300 / 0.15)' : 'oklch(0.65 0.18 145 / 0.15)',
          color: a.agent_type === 'disruptor' ? 'oklch(0.75 0.18 300)' : 'var(--color-success)'
        }}>
          {a.agent_type === 'disruptor' ? 'Disruptor' : 'Conformista'}
        </div>
      </div>
      <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '10px', lineHeight: '1.5' }}>
        {a.description}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>Poder</div>
        <div style={{ flex: 1, height: '4px', borderRadius: '2px', background: 'var(--color-border)' }}>
          <div style={{
            height: '100%', borderRadius: '2px',
            width: `${a.memory_weight * 100}%`,
            background: a.agent_type === 'disruptor' ? 'oklch(0.55 0.18 300)' : 'var(--color-accent)'
          }} />
        </div>
        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{Math.round(a.memory_weight * 100)}%</div>
      </div>
    </div>
  )

  if (loading) {
    return (
      <div style={{ padding: '32px', color: 'var(--color-text-muted)', fontSize: '13px' }}>
        Carregando arquétipos...
      </div>
    )
  }

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <div style={{ marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Arquétipos</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>
          {archetypes.length} agentes que simulam o comportamento de mercado.{' '}
          {conformists.length} Conformistas defendem o status quo.{' '}
          {disruptors.length} Disruptores tentam quebrá-lo.
        </p>
      </div>

      <div style={{ marginBottom: '28px' }}>
        <div style={{ fontSize: '12px', fontWeight: '600', color: 'var(--color-text-muted)', letterSpacing: '0.5px', textTransform: 'uppercase', marginBottom: '12px' }}>
          Conformistas ({conformists.length})
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px' }}>
          {conformists.map(a => <Card key={a.id} a={a} />)}
        </div>
      </div>

      <div>
        <div style={{ fontSize: '12px', fontWeight: '600', color: 'var(--color-text-muted)', letterSpacing: '0.5px', textTransform: 'uppercase', marginBottom: '12px' }}>
          Disruptores ({disruptors.length})
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px' }}>
          {disruptors.map(a => <Card key={a.id} a={a} />)}
        </div>
      </div>
    </div>
  )
}
