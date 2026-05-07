import type { Run, RunDetail, Workflow } from './types'

const API_BASE = 'http://localhost:8000'

async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!response.ok) {
    throw new Error(`${init?.method ?? 'GET'} ${path} failed with status ${response.status}`)
  }
  return response.json() as Promise<T>
}

export const api = {
  listWorkflows: () => fetchJSON<Workflow[]>('/workflows/'),
  getWorkflow: (id: string) => fetchJSON<Workflow>(`/workflows/${id}`),
  listRuns: (workflowId?: string) =>
    fetchJSON<Run[]>(`/runs${workflowId ? `?workflow_id=${workflowId}` : ''}`),
  getRun: (id: string) => fetchJSON<RunDetail>(`/runs/${id}`),
  triggerRun: (workflowId: string) =>
    fetchJSON<RunDetail>(`/workflows/${workflowId}/runs`, { method: 'POST' }),
  cancelRun: (runId: string) =>
    fetchJSON<RunDetail>(`/runs/${runId}/cancel`, { method: 'POST' }),
}
