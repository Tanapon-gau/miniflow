import { buildDagLayout } from '../dagLayout'
import type { DAGTask } from '../types'

function task(name: string, depends_on?: string[]): DAGTask {
  return { name, type: 'shell', depends_on }
}

test('root tasks are placed at x=0', () => {
  const { nodes } = buildDagLayout([task('a'), task('b')])
  expect(nodes.find((n) => n.id === 'a')!.position.x).toBe(0)
  expect(nodes.find((n) => n.id === 'b')!.position.x).toBe(0)
})

test('dependent task is placed to the right of its dependency', () => {
  const { nodes } = buildDagLayout([task('a'), task('b', ['a'])])
  const aX = nodes.find((n) => n.id === 'a')!.position.x
  const bX = nodes.find((n) => n.id === 'b')!.position.x
  expect(bX).toBeGreaterThan(aX)
})

test('task at depth 2 is further right than task at depth 1', () => {
  const { nodes } = buildDagLayout([task('a'), task('b', ['a']), task('c', ['b'])])
  const aX = nodes.find((n) => n.id === 'a')!.position.x
  const bX = nodes.find((n) => n.id === 'b')!.position.x
  const cX = nodes.find((n) => n.id === 'c')!.position.x
  expect(bX).toBeGreaterThan(aX)
  expect(cX).toBeGreaterThan(bX)
})

test('creates one edge per dependency', () => {
  const { edges } = buildDagLayout([task('a'), task('b', ['a']), task('c', ['a'])])
  expect(edges).toHaveLength(2)
  expect(edges.map((e) => e.source)).toEqual(expect.arrayContaining(['a', 'a']))
  expect(edges.map((e) => e.target)).toEqual(expect.arrayContaining(['b', 'c']))
})

test('applies status from statusByName', () => {
  const { nodes } = buildDagLayout([task('a')], { a: 'success' })
  expect(nodes[0].data.status).toBe('success')
})

test('status is undefined when not provided', () => {
  const { nodes } = buildDagLayout([task('a')])
  expect(nodes[0].data.status).toBeUndefined()
})

test('returns empty nodes and edges for empty DAG', () => {
  const { nodes, edges } = buildDagLayout([])
  expect(nodes).toHaveLength(0)
  expect(edges).toHaveLength(0)
})
