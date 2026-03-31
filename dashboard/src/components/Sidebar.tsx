import { type Page } from '../App'

const nav = [
  { id: 'home',           label: 'Dashboard',       icon: '⬡' },
  { id: 'new-simulation', label: 'Nova Simulação',   icon: '◈' },
  { id: 'simulations',    label: 'Histórico',        icon: '◎' },
  { id: 'schedules',      label: 'Agendamentos',     icon: '⏱' },
  { id: 'accuracy',       label: 'Acurácia',         icon: '◉' },
  { id: 'archetypes',     label: 'Arquétipos',       icon: '◇' },
  { id: 'api-keys',       label: 'API Pública',      icon: '⌘' },
  { id: 'settings',       label: 'Configurações',    icon: '⚙' },
] as const

export default function Sidebar({ currentPage, onNavigate }: {
  currentPage: Page
  onNavigate: (p: Page) => void
}) {
  return (
    <aside style={{
      width: '220px', minHeight: '100%', flexShrink: 0,
      background: 'var(--color-surface)',
      borderRight: '1px solid var(--color-border)',
      display: 'flex', flexDirection: 'column', padding: '0'
    }}>
      {/* Logo */}
      <div style={{ padding: '24px 20px 20px', borderBottom: '1px solid var(--color-border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <div className="fracture-glow" style={{
            width: '32px', height: '32px', borderRadius: '8px',
            background: 'var(--color-accent)', display: 'flex',
            alignItems: 'center', justifyContent: 'center',
            fontSize: '16px', fontWeight: '700', color: '#fff'
          }}>F</div>
          <div>
            <div style={{ fontWeight: '700', fontSize: '15px', letterSpacing: '-0.3px', color: 'var(--color-text)' }}>FRACTURE</div>
            <div style={{ fontSize: '10px', color: 'var(--color-text-muted)', letterSpacing: '0.5px' }}>MOTOR DE SIMULAÇÃO</div>
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav style={{ flex: 1, padding: '12px 8px' }}>
        {nav.map(item => {
          const active = currentPage === item.id
          return (
            <button key={item.id} onClick={() => onNavigate(item.id as Page)}
              style={{
                width: '100%', display: 'flex', alignItems: 'center', gap: '10px',
                padding: '9px 12px', borderRadius: '8px', border: 'none', cursor: 'pointer',
                background: active ? 'oklch(0.65 0.22 30 / 0.15)' : 'transparent',
                color: active ? 'var(--color-accent)' : 'var(--color-text-muted)',
                fontSize: '13px', fontWeight: active ? '600' : '400',
                transition: 'all 0.15s', textAlign: 'left',
                borderLeft: active ? '2px solid var(--color-accent)' : '2px solid transparent',
              }}>
              <span style={{ fontSize: '16px', lineHeight: 1 }}>{item.icon}</span>
              {item.label}
            </button>
          )
        })}
      </nav>

      {/* Version */}
      <div style={{ padding: '16px 20px', borderTop: '1px solid var(--color-border)' }}>
        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>v1.0.0-alpha</div>
      </div>
    </aside>
  )
}
