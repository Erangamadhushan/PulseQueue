
import { useEffect, useState } from 'react'

type HealthResponse = {
  status: string
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    async function checkHealth() {
      try {
        const response = await fetch('/healthz', {
          signal: controller.signal,
        })

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`)
        }

        const data: HealthResponse = await response.json()
        setHealth(data)
      } catch (err) {
        if (controller.signal.aborted) return

        setError(
          err instanceof Error ? err.message : 'Request failed',
        )
      }
    }

    void checkHealth()

    return () => controller.abort()
  }, [])

  return (
    <main>
      <h1>PulseQueue</h1>
      <p>Go Background Job Engine</p>

      {health ? (
        <p>API status: {health.status}</p>
      ) : error ? (
        <p>API connection failed: {error}</p>
      ) : (
        <p>Connecting to API...</p>
      )}
    </main>
  )
}

export default App
