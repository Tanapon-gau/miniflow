import ReactFlow, { Background, Controls } from 'reactflow'
import 'reactflow/dist/style.css'
import type { RunDetail, Workflow } from '../types'
import { buildDagLayout } from '../dagLayout'

const STATUS_COLOR: Record<string, string> = {
  pending: '#e5e7eb',
  queued: '#bfdbfe',
  running: '#fde68a',
  success: '#a7f3d0',
  failed: '#fca5a5',
  retrying: '#fed7aa',
  cancelled: '#f3f4f6',
}

interface Props {
  workflow: Workflow
  run?: RunDetail
}

export function DagView({ workflow, run }: Props) {
  const statusByName = Object.fromEntries((run?.tasks ?? []).map((t) => [t.name, t.status]))
  const { nodes: rawNodes, edges } = buildDagLayout(workflow.dag.tasks, statusByName)

  const nodes = rawNodes.map((node) => ({
    ...node,
    style: {
      background: STATUS_COLOR[node.data.status ?? 'pending'] ?? '#e5e7eb',
      border: '1px solid #d1d5db',
      borderRadius: 6,
      padding: '6px 12px',
      fontSize: 13,
      width: 200,
    },
    data: {
      ...node.data,
      label: (
        <div>
          <div style={{ fontWeight: 600 }}>{node.data.label}</div>
          <div style={{ color: '#6b7280', fontSize: 11, marginTop: 2 }}>{node.data.taskType}</div>
        </div>
      ),
    },
  }))

  return (
    <div style={{ height: 300, border: '1px solid #e5e7eb', borderRadius: 8 }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
      >
        <Background color="#f3f4f6" />
        <Controls showInteractive={false} />
      </ReactFlow>
    </div>
  )
}
