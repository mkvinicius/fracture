const ARCHETYPES = [
  { name: 'The Pragmatist',      type: 'Conformist', power: 0.7, desc: 'Evaluates changes through cost-benefit lens. Resists unless ROI is clear.' },
  { name: 'The Loyalist',        type: 'Conformist', power: 0.6, desc: 'Deeply attached to current rules. Defends the status quo actively.' },
  { name: 'The Analyst',         type: 'Conformist', power: 0.5, desc: 'Data-driven. Accepts change only with strong evidence.' },
  { name: 'The Opportunist',     type: 'Conformist', power: 0.8, desc: 'Adapts quickly to whoever holds power. Follows winning side.' },
  { name: 'The Traditionalist',  type: 'Conformist', power: 0.6, desc: 'Values proven methods. Skeptical of unproven innovations.' },
  { name: 'The Regulator',       type: 'Conformist', power: 0.9, desc: 'Enforces compliance. High institutional power.' },
  { name: 'The Consumer',        type: 'Conformist', power: 0.4, desc: 'Passive but numerous. Collective behavior shapes outcomes.' },
  { name: 'The Investor',        type: 'Conformist', power: 0.85,'desc': 'Optimizes for returns. Will support disruption if profitable.' },
  { name: 'The Visionary',       type: 'Disruptor',  power: 0.7, desc: 'Sees 10 years ahead. Proposes radical rule rewrites.' },
  { name: 'The Rebel',           type: 'Disruptor',  power: 0.5, desc: 'Challenges authority on principle. Unpredictable but catalytic.' },
  { name: 'The Tech Accelerator',type: 'Disruptor',  power: 0.8, desc: 'Uses technology to make existing rules obsolete overnight.' },
  { name: 'The Arbitrageur',     type: 'Disruptor',  power: 0.65,'desc': 'Exploits gaps between rules. Profits from inconsistencies.' },
]

export default function ArchetypesPage() {
  const conformists = ARCHETYPES.filter(a => a.type === 'Conformist')
  const disruptors = ARCHETYPES.filter(a => a.type === 'Disruptor')

  const Card = ({ a }: { a: typeof ARCHETYPES[0] }) => (
    <div style={{ background: 'var(--color-surface)', borderRadius: '10px', border: `1px solid ${a.type === 'Disruptor' ? 'oklch(0.55 0.18 300 / 0.4)' : 'var(--color-border)'}`, padding: '16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
        <div style={{ fontWeight: '600', fontSize: '13px', color: 'var(--color-text)' }}>{a.name}</div>
        <div style={{ fontSize: '11px', padding: '2px 8px', borderRadius: '4px', background: a.type === 'Disruptor' ? 'oklch(0.55 0.18 300 / 0.15)' : 'oklch(0.65 0.18 145 / 0.15)', color: a.type === 'Disruptor' ? 'oklch(0.75 0.18 300)' : 'var(--color-success)' }}>{a.type}</div>
      </div>
      <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '10px', lineHeight: '1.5' }}>{a.desc}</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>Power</div>
        <div style={{ flex: 1, height: '4px', borderRadius: '2px', background: 'var(--color-border)' }}>
          <div style={{ height: '100%', borderRadius: '2px', width: `${a.power * 100}%`, background: a.type === 'Disruptor' ? 'oklch(0.55 0.18 300)' : 'var(--color-accent)' }} />
        </div>
        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{Math.round(a.power * 100)}%</div>
      </div>
    </div>
  )

  return (
    <div style={{ padding: '32px', maxWidth: '960px' }}>
      <div style={{ marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Archetypes</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>12 agents that simulate market behavior. 8 Conformists defend the status quo. 4 Disruptors try to break it.</p>
      </div>

      <div style={{ marginBottom: '28px' }}>
        <div style={{ fontSize: '12px', fontWeight: '600', color: 'var(--color-text-muted)', letterSpacing: '0.5px', textTransform: 'uppercase', marginBottom: '12px' }}>Conformists (8)</div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px' }}>
          {conformists.map(a => <Card key={a.name} a={a} />)}
        </div>
      </div>

      <div>
        <div style={{ fontSize: '12px', fontWeight: '600', color: 'var(--color-text-muted)', letterSpacing: '0.5px', textTransform: 'uppercase', marginBottom: '12px' }}>Disruptors (4)</div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px' }}>
          {disruptors.map(a => <Card key={a.name} a={a} />)}
        </div>
      </div>
    </div>
  )
}
