import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api'
import { StatusBadge } from '../components/StatusBadge'
import type { Run, Workflow } from '../types'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

export function WorkflowRunsPage() {
  const { workflowId } = useParams<{ workflowId: string }>()
  const [workflow, setWorkflow] = useState<Workflow | null>(null)
  const [runs, setRuns] = useState<Run[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!workflowId) return
    Promise.all([api.getWorkflow(workflowId), api.listRuns(workflowId)])
      .then(([wf, rns]) => {
        setWorkflow(wf)
        setRuns(rns)
      })
      .catch((err: unknown) => setError(String(err)))
  }, [workflowId])

  if (error) return <p style={{ color: '#991b1b' }}>{error}</p>
  if (!workflow) return <p style={{ color: '#6b7280' }}>Loading…</p>

  return (
    <div>
      <p style={{ color: '#6b7280', margin: '0 0 8px' }}>
        <Link to="/">Workflows</Link> /
      </p>
      <h1 style={{ marginTop: 0 }}>{workflow.name}</h1>
      <h2>Runs</h2>
      {runs.length === 0 && <p style={{ color: '#6b7280' }}>No runs yet.</p>}
      {runs.length > 0 && (
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
          <thead>
            <tr style={{ borderBottom: '2px solid #e5e7eb' }}>
              <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Run ID</th>
              <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Status</th>
              <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Triggered</th>
            </tr>
          </thead>
          <tbody>
            {runs.map((run) => (
              <tr key={run.id} style={{ borderBottom: '1px solid #f3f4f6' }}>
                <td style={{ padding: '10px 12px', fontFamily: 'monospace' }}>
                  <Link to={`/runs/${run.id}`}>{run.id.slice(0, 8)}…</Link>
                </td>
                <td style={{ padding: '10px 12px' }}>
                  <StatusBadge status={run.status} />
                </td>
                <td style={{ padding: '10px 12px', color: '#6b7280' }}>
                  {formatDate(run.triggered_at)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
