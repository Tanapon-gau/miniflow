import type { RunStatus, TaskStatus } from '../types'

const STATUS_STYLE: Record<string, { background: string; color: string }> = {
  pending: { background: '#e5e7eb', color: '#374151' },
  queued: { background: '#dbeafe', color: '#1d4ed8' },
  running: { background: '#fef3c7', color: '#92400e' },
  success: { background: '#d1fae5', color: '#065f46' },
  failed: { background: '#fee2e2', color: '#991b1b' },
  retrying: { background: '#ffedd5', color: '#9a3412' },
  cancelled: { background: '#f3f4f6', color: '#6b7280' },
}

interface Props {
  status: TaskStatus | RunStatus
}

export function StatusBadge({ status }: Props) {
  const style = STATUS_STYLE[status] ?? { background: '#f3f4f6', color: '#6b7280' }
  return (
    <span
      style={{
        ...style,
        padding: '2px 8px',
        borderRadius: 4,
        fontSize: 12,
        fontWeight: 600,
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
      }}
    >
      {status}
    </span>
  )
}
