import { useState } from 'react'
import './index.css'
import Sidebar from './components/Sidebar'
import HomePage from './pages/HomePage'
import NewSimulationPage from './pages/NewSimulationPage'
import SimulationsPage from './pages/SimulationsPage'
import ArchetypesPage from './pages/ArchetypesPage'
import SettingsPage from './pages/SettingsPage'
import OnboardingWizard from './components/OnboardingWizard'
import { useOnboarding } from './hooks/useOnboarding'

export type Page = 'home' | 'new-simulation' | 'simulations' | 'archetypes' | 'settings'

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('home')
  const { isOnboarded, loading } = useOnboarding()

  if (loading) {
    return (
      <div style={{ display:'flex', alignItems:'center', justifyContent:'center', height:'100%', background:'var(--color-background)' }}>
        <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:'12px' }}>
          <div style={{ width:'32px', height:'32px', border:'2px solid var(--color-accent)', borderTopColor:'transparent', borderRadius:'50%', animation:'spin 0.8s linear infinite' }} />
          <span style={{ color:'var(--color-text-muted)', fontSize:'13px' }}>Loading FRACTURE...</span>
        </div>
      </div>
    )
  }

  if (!isOnboarded) {
    return <OnboardingWizard onComplete={() => window.location.reload()} />
  }

  const renderPage = () => {
    switch (currentPage) {
      case 'home': return <HomePage onNavigate={setCurrentPage} />
      case 'new-simulation': return <NewSimulationPage onNavigate={setCurrentPage} />
      case 'simulations': return <SimulationsPage onNavigate={setCurrentPage} />
      case 'archetypes': return <ArchetypesPage />
      case 'settings': return <SettingsPage />
      default: return <HomePage onNavigate={setCurrentPage} />
    }
  }

  return (
    <div style={{ display:'flex', height:'100%', background:'var(--color-background)' }}>
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <main style={{ flex:1, overflow:'auto' }}>
        {renderPage()}
      </main>
    </div>
  )
}

export default App
