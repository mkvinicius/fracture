import { useState, useEffect } from 'react'

export function useOnboarding() {
  const [isOnboarded, setIsOnboarded] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/onboarding/status')
      .then(r => r.json())
      .then(d => { setIsOnboarded(d.complete); setLoading(false) })
      .catch(() => { setIsOnboarded(false); setLoading(false) })
  }, [])

  return { isOnboarded, loading }
}
