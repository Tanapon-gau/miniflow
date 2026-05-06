import type { Task } from '../types'
import { StatusBadge } from './StatusBadge'

function formatDuration(start: string | null, end: string | null): string {
  if (!start) return '—'
  const startMs = new Date(start).getTime()
  const endMs = end ? new Date(end).getTime() : Date.now()
  const seconds = Math.round((endMs - startMs) / 1000)
  return seconds < 60 ? `${seconds}s` : `${Math.floor(seconds / 60)}m ${seconds % 60}s`
}

interface Props {
  tasks: Task[]
}

export function TaskTable({ tasks }: Props) {
  return (
    <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
      <thead>
        <tr style={{ borderBottom: '2px solid #e5e7eb' }}>
          <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Task</th>
          <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Type</th>
          <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Status</th>
          <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Duration</th>
          <th style={{ textAlign: 'left', padding: '8px 12px', fontWeight: 600 }}>Attempt</th>
        </tr>
      </thead>
      <tbody>
        {tasks.map((task) => (
          <tr key={task.id} style={{ borderBottom: '1px solid #f3f4f6' }}>
            <td style={{ padding: '10px 12px', fontFamily: 'monospace', fontWeight: 500 }}>
              {task.name}
            </td>
            <td style={{ padding: '10px 12px', color: '#6b7280' }}>{task.type}</td>
            <td style={{ padding: '10px 12px' }}>
              <StatusBadge status={task.status} />
            </td>
            <td style={{ padding: '10px 12px', color: '#6b7280' }}>
              {formatDuration(task.started_at, task.finished_at)}
            </td>
            <td style={{ padding: '10px 12px', color: '#6b7280' }}>{task.attempt}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
