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
  const steps = ['Bem-vindo', 'Perfil da Empresa', 'Chaves de IA', 'Privacidade', 'Pronto']

  async function handleFinish() {
    setSaving(true)
    setError('')
    try {
      // Save company
      await fetch('/api/v1/company', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(company) })
      // Save keys
      for (const [provider, key] of Object.entries(keys)) {
        if (key.trim()) {
          await fetch('/api/v1/keys/validate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ provider, key: key.trim() }) })
        }
      }
      // Save telemetry preference
      await fetch('/api/v1/telemetry', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ enabled: telemetryEnabled }) })
      // Mark onboarding complete
      await fetch('/api/v1/onboarding/complete', { method: 'POST' })
      onComplete()
    } catch {
      setError('Falha ao salvar. Verifique se o FRACTURE está em execução.')
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
              Bem-vindo ao <span style={{ color: 'var(--color-accent)' }}>FRACTURE</span>
            </div>
            <p style={{ color: 'var(--color-text-muted)', lineHeight: '1.7', marginBottom: '32px' }}>
              FRACTURE simula como as regras de mercado se rompem — e como você pode ser o primeiro a quebrá-las.
              Vamos configurar seu espaço de trabalho em 4 etapas rápidas.
            </p>
            <button style={Btn} onClick={() => setStep(1)}>Começar →</button>
          </div>
        )}

        {step === 1 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px', color: 'var(--color-text)' }}>Perfil da Empresa</div>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '24px', fontSize: '13px' }}>
              Este contexto é injetado em cada simulação para que os agentes entendam seu mercado.
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
              <div>
                <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Nome da Empresa</label>
                <input style={Input} value={company.name} onChange={e => setCompany(c => ({...c, name: e.target.value}))} placeholder="Acme Corp" />
              </div>
              <div>
                <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Setor / Indústria</label>
                <input style={Input} value={company.sector} onChange={e => setCompany(c => ({...c, sector: e.target.value}))} placeholder="ex.: SaaS, Varejo, Saúde" />
              </div>
              <div>
                <label style={{ fontSize: '12px', color: 'var(--color-text-muted)', display: 'block', marginBottom: '6px' }}>Breve Descrição</label>
                <textarea style={{...Input, height: '80px', resize: 'vertical'}} value={company.description} onChange={e => setCompany(c => ({...c, description: e.target.value}))} placeholder="O que sua empresa faz? Quais são seus principais concorrentes?" />
              </div>
            </div>
            <div style={{ display: 'flex', gap: '12px', marginTop: '24px' }}>
              <button style={{...Btn, background: 'var(--color-surface-2)', color: 'var(--color-text)'}} onClick={() => setStep(0)}>← Voltar</button>
              <button style={Btn} onClick={() => setStep(2)} disabled={!company.name}>Próximo →</button>
            </div>
          </div>
        )}

        {step === 2 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px', color: 'var(--color-text)' }}>Chaves de IA</div>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '24px', fontSize: '13px' }}>
              As chaves são armazenadas localmente na sua máquina. Elas nunca saem do seu computador.
              Adicione pelo menos uma para executar simulações.
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
              <button style={{...Btn, background: 'var(--color-surface-2)', color: 'var(--color-text)'}} onClick={() => setStep(1)}>← Voltar</button>
              <button style={Btn} onClick={() => setStep(3)} disabled={Object.values(keys).every(k => !k.trim())}>Próximo →</button>
            </div>
          </div>
        )}

        {step === 3 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px', color: 'var(--color-text)' }}>
              Privacidade &amp; Telemetria
            </div>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '20px', fontSize: '13px', lineHeight: '1.7' }}>
              FRACTURE coleta <strong style={{ color: 'var(--color-text)' }}>dados de uso anônimos</strong> para entender
              como a ferramenta está sendo utilizada e melhorar versões futuras. Nenhum dado pessoal, conteúdo de simulação
              ou chave de API é coletado.
            </p>

            {/* What is collected */}
            <div style={{ padding: '14px 16px', borderRadius: '10px', background: 'var(--color-surface-2)', border: '1px solid var(--color-border)', marginBottom: '20px' }}>
              <div style={{ fontSize: '12px', fontWeight: '600', color: 'var(--color-text-muted)', marginBottom: '10px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                O que é coletado
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px' }}>
                {[
                  ['🔑', 'ID de instalação anônimo'],
                  ['🖥️', 'SO &amp; arquitetura'],
                  ['🌍', 'País (pelo IP)'],
                  ['📦', 'Versão do FRACTURE'],
                ].map(([icon, label]) => (
                  <div key={label} style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: 'var(--color-text-muted)' }}>
                    <span>{icon}</span>
                    <span dangerouslySetInnerHTML={{ __html: label }} />
                  </div>
                ))}
              </div>
              <div style={{ marginTop: '10px', fontSize: '11px', color: 'var(--color-text-muted)', opacity: 0.7 }}>
                Seu IP é mascarado (último octeto removido). Nenhum dado de simulação é enviado.
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
                  {telemetryEnabled ? 'Telemetria ativada' : 'Telemetria desativada'}
                </div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '2px' }}>
                  {telemetryEnabled
                    ? 'Ajudando a melhorar o FRACTURE — obrigado!'
                    : 'Você pode reativar a qualquer momento nas Configurações.'}
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
              <button style={{...Btn, background: 'var(--color-surface-2)', color: 'var(--color-text)'}} onClick={() => setStep(2)}>← Voltar</button>
              <button style={Btn} onClick={() => setStep(4)}>Próximo →</button>
            </div>
          </div>
        )}

        {step === 4 && (
          <div>
            <div style={{ fontSize: '20px', fontWeight: '700', marginBottom: '12px', color: 'var(--color-text)' }}>
              Você está pronto 🎯
            </div>
            <p style={{ color: 'var(--color-text-muted)', lineHeight: '1.7', marginBottom: '32px' }}>
              FRACTURE está configurado. Agora você pode executar sua primeira simulação — faça qualquer pergunta estratégica e veja o mercado se simular.
            </p>
            <div style={{ padding: '16px', borderRadius: '10px', background: 'oklch(0.65 0.22 30 / 0.1)', border: '1px solid oklch(0.65 0.22 30 / 0.3)', marginBottom: '28px' }}>
              <div style={{ fontSize: '13px', color: 'var(--color-accent)', fontWeight: '600', marginBottom: '4px' }}>Dica: Comece com uma pergunta como</div>
              <div style={{ fontSize: '13px', color: 'var(--color-text)', fontStyle: 'italic' }}>
                "Se um concorrente oferecesse nosso produto principal gratuitamente, como o mercado reagiria em 12 meses?"
              </div>
            </div>
            <button style={{...Btn, opacity: saving ? 0.6 : 1}} onClick={handleFinish} disabled={saving}>
              {saving ? 'Salvando...' : 'Iniciar FRACTURE →'}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
