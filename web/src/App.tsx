import { BrowserRouter, Link, Route, Routes } from 'react-router-dom'
import { RunDetailPage } from './pages/RunDetailPage'
import { WorkflowRunsPage } from './pages/WorkflowRunsPage'
import { WorkflowsPage } from './pages/WorkflowsPage'

export function App() {
  return (
    <BrowserRouter>
      <div style={{ maxWidth: 960, margin: '0 auto', padding: '24px 16px' }}>
        <header
          style={{
            marginBottom: 32,
            borderBottom: '1px solid #e5e7eb',
            paddingBottom: 16,
            display: 'flex',
            alignItems: 'center',
          }}
        >
          <Link to="/" style={{ textDecoration: 'none', color: '#111827' }}>
            <span style={{ fontWeight: 700, fontSize: 20, letterSpacing: '-0.01em' }}>
              MiniFlow
            </span>
          </Link>
        </header>
        <main>
          <Routes>
            <Route path="/" element={<WorkflowsPage />} />
            <Route path="/workflows/:workflowId" element={<WorkflowRunsPage />} />
            <Route path="/runs/:runId" element={<RunDetailPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}
