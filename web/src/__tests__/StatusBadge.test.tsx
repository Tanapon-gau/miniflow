import { render, screen } from '@testing-library/react'
import { StatusBadge } from '../components/StatusBadge'
import type { TaskStatus } from '../types'

const ALL_STATUSES: TaskStatus[] = ['pending', 'queued', 'running', 'success', 'failed', 'retrying']

test.each(ALL_STATUSES)('renders label for status "%s"', (status) => {
  render(<StatusBadge status={status} />)
  expect(screen.getByText(status)).toBeInTheDocument()
})

test('success badge has green text color', () => {
  const { container } = render(<StatusBadge status="success" />)
  const span = container.querySelector('span')!
  expect(span.style.color).toBe('rgb(6, 95, 70)')
})

test('failed badge has red text color', () => {
  const { container } = render(<StatusBadge status="failed" />)
  const span = container.querySelector('span')!
  expect(span.style.color).toBe('rgb(153, 27, 27)')
})
