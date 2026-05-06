import type { DAGTask } from './types'

export interface DagNode {
  id: string
  position: { x: number; y: number }
  data: { label: string; taskType: string; status: string | undefined }
}

export interface DagEdge {
  id: string
  source: string
  target: string
}

const NODE_WIDTH = 200
const NODE_HEIGHT = 60
const H_GAP = 80
const V_GAP = 20

export function buildDagLayout(
  dagTasks: DAGTask[],
  statusByName: Record<string, string> = {},
): { nodes: DagNode[]; edges: DagEdge[] } {
  const taskMap = new Map(dagTasks.map((t) => [t.name, t]))
  const levelByName = new Map<string, number>()

  function computeLevel(name: string): number {
    if (levelByName.has(name)) return levelByName.get(name)!
    const task = taskMap.get(name)
    if (!task?.depends_on?.length) {
      levelByName.set(name, 0)
      return 0
    }
    const level = Math.max(...task.depends_on.map(computeLevel)) + 1
    levelByName.set(name, level)
    return level
  }
  dagTasks.forEach((t) => computeLevel(t.name))

  const tasksByLevel = new Map<number, string[]>()
  levelByName.forEach((level, name) => {
    if (!tasksByLevel.has(level)) tasksByLevel.set(level, [])
    tasksByLevel.get(level)!.push(name)
  })

  const positionByName = new Map<string, { x: number; y: number }>()
  tasksByLevel.forEach((names, level) => {
    names.forEach((name, index) => {
      positionByName.set(name, {
        x: level * (NODE_WIDTH + H_GAP),
        y: index * (NODE_HEIGHT + V_GAP),
      })
    })
  })

  const nodes: DagNode[] = dagTasks.map((task) => ({
    id: task.name,
    position: positionByName.get(task.name) ?? { x: 0, y: 0 },
    data: {
      label: task.name,
      taskType: task.type,
      status: statusByName[task.name],
    },
  }))

  const edges: DagEdge[] = dagTasks.flatMap((task) =>
    (task.depends_on ?? []).map((dep) => ({
      id: `${dep}->${task.name}`,
      source: dep,
      target: task.name,
    })),
  )

  return { nodes, edges }
}
