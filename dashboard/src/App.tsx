import { useState, useEffect } from 'react'
import './index.css'
import Sidebar from './components/Sidebar'
import HomePage from './pages/HomePage'
import NewSimulationPage from './pages/NewSimulationPage'
import SimulationsPage from './pages/SimulationsPage'
import ArchetypesPage from './pages/ArchetypesPage'
import SettingsPage from './pages/SettingsPage'
import ResultPage from './pages/ResultPage'
import FeedbackPage from './pages/FeedbackPage'
import ComparisonPage from './pages/ComparisonPage'
import ConvergencePage from './pages/ConvergencePage'
import OnboardingWizard from './components/OnboardingWizard'
import SchedulesPage from './pages/SchedulesPage'
import AccuracyPage from './pages/AccuracyPage'
import APIKeysPage from './pages/APIKeysPage'
import { useOnboarding } from './hooks/useOnboarding'

export type Page = 'home' | 'new-simulation' | 'simulations' | 'archetypes' | 'settings' | 'result' | 'feedback' | 'comparison' | 'convergence' | 'schedules' | 'accuracy' | 'api-keys'

type UpdateInfo = {
  has_update: boolean
  current_version: string
  latest_version?: string
  release_url?: string
  release_name?: string
}

function UpdateBanner({ info, onDismiss }: { info: UpdateInfo; onDismiss: () => void }) {
  if (!info.has_update) return null
  return (
    <div style={{
      position: 'fixed', top: '12px', right: '12px', zIndex: 9999,
      background: 'oklch(0.18 0.02 240)', border: '1px solid var(--color-accent)',
      borderRadius: '10px', padding: '14px 18px', maxWidth: '340px',
      boxShadow: '0 4px 24px oklch(0 0 0 / 0.5)',
      display: 'flex', flexDirection: 'column', gap: '8px'
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <div style={{ fontSize: '13px', fontWeight: '700', color: 'var(--color-accent)' }}>
            🔴 Nova versão disponível
          </div>
          <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginTop: '2px' }}>
            v{info.current_version} → <strong style={{ color: 'var(--color-text)' }}>v{info.latest_version}</strong>
          </div>
        </div>
        <button onClick={onDismiss} style={{ background: 'none', border: 'none', color: 'var(--color-text-muted)', cursor: 'pointer', fontSize: '16px', padding: '0 0 0 12px' }}>✕</button>
      </div>
      <div style={{ display: 'flex', gap: '8px' }}>
        <a
          href={info.release_url}
          target="_blank"
          rel="noopener noreferrer"
          style={{ padding: '6px 14px', borderRadius: '6px', background: 'var(--color-accent)', color: '#fff', fontSize: '12px', fontWeight: '600', textDecoration: 'none' }}
        >
          Baixar atualização
        </a>
        <button onClick={onDismiss} style={{ padding: '6px 14px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'transparent', color: 'var(--color-text-muted)', fontSize: '12px', cursor: 'pointer' }}>
          Depois
        </button>
      </div>
    </div>
  )
}

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('home')
  const [selectedSimId, setSelectedSimId] = useState<string>('')
  const [selectedSimIds, setSelectedSimIds] = useState<string[]>([])
  const { isOnboarded, loading } = useOnboarding()
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null)
  const [updateDismissed, setUpdateDismissed] = useState(false)

  const navigate = (p: Page, simId?: string, simIds?: string[]) => {
    if (simId !== undefined) setSelectedSimId(simId)
    if (simIds !== undefined) setSelectedSimIds(simIds)
    setCurrentPage(p)
  }

  // Check for updates once on startup (non-blocking)
  useEffect(() => {
    const check = async () => {
      try {
        const res = await fetch('/api/v1/update-check')
        if (res.ok) {
          const data: UpdateInfo = await res.json()
          if (data.has_update) setUpdateInfo(data)
        }
      } catch {
        // silent fail — no internet or server not ready
      }
    }
    // Delay 3s to not block initial render
    const t = setTimeout(check, 3000)
    return () => clearTimeout(t)
  }, [])

  if (loading) {
    return (
      <div style={{ display:'flex', alignItems:'center', justifyContent:'center', height:'100%', background:'var(--color-background)' }}>
        <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:'12px' }}>
          <div style={{ width:'32px', height:'32px', border:'2px solid var(--color-accent)', borderTopColor:'transparent', borderRadius:'50%', animation:'spin 0.8s linear infinite' }} />
          <span style={{ color:'var(--color-text-muted)', fontSize:'13px' }}>Carregando FRACTURE...</span>
        </div>
      </div>
    )
  }

  if (!isOnboarded) {
    return <OnboardingWizard onComplete={() => window.location.reload()} />
  }

  const renderPage = () => {
    switch (currentPage) {
      case 'home': return <HomePage onNavigate={navigate} />
      case 'new-simulation': return <NewSimulationPage onNavigate={navigate} />
      case 'simulations': return <SimulationsPage onNavigate={navigate} />
      case 'archetypes': return <ArchetypesPage />
      case 'settings': return <SettingsPage />
      case 'result': return <ResultPage simId={selectedSimId} onNavigate={navigate} />
      case 'feedback': return <FeedbackPage simId={selectedSimId} onNavigate={navigate} />
      case 'comparison': return <ComparisonPage simIds={selectedSimIds} onNavigate={navigate} />
      case 'convergence': return <ConvergencePage simId={selectedSimId} onNavigate={navigate} />
      case 'schedules': return <SchedulesPage onNavigate={navigate} />
      case 'accuracy': return <AccuracyPage onNavigate={navigate} />
      case 'api-keys': return <APIKeysPage onNavigate={navigate} />
      default: return <HomePage onNavigate={navigate} />
    }
  }

  return (
    <div style={{ display:'flex', height:'100%', background:'var(--color-background)' }}>
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <main style={{ flex:1, overflow:'auto' }}>
        {renderPage()}
      </main>
      {updateInfo && !updateDismissed && (
        <UpdateBanner info={updateInfo} onDismiss={() => setUpdateDismissed(true)} />
      )}
    </div>
  )
}

export default App
