<template>
  <div ref="container" class="graph-container">
    <svg ref="svgEl" class="graph-svg">
      <defs>
        <marker
          id="arrowhead"
          viewBox="0 -5 10 10"
          refX="18"
          refY="0"
          markerWidth="6"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 0,-5 L 10,0 L 0,5" fill="#5a6478" />
        </marker>
        <marker
          id="arrowhead-highlight"
          viewBox="0 -5 10 10"
          refX="18"
          refY="0"
          markerWidth="6"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 0,-5 L 10,0 L 0,5" fill="#7aa2f7" />
        </marker>
      </defs>
      <g ref="gZoom">
        <g ref="gLinks" class="links"></g>
        <g ref="gNodes" class="nodes"></g>
      </g>
    </svg>
    <div v-if="hoveredNode" class="tooltip">
      <div class="tooltip-title">{{ hoveredNode.title }}</div>
      <div class="tooltip-meta">
        出链: {{ hoveredNode.links?.length || 0 }} | 入链: {{ hoveredNode.linkFrom?.length || 0 }}
      </div>
      <div class="tooltip-path">{{ hoveredNode.id }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import * as d3 from 'd3'
import type { GraphData, NoteNode } from '../api'

interface Props {
  data: GraphData | null
  selectedId: string | null
}

interface Emits {
  (e: 'select', node: NoteNode): void
  (e: 'open', node: NoteNode): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const container = ref<HTMLDivElement | null>(null)
const svgEl = ref<SVGSVGElement | null>(null)
const gZoom = ref<SVGGElement | null>(null)
const gLinks = ref<SVGGElement | null>(null)
const gNodes = ref<SVGGElement | null>(null)
const hoveredNode = ref<NoteNode | null>(null)

interface SimNode extends d3.SimulationNodeDatum {
  id: string
  title: string
  path: string
  modTime: number
  links: string[]
  linkFrom?: string[]
  degree: number
}

interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  source: SimNode | string
  target: SimNode | string
}

let simulation: d3.Simulation<SimNode, SimLink> | null = null
let zoomBehavior: d3.ZoomBehavior<SVGSVGElement, unknown> | null = null
let resizeObserver: ResizeObserver | null = null
let width = 0
let height = 0

function getConnected(nodeId: string): Set<string> {
  const set = new Set<string>([nodeId])
  if (!props.data) return set
  for (const link of props.data.links) {
    if (link.source === nodeId) set.add(typeof link.target === 'string' ? link.target : link.target.id)
    if (link.target === nodeId) set.add(typeof link.source === 'string' ? link.source : link.source.id)
  }
  return set
}

function renderGraph() {
  if (!svgEl.value || !gZoom.value || !gLinks.value || !gNodes.value || !props.data) return

  const svg = d3.select(svgEl.value)
  const gZoomSel = d3.select(gZoom.value)
  const gLinksSel = d3.select(gLinks.value)
  const gNodesSel = d3.select(gNodes.value)

  gZoomSel.selectAll('*').remove()
  if (simulation) {
    simulation.stop()
    simulation = null
  }

  const nodeMap = new Map<string, SimNode>()
  const simNodes: SimNode[] = props.data.nodes.map((n) => {
    const degree = (n.links?.length || 0) + (n.linkFrom?.length || 0)
    const node: SimNode = {
      id: n.id,
      title: n.title,
      path: n.path,
      modTime: n.modTime,
      links: n.links,
      linkFrom: n.linkFrom,
      degree,
    }
    nodeMap.set(n.id, node)
    return node
  })

  const simLinks: SimLink[] = props.data.links.map((l) => ({
    source: l.source,
    target: l.target,
  }))

  if (simNodes.length === 0) return

  const minR = 5
  const maxR = 24
  const degrees = simNodes.map((n) => n.degree)
  const minDeg = Math.min(...degrees)
  const maxDeg = Math.max(...degrees)
  const radiusScale = d3.scaleSqrt().domain([minDeg, Math.max(maxDeg, 1)]).range([minR, maxR])

  const colorScale = d3.scaleOrdinal<string, string>(d3.schemeTableau10)

  simulation = d3
    .forceSimulation<SimNode>(simNodes)
    .force(
      'link',
      d3
        .forceLink<SimNode, SimLink>(simLinks)
        .id((d) => d.id)
        .distance(100)
        .strength(0.6)
    )
    .force('charge', d3.forceManyBody<SimNode>().strength(-350))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide<SimNode>().radius((d) => radiusScale(d.degree) + 4))

  const linkSel = gLinksSel
    .selectAll<SVGLineElement, SimLink>('line')
    .data(simLinks)
    .join('line')
    .attr('stroke', '#3d4658')
    .attr('stroke-opacity', 0.6)
    .attr('stroke-width', 1.2)
    .attr('marker-end', 'url(#arrowhead)')

  const nodeGroup = gNodesSel
    .selectAll<SVGGElement, SimNode>('g.node')
    .data(simNodes, (d) => d.id)
    .join('g')
    .attr('class', 'node')
    .style('cursor', 'pointer')
    .call(
      d3
        .drag<SVGGElement, SimNode>()
        .on('start', (event, d) => {
          if (!event.active) simulation?.alphaTarget(0.3).restart()
          d.fx = d.x
          d.fy = d.y
        })
        .on('drag', (event, d) => {
          d.fx = event.x
          d.fy = event.y
        })
        .on('end', (event, d) => {
          if (!event.active) simulation?.alphaTarget(0)
          d.fx = null
          d.fy = null
        })
    )

  nodeGroup
    .append('circle')
    .attr('r', (d) => radiusScale(d.degree))
    .attr('fill', (d) => {
      const bucket = Math.floor(colorScale.domain().indexOf(d.id) % 10)
      return bucket >= 0 ? colorScale(d.id) : '#7aa2f7'
    })
    .attr('stroke', '#1a1d24')
    .attr('stroke-width', 1.5)
    .attr('opacity', 0.92)

  nodeGroup
    .append('text')
    .text((d) => d.title)
    .attr('text-anchor', 'middle')
    .attr('dy', (d) => -radiusScale(d.degree) - 6)
    .attr('fill', '#c8cfdb')
    .attr('font-size', '11px')
    .attr('font-weight', '500')
    .attr('pointer-events', 'none')
    .style('text-shadow', '0 1px 2px rgba(0,0,0,0.8)')

  function updateHighlight(selected: string | null) {
    const connected = selected ? getConnected(selected) : null
    nodeGroup
      .select('circle')
      .attr('opacity', (d) => {
        if (!selected) return 0.92
        return connected!.has(d.id) ? 1 : 0.15
      })
      .attr('stroke', (d) => (d.id === selected ? '#ffffff' : '#1a1d24'))
      .attr('stroke-width', (d) => (d.id === selected ? 3 : 1.5))

    nodeGroup
      .select('text')
      .attr('opacity', (d) => {
        if (!selected) return 1
        return connected!.has(d.id) ? 1 : 0.2
      })

    linkSel
      .attr('stroke-opacity', (d) => {
        if (!selected) return 0.6
        const src = typeof d.source === 'object' ? d.source.id : d.source
        const tgt = typeof d.target === 'object' ? d.target.id : d.target
        return src === selected || tgt === selected ? 1 : 0.08
      })
      .attr('stroke', (d) => {
        if (!selected) return '#3d4658'
        const src = typeof d.source === 'object' ? d.source.id : d.source
        const tgt = typeof d.target === 'object' ? d.target.id : d.target
        return src === selected || tgt === selected ? '#7aa2f7' : '#3d4658'
      })
      .attr('marker-end', (d) => {
        if (!selected) return 'url(#arrowhead)'
        const src = typeof d.source === 'object' ? d.source.id : d.source
        const tgt = typeof d.target === 'object' ? d.target.id : d.target
        return src === selected || tgt === selected ? 'url(#arrowhead-highlight)' : 'url(#arrowhead)'
      })
  }

  nodeGroup
    .on('mouseenter', function (event, d) {
      hoveredNode.value = d
      updateHighlight(d.id)
      d3.select(this).select('circle').attr('opacity', 1)
    })
    .on('mouseleave', function () {
      hoveredNode.value = null
      updateHighlight(props.selectedId)
    })
    .on('click', function (event, d) {
      event.stopPropagation()
      updateHighlight(d.id)
      const plainNode: NoteNode = {
        id: d.id,
        title: d.title,
        path: d.path,
        modTime: d.modTime,
        links: d.links,
        linkFrom: d.linkFrom,
      }
      emit('select', plainNode)
    })
    .on('dblclick', function (_event, d) {
      const plainNode: NoteNode = {
        id: d.id,
        title: d.title,
        path: d.path,
        modTime: d.modTime,
        links: d.links,
        linkFrom: d.linkFrom,
      }
      emit('open', plainNode)
    })

  svg.on('click', () => {
    updateHighlight(null)
    emit('select', {} as NoteNode)
  })

  simulation.on('tick', () => {
    linkSel
      .attr('x1', (d) => (typeof d.source === 'object' ? d.source.x ?? 0 : 0))
      .attr('y1', (d) => (typeof d.source === 'object' ? d.source.y ?? 0 : 0))
      .attr('x2', (d) => (typeof d.target === 'object' ? d.target.x ?? 0 : 0))
      .attr('y2', (d) => (typeof d.target === 'object' ? d.target.y ?? 0 : 0))

    nodeGroup.attr('transform', (d) => `translate(${d.x ?? 0},${d.y ?? 0})`)
  })

  updateHighlight(props.selectedId)
}

function setupZoom() {
  if (!svgEl.value || !gZoom.value) return
  const svg = d3.select(svgEl.value)
  const gZoomSel = d3.select(gZoom.value)

  zoomBehavior = d3
    .zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.1, 4])
    .on('zoom', (event) => {
      gZoomSel.attr('transform', event.transform.toString())
    })

  svg.call(zoomBehavior)
}

function onResize() {
  if (!container.value) return
  const rect = container.value.getBoundingClientRect()
  width = rect.width
  height = rect.height
  if (svgEl.value) {
    d3.select(svgEl.value).attr('width', width).attr('height', height)
  }
  if (simulation) {
    simulation.force('center', d3.forceCenter(width / 2, height / 2))
    simulation.alpha(0.3).restart()
  }
}

function resetView() {
  if (!svgEl.value || !zoomBehavior) return
  d3.select(svgEl.value)
    .transition()
    .duration(600)
    .call(zoomBehavior.transform, d3.zoomIdentity)
}

watch(
  () => props.data,
  () => {
    nextTick(() => {
      onResize()
      renderGraph()
    })
  },
  { deep: true }
)

watch(
  () => props.selectedId,
  (newId) => {
    if (!svgEl.value || !gNodes.value || !gLinks.value) return
    if (!props.data) return

    const connected = newId ? getConnected(newId) : null
    const gNodesSel = d3.select(gNodes.value)
    const gLinksSel = d3.select(gLinks.value)

    gNodesSel
      .selectAll<SVGGElement, SimNode>('g.node')
      .select('circle')
      .attr('opacity', (d) => {
        if (!newId) return 0.92
        return connected!.has(d.id) ? 1 : 0.15
      })
      .attr('stroke', (d) => (d.id === newId ? '#ffffff' : '#1a1d24'))
      .attr('stroke-width', (d) => (d.id === newId ? 3 : 1.5))

    gNodesSel
      .selectAll<SVGGElement, SimNode>('g.node')
      .select('text')
      .attr('opacity', (d) => {
        if (!newId) return 1
        return connected!.has(d.id) ? 1 : 0.2
      })

    gLinksSel
      .selectAll<SVGLineElement, SimLink>('line')
      .attr('stroke-opacity', (d) => {
        if (!newId) return 0.6
        const src = typeof d.source === 'object' ? d.source.id : d.source
        const tgt = typeof d.target === 'object' ? d.target.id : d.target
        return src === newId || tgt === newId ? 1 : 0.08
      })
      .attr('stroke', (d) => {
        if (!newId) return '#3d4658'
        const src = typeof d.source === 'object' ? d.source.id : d.source
        const tgt = typeof d.target === 'object' ? d.target.id : d.target
        return src === newId || tgt === newId ? '#7aa2f7' : '#3d4658'
      })
  }
)

onMounted(() => {
  onResize()
  setupZoom()
  if (container.value) {
    resizeObserver = new ResizeObserver(onResize)
    resizeObserver.observe(container.value)
  }
  if (props.data) {
    renderGraph()
  }
})

onBeforeUnmount(() => {
  if (simulation) simulation.stop()
  if (resizeObserver && container.value) resizeObserver.unobserve(container.value)
})

defineExpose({ resetView })
</script>

<style scoped>
.graph-container {
  position: relative;
  width: 100%;
  height: 100%;
  background: radial-gradient(ellipse at center, #151922 0%, #0f1115 100%);
  overflow: hidden;
}

.graph-svg {
  display: block;
}

.tooltip {
  position: absolute;
  pointer-events: none;
  background: rgba(22, 26, 36, 0.95);
  border: 1px solid #2d3340;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
  max-width: 320px;
  bottom: 16px;
  left: 16px;
  z-index: 10;
}

.tooltip-title {
  font-weight: 600;
  color: #e8e8e8;
  margin-bottom: 4px;
  font-size: 13px;
}

.tooltip-meta {
  color: #8892a6;
  font-size: 11px;
  margin-bottom: 4px;
}

.tooltip-path {
  color: #656d80;
  font-size: 10px;
  word-break: break-all;
  font-family: 'SF Mono', Menlo, monospace;
}

.node text {
  user-select: none;
}
</style>
