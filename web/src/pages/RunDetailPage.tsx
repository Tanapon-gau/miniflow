import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api'
import { DagView } from '../components/DagView'
import { StatusBadge } from '../components/StatusBadge'
import { TaskTable } from '../components/TaskTable'
import type { RunDetail, Workflow } from '../types'

const TERMINAL_STATUSES = new Set(['success', 'failed'])
const POLL_INTERVAL_MS = 3000

export function RunDetailPage() {
  const { runId } = useParams<{ runId: string }>()
  const [run, setRun] = useState<RunDetail | null>(null)
  const [workflow, setWorkflow] = useState<Workflow | null>(null)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  function stopPolling() {
    if (intervalRef.current !== null) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }

  useEffect(() => {
    if (!runId) return

    let workflowLoaded = false

    async function loadRun() {
      const runData = await api.getRun(runId!)
      setRun(runData)
      if (!workflowLoaded) {
        const wfData = await api.getWorkflow(runData.workflow_id)
        setWorkflow(wfData)
        workflowLoaded = true
      }
      if (TERMINAL_STATUSES.has(runData.status)) {
        stopPolling()
      }
    }

    loadRun().catch((err: unknown) => setError(String(err)))
    intervalRef.current = setInterval(() => {
      loadRun().catch(() => {})
    }, POLL_INTERVAL_MS)

    return stopPolling
  }, [runId])

  if (error) return <p style={{ color: '#991b1b' }}>{error}</p>
  if (!run || !workflow) return <p style={{ color: '#6b7280' }}>Loading…</p>

  const isActive = !TERMINAL_STATUSES.has(run.status)

  return (
    <div>
      <p style={{ color: '#6b7280', margin: '0 0 8px' }}>
        <Link to="/">Workflows</Link> /{' '}
        <Link to={`/workflows/${run.workflow_id}`}>{workflow.name}</Link> /
      </p>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 24 }}>
        <h1 style={{ margin: 0 }}>Run {run.id.slice(0, 8)}…</h1>
        <StatusBadge status={run.status} />
        {isActive && (
          <span style={{ color: '#9ca3af', fontSize: 13 }}>
            Updating every {POLL_INTERVAL_MS / 1000}s…
          </span>
        )}
      </div>

      <h2 style={{ marginTop: 0 }}>DAG</h2>
      <DagView workflow={workflow} run={run} />

      <h2 style={{ marginTop: 28 }}>Tasks</h2>
      <TaskTable tasks={run.tasks} />
    </div>
  )
}
