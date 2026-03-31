import { useEffect, useState } from 'react'
import { type Page } from '../App'

interface AccuracyStats { total: number; confirmed: number; refuted: number; partial: number; pending: number; score: number }

export default function AccuracyPage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const [stats, setStats] = useState<AccuracyStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/v1/accuracy').then(r => r.json()).then(d => { setStats(d); setLoading(false) }).catch(() => setLoading(false))
  }, [])

  const pct = (n: number, total: number) => total > 0 ? Math.round((n / total) * 100) : 0
  const scoreColor = (s: number) => s >= 0.7 ? 'var(--color-success)' : s >= 0.5 ? 'oklch(0.75 0.18 55)' : 'var(--color-danger)'

  return (
    <div style={{ padding: '32px', maxWidth: '860px' }}>
      <div style={{ marginBottom: '28px' }}>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: '700', color: 'var(--color-text)' }}>Acurácia das Previsões</h1>
        <p style={{ margin: '6px 0 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>
          Rastreie quais previsões do FRACTURE se confirmaram. Quanto mais você validar, mais preciso o sistema fica.
        </p>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px', color: 'var(--color-text-muted)' }}>Carregando...</div>
      ) : !stats || stats.total === 0 ? (
        <div style={{ textAlign: 'center', padding: '60px', background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', color: 'var(--color-text-muted)' }}>
          <div style={{ fontSize: '28px', marginBottom: '12px' }}>◎</div>
          <div style={{ fontWeight: '600', marginBottom: '6px' }}>Nenhuma previsão validada ainda</div>
          <div style={{ fontSize: '13px', marginBottom: '20px' }}>Abra um relatório de simulação e marque se as previsões aconteceram ou não</div>
          <button onClick={() => onNavigate('simulations')} style={{ padding: '9px 20px', borderRadius: '8px', border: 'none', background: 'var(--color-accent)', color: '#fff', fontSize: '13px', fontWeight: '600', cursor: 'pointer' }}>
            Ver Simulações
          </button>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
          {/* Score card */}
          <div style={{ background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', padding: '28px', display: 'flex', gap: '40px', alignItems: 'center' }}>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '48px', fontWeight: '800', color: scoreColor(stats.score) }}>{Math.round(stats.score * 100)}%</div>
              <div style={{ fontSize: '13px', color: 'var(--color-text-muted)', marginTop: '4px' }}>Score de Acurácia</div>
            </div>
            <div style={{ flex: 1, display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: '20px' }}>
              {[
                { label: 'Confirmadas', value: stats.confirmed, color: 'var(--color-success)' },
                { label: 'Parciais', value: stats.partial, color: 'oklch(0.75 0.18 55)' },
                { label: 'Refutadas', value: stats.refuted, color: 'var(--color-danger)' },
                { label: 'Pendentes', value: stats.pending, color: 'var(--color-text-muted)' },
              ].map(item => (
                <div key={item.label} style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '28px', fontWeight: '700', color: item.color }}>{item.value}</div>
                  <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '2px' }}>{item.label}</div>
                  <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{pct(item.value, stats.total)}%</div>
                </div>
              ))}
            </div>
          </div>

          {/* Progress bar */}
          <div style={{ background: 'var(--color-surface)', borderRadius: '12px', border: '1px solid var(--color-border)', padding: '20px' }}>
            <div style={{ fontSize: '13px', fontWeight: '600', marginBottom: '12px' }}>Distribuição de {stats.total} previsões validadas</div>
            <div style={{ height: '12px', borderRadius: '6px', background: 'var(--color-border)', overflow: 'hidden', display: 'flex' }}>
              <div style={{ width: `${pct(stats.confirmed, stats.total)}%`, background: 'var(--color-success)', transition: 'width 0.5s' }} />
              <div style={{ width: `${pct(stats.partial, stats.total)}%`, background: 'oklch(0.75 0.18 55)', transition: 'width 0.5s' }} />
              <div style={{ width: `${pct(stats.refuted, stats.total)}%`, background: 'var(--color-danger)', transition: 'width 0.5s' }} />
            </div>
            <div style={{ display: 'flex', gap: '16px', marginTop: '10px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
              <span style={{ color: 'var(--color-success)' }}>■ Confirmadas</span>
              <span style={{ color: 'oklch(0.75 0.18 55)' }}>■ Parciais</span>
              <span style={{ color: 'var(--color-danger)' }}>■ Refutadas</span>
              <span>■ Pendentes</span>
            </div>
          </div>

          <div style={{ padding: '16px', borderRadius: '10px', background: 'oklch(0.65 0.22 30 / 0.05)', border: '1px solid oklch(0.65 0.22 30 / 0.2)', fontSize: '13px', color: 'var(--color-text-muted)' }}>
            Para validar previsões: abra um relatório de simulação → clique em "Validar Previsão" em cada cenário de ruptura
          </div>
        </div>
      )}
    </div>
  )
}
