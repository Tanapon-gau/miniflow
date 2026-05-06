import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api'
import type { Workflow } from '../types'

export function WorkflowsPage() {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [triggering, setTriggering] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    api.listWorkflows().then(setWorkflows).catch((err: unknown) => setError(String(err)))
  }, [])

  async function handleTrigger(workflowId: string) {
    setTriggering(workflowId)
    try {
      const run = await api.triggerRun(workflowId)
      navigate(`/runs/${run.id}`)
    } catch (err: unknown) {
      setError(String(err))
    } finally {
      setTriggering(null)
    }
  }

  if (error) return <p style={{ color: '#991b1b' }}>{error}</p>

  return (
    <div>
      <h1 style={{ marginTop: 0 }}>Workflows</h1>
      {workflows.length === 0 && <p style={{ color: '#6b7280' }}>No workflows found.</p>}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {workflows.map((wf) => (
          <div
            key={wf.id}
            style={{
              border: '1px solid #e5e7eb',
              borderRadius: 8,
              padding: '16px 20px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <div>
              <Link
                to={`/workflows/${wf.id}`}
                style={{ fontWeight: 600, fontSize: 16, textDecoration: 'none', color: '#111827' }}
              >
                {wf.name}
              </Link>
              {wf.description && (
                <p style={{ margin: '4px 0 0', color: '#6b7280', fontSize: 14 }}>
                  {wf.description}
                </p>
              )}
              <p style={{ margin: '6px 0 0', color: '#9ca3af', fontSize: 12 }}>
                {wf.dag.tasks.length} task{wf.dag.tasks.length !== 1 ? 's' : ''}
              </p>
            </div>
            <button
              onClick={() => handleTrigger(wf.id)}
              disabled={triggering === wf.id}
              style={{
                padding: '8px 16px',
                background: '#2563eb',
                color: 'white',
                border: 'none',
                borderRadius: 6,
                cursor: triggering === wf.id ? 'not-allowed' : 'pointer',
                opacity: triggering === wf.id ? 0.7 : 1,
                fontWeight: 500,
              }}
            >
              {triggering === wf.id ? 'Triggering…' : 'Trigger Run'}
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
