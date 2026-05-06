export type TaskStatus = 'pending' | 'queued' | 'running' | 'success' | 'failed' | 'retrying'
export type RunStatus = 'pending' | 'running' | 'success' | 'failed'

export interface Task {
  id: string
  run_id: string
  name: string
  type: string
  status: TaskStatus
  payload: Record<string, unknown> | null
  attempt: number
  max_retries: number
  timeout_seconds: number
  started_at: string | null
  finished_at: string | null
  created_at: string
}

export interface Run {
  id: string
  workflow_id: string
  status: RunStatus
  triggered_at: string
  started_at: string | null
  finished_at: string | null
}

export interface RunDetail extends Run {
  tasks: Task[]
}

export interface DAGTask {
  name: string
  type: string
  depends_on?: string[]
  [key: string]: unknown
}

export interface Workflow {
  id: string
  name: string
  description: string | null
  dag: { tasks: DAGTask[] }
  version: number
  created_at: string
  updated_at: string
}
