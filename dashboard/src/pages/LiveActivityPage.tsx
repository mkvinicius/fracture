import { useEffect, useRef, useState } from 'react'
import { type Page } from '../App'

// ── types ────────────────────────────────────────────────────────────────────

interface ActivityEvent {
  simulation_id: string
  round: number
  agent_id: string
  agent_name: string
  archetype: 'disruptor' | 'conformist' | 'system'
  action_type: 'react' | 'propose' | 'fracture' | 'council'
  snippet: string
  tokens_used: number
  total_tokens: number
  tension: number
  rule_id?: string
  accepted?: boolean
  ts: number
}

interface SimInfo { id: string; question: string; rounds: number; status: string }

type Filter = 'all' | 'disruptor' | 'fracture'

// ── helpers ──────────────────────────────────────────────────────────────────

function archetypeColor(a: string) {
  if (a === 'disruptor') return 'var(--color-danger)'
  if (a === 'system') return 'oklch(0.6 0.15 280)'
  return 'oklch(0.55 0.08 220)'
}

function archetypeLabel(a: string) {
  if (a === 'disruptor') return 'DISRUPTOR'
  if (a === 'system') return 'SISTEMA'
  return 'CONFORMISTA'
}

function actionLabel(t: string) {
  switch (t) {
    case 'propose': return 'PROPÕE'
    case 'fracture': return '⚡ FRATURA'
    case 'council': return 'CONSELHO'
    default: return 'AGE'
  }
}

function tensionColor(v: number) {
  if (v >= 0.7) return 'var(--color-danger)'
  if (v >= 0.5) return '#f97316'
  if (v >= 0.3) return 'var(--color-warning)'
  return 'var(--color-success)'
}

function fmtTime(ts: number) {
  const d = new Date(ts)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

// ── subcomponents ─────────────────────────────────────────────────────────────

function TokenMeter({ total, label }: { total: number; label: string }) {
  const max = 200_000
  const pct = Math.min(100, (total / max) * 100)
  const color = pct > 80 ? 'var(--color-danger)' : pct > 50 ? '#f97316' : 'var(--color-accent)'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', minWidth: '180px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: 'var(--color-text-muted)' }}>
        <span>{label}</span>
        <span style={{ fontWeight: '700', color: 'var(--color-text)', fontVariantNumeric: 'tabular-nums' }}>
          {total.toLocaleString()}
        </span>
      </div>
      <div style={{ height: '5px', borderRadius: '3px', background: 'var(--color-background)', overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${pct}%`, background: color, borderRadius: '3px', transition: 'width 0.4s ease' }} />
      </div>
    </div>
  )
}

function EventBubble({ ev, isNew }: { ev: ActivityEvent; isNew: boolean }) {
  const isFracture = ev.action_type === 'fracture'
  const accentCol = archetypeColor(ev.archetype)

  const borderColor = isFracture
    ? (ev.accepted ? 'var(--color-success)' : 'var(--color-danger)')
    : 'var(--color-border)'

  const bg = isFracture
    ? (ev.accepted ? 'oklch(0.13 0.03 145)' : 'oklch(0.13 0.03 25)')
    : 'var(--color-surface)'

  return (
    <div style={{
      background: bg,
      border: `1px solid ${borderColor}`,
      borderLeft: `3px solid ${isFracture ? borderColor : accentCol}`,
      borderRadius: '8px',
      padding: '10px 14px',
      display: 'flex',
      flexDirection: 'column',
      gap: '6px',
      opacity: isNew ? 0.85 : 1,
      transition: 'opacity 0.3s',
    }}>
      {/* header row */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
        <span style={{
          fontSize: '10px', fontWeight: '700', letterSpacing: '0.5px',
          padding: '2px 7px', borderRadius: '20px',
          background: `${accentCol}22`, color: accentCol,
        }}>{archetypeLabel(ev.archetype)}</span>

        <span style={{ fontSize: '13px', fontWeight: '600', color: 'var(--color-text)' }}>{ev.agent_name}</span>

        <span style={{
          fontSize: '10px', fontWeight: '700', letterSpacing: '0.4px',
          padding: '2px 7px', borderRadius: '20px',
          background: 'var(--color-background)', color: 'var(--color-text-muted)',
        }}>{actionLabel(ev.action_type)}</span>

        {ev.rule_id && (
          <span style={{ fontSize: '10px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-mono)' }}>
            {ev.rule_id}
          </span>
        )}

        <div style={{ flex: 1 }} />

        {ev.action_type === 'fracture' && ev.accepted !== undefined && (
          <span style={{
            fontSize: '11px', fontWeight: '700', padding: '2px 9px', borderRadius: '20px',
            background: ev.accepted ? 'var(--color-success)22' : 'var(--color-danger)22',
            color: ev.accepted ? 'var(--color-success)' : 'var(--color-danger)',
          }}>{ev.accepted ? 'ACEITA' : 'REJEITADA'}</span>
        )}

        <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontVariantNumeric: 'tabular-nums' }}>
          R{ev.round}
        </span>
        <span style={{ fontSize: '11px', color: tensionColor(ev.tension), fontVariantNumeric: 'tabular-nums' }}>
          T{Math.round(ev.tension * 100)}%
        </span>
        <span style={{ fontSize: '10px', color: 'var(--color-text-muted)' }}>{fmtTime(ev.ts)}</span>
      </div>

      {/* snippet */}
      {ev.snippet && (
        <div style={{ fontSize: '13px', color: 'var(--color-text)', lineHeight: '1.55', paddingLeft: '2px' }}>
          {ev.snippet}
        </div>
      )}

      {/* tokens */}
      {ev.tokens_used > 0 && (
        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>
          {ev.tokens_used.toLocaleString()} tokens · acumulado: {ev.total_tokens.toLocaleString()}
        </div>
      )}
    </div>
  )
}

function RoundDivider({ round }: { round: number }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', margin: '4px 0' }}>
      <div style={{ flex: 1, height: '1px', background: 'var(--color-border)' }} />
      <span style={{ fontSize: '11px', fontWeight: '700', color: 'var(--color-accent)', letterSpacing: '1px', whiteSpace: 'nowrap' }}>
        — RODADA {round} —
      </span>
      <div style={{ flex: 1, height: '1px', background: 'var(--color-border)' }} />
    </div>
  )
}

// ── main page ─────────────────────────────────────────────────────────────────

export default function LiveActivityPage({ simId, onNavigate }: { simId: string; onNavigate: (p: Page, simId?: string) => void }) {
  const [events, setEvents] = useState<ActivityEvent[]>([])
  const [simInfo, setSimInfo] = useState<SimInfo | null>(null)
  const [connected, setConnected] = useState(false)
  const [done, setDone] = useState(false)
  const [filter, setFilter] = useState<Filter>('all')
  const [autoScroll, setAutoScroll] = useState(true)
  const [newIds, setNewIds] = useState<Set<number>>(new Set())
  const feedRef = useRef<HTMLDivElement>(null)
  const esRef = useRef<EventSource | null>(null)

  // Fetch simulation metadata
  useEffect(() => {
    fetch(`/api/v1/simulations/${simId}`)
      .then(r => r.ok ? r.json() : null)
      .then(d => { if (d) setSimInfo(d) })
      .catch(() => {})
  }, [simId])

  // SSE connection
  useEffect(() => {
    const es = new EventSource(`/api/v1/simulations/${simId}/activity`)
    esRef.current = es

    es.addEventListener('connected', () => setConnected(true))
    es.addEventListener('done', () => { setDone(true); setConnected(false); es.close() })
    es.addEventListener('timeout', () => { setDone(true); setConnected(false); es.close() })

    es.addEventListener('activity', (e: MessageEvent) => {
      try {
        const ev: ActivityEvent = JSON.parse(e.data)
        setEvents(prev => {
          const next = [...prev, ev]
          // Mark as "new" for fade animation
          setNewIds(ids => {
            const s = new Set(ids)
            s.add(next.length - 1)
            setTimeout(() => setNewIds(cur => { const c = new Set(cur); c.delete(next.length - 1); return c }), 600)
            return s
          })
          return next
        })
      } catch {}
    })

    es.onerror = () => {
      // If the simulation is already done, this is expected
      setConnected(false)
    }

    return () => { es.close() }
  }, [simId])

  // Auto-scroll
  useEffect(() => {
    if (!autoScroll || !feedRef.current) return
    feedRef.current.scrollTop = feedRef.current.scrollHeight
  }, [events, autoScroll])

  // Pause auto-scroll when user scrolls up
  const handleScroll = () => {
    if (!feedRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = feedRef.current
    setAutoScroll(scrollHeight - scrollTop - clientHeight < 60)
  }

  // Filter events
  const visible = events.filter(ev => {
    if (filter === 'disruptor') return ev.archetype === 'disruptor'
    if (filter === 'fracture') return ev.action_type === 'fracture'
    return true
  })

  // Stats
  const totalTokens = events.length > 0 ? events[events.length - 1].total_tokens : 0
  const currentRound = events.length > 0 ? events[events.length - 1].round : 0
  const currentTension = events.length > 0 ? events[events.length - 1].tension : 0
  const fractureCount = events.filter(e => e.action_type === 'fracture').length
  const acceptedCount = events.filter(e => e.action_type === 'fracture' && e.accepted).length
  const disruptorCount = events.filter(e => e.archetype === 'disruptor').length

  // Group visible events by round for dividers
  const grouped: Array<{ type: 'divider'; round: number } | { type: 'event'; ev: ActivityEvent; idx: number }> = []
  let lastRound = -1
  visible.forEach((ev, _i) => {
    if (ev.round !== lastRound) {
      grouped.push({ type: 'divider', round: ev.round })
      lastRound = ev.round
    }
    grouped.push({ type: 'event', ev, idx: events.indexOf(ev) })
  })

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>

      {/* ── top bar ── */}
      <div style={{ padding: '16px 24px', borderBottom: '1px solid var(--color-border)', background: 'var(--color-surface)', flexShrink: 0 }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px', flexWrap: 'wrap' }}>
          <button onClick={() => onNavigate('simulations')} style={backBtn}>← Voltar</button>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap', marginBottom: '10px' }}>
              <h1 style={{ margin: 0, fontSize: '16px', fontWeight: '700', color: 'var(--color-text)' }}>
                Atividade ao Vivo
              </h1>
              <StatusPill connected={connected} done={done} />
              {simInfo && (
                <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '400px' }}>
                  {simInfo.question}
                </span>
              )}
            </div>

            {/* stats strip */}
            <div style={{ display: 'flex', gap: '24px', flexWrap: 'wrap', alignItems: 'center' }}>
              <TokenMeter total={totalTokens} label="Tokens consumidos" />

              <StatChip label="Rodada" value={simInfo ? `${currentRound}/${simInfo.rounds}` : String(currentRound)} />
              <StatChip label="Tensão" value={`${Math.round(currentTension * 100)}%`} color={tensionColor(currentTension)} />
              <StatChip label="Fraturas" value={`${acceptedCount}/${fractureCount}`} color="var(--color-danger)" />
              <StatChip label="Disruptores" value={String(disruptorCount)} color="var(--color-danger)" />
              <StatChip label="Eventos" value={String(events.length)} />
            </div>
          </div>
        </div>

        {/* filter tabs */}
        <div style={{ display: 'flex', gap: '6px', marginTop: '12px' }}>
          {(['all', 'disruptor', 'fracture'] as Filter[]).map(f => (
            <button key={f} onClick={() => setFilter(f)} style={{
              padding: '5px 14px', borderRadius: '20px', border: 'none', cursor: 'pointer', fontSize: '12px', fontWeight: '600',
              background: filter === f ? 'var(--color-accent)' : 'var(--color-background)',
              color: filter === f ? '#fff' : 'var(--color-text-muted)',
            }}>
              {f === 'all' ? `Todos (${events.length})` : f === 'disruptor' ? `Disruptores (${events.filter(e => e.archetype === 'disruptor').length})` : `Fraturas (${fractureCount})`}
            </button>
          ))}

          <div style={{ flex: 1 }} />

          {!autoScroll && (
            <button onClick={() => { setAutoScroll(true); feedRef.current?.scrollTo({ top: feedRef.current.scrollHeight, behavior: 'smooth' }) }}
              style={{ padding: '5px 14px', borderRadius: '20px', border: '1px solid var(--color-accent)', background: 'transparent', color: 'var(--color-accent)', fontSize: '12px', fontWeight: '600', cursor: 'pointer' }}>
              ↓ Ir para o final
            </button>
          )}
        </div>
      </div>

      {/* ── feed ── */}
      <div
        ref={feedRef}
        onScroll={handleScroll}
        style={{ flex: 1, overflowY: 'auto', padding: '16px 24px', display: 'flex', flexDirection: 'column', gap: '8px' }}
      >
        {visible.length === 0 && (
          <div style={{ textAlign: 'center', color: 'var(--color-text-muted)', paddingTop: '80px' }}>
            {connected
              ? <><div style={{ fontSize: '28px', marginBottom: '10px' }}>⏳</div><div>Aguardando atividade dos agentes…</div></>
              : done
                ? <><div style={{ fontSize: '28px', marginBottom: '10px' }}>✓</div><div>Simulação encerrada</div></>
                : <><div style={{ fontSize: '28px', marginBottom: '10px' }}>◉</div><div>Conectando ao feed de atividade…</div></>
            }
          </div>
        )}

        {grouped.map((item, i) =>
          item.type === 'divider'
            ? <RoundDivider key={`div-${item.round}-${i}`} round={item.round} />
            : <EventBubble key={`ev-${item.idx}`} ev={item.ev} isNew={newIds.has(item.idx)} />
        )}

        {done && events.length > 0 && (
          <div style={{ textAlign: 'center', padding: '20px', color: 'var(--color-text-muted)', fontSize: '13px' }}>
            ── Simulação encerrada · {events.length} eventos · {totalTokens.toLocaleString()} tokens ──
          </div>
        )}
      </div>
    </div>
  )
}

// ── mini components ───────────────────────────────────────────────────────────

function StatusPill({ connected, done }: { connected: boolean; done: boolean }) {
  const color = done ? 'var(--color-text-muted)' : connected ? 'var(--color-success)' : 'var(--color-warning)'
  const label = done ? 'ENCERRADO' : connected ? 'AO VIVO' : 'CONECTANDO'
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '5px', padding: '3px 10px', borderRadius: '20px', background: `${color}22`, border: `1px solid ${color}44` }}>
      {!done && (
        <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: color, animation: connected ? 'pulse-dot 1.5s ease-in-out infinite' : 'none' }} />
      )}
      <span style={{ fontSize: '10px', fontWeight: '700', letterSpacing: '0.5px', color }}>{label}</span>
    </div>
  )
}

function StatChip({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1px' }}>
      <span style={{ fontSize: '10px', color: 'var(--color-text-muted)', letterSpacing: '0.3px' }}>{label}</span>
      <span style={{ fontSize: '15px', fontWeight: '700', color: color ?? 'var(--color-text)', fontVariantNumeric: 'tabular-nums' }}>{value}</span>
    </div>
  )
}

const backBtn: React.CSSProperties = {
  background: 'none', border: 'none', color: 'var(--color-text-muted)', fontSize: '13px',
  cursor: 'pointer', padding: '0', flexShrink: 0, marginTop: '2px',
}
