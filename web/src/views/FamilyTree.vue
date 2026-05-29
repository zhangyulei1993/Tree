<template>
  <section class="section family-tree-page">
    <div class="container">
      <div class="intro-card">
        <div>
          <p class="eyebrow">公开关系树</p>
          <h1>{{ displayTitle }}</h1>
          <p>{{ displayDescription }}</p>
        </div>

        <div class="intro-stats">
          <div>
            <strong>{{ tree.nodes.length }}</strong>
            <span>人物节点</span>
          </div>
          <div>
            <strong>{{ tree.relationships.length }}</strong>
            <span>关系线索</span>
          </div>
        </div>
      </div>

      <div v-if="loading" class="empty">加载中...</div>

      <div v-else-if="!tree.nodes.length" class="empty">
        暂无公开关系树数据
      </div>

      <div v-else class="tree-shell">
        <div class="tree-header">
          <div>
            <h2>以 {{ coreNode?.name || '核心人物' }} 为中心的亲属关系网</h2>
            <p>外围人物按长辈、配偶亲属、后代等关系分区排列，关系线展示人物之间的亲属线索。</p>
          </div>

          <div class="legend">
            <span><i class="line-dot parent"></i>直系关系</span>
            <span><i class="line-dot spouse"></i>配偶关系</span>
            <span><i class="line-dot relative"></i>亲属关系</span>
          </div>
        </div>

        <div class="tree-scroll">
          <div
            class="tree-board"
            :style="{ width: board.width + 'px', height: board.height + 'px' }"
          >
            <svg class="relation-layer" :width="board.width" :height="board.height">
    <template v-for="(line, index) in lines" :key="line.id || index">
      <path
        :id="`family-relation-path-${line.id || index}`"
        class="relation-line"
        :class="[
          line.className || line.class || line.relation_type || line.relationType,
          {
            spouse:
              (line.relation_type || line.relationType || line.className || line.class) === 'spouse' ||
              (line.relation_name || line.relationName || line.label || line.text) === '配偶'
          }
        ]"
        :d="line.path || line.d || line.curve || line.pathD"
        fill="none"
      />

      <text class="line-label line-label-on-path" dy="-4">
        <textPath
          :href="`#family-relation-path-${line.id || index}`"
          startOffset="50%"
          text-anchor="middle"
        >
          {{ line.relation_name || line.relationName || line.label || line.text || '关系' }}
        </textPath>
      </text>
    </template>
</svg>

            <div class="center-halo"></div>

            <article
              v-for="node in positionedNodes"
              :key="node.id"
              class="person-card"
              :class="node.kind"
              :style="{ left: node.x + 'px', top: node.y + 'px' }"
            >
              <div class="avatar">{{ firstChar(node.name) }}</div>

              <div class="person-info">
                <h3>{{ node.name }}</h3>
                <p class="role">{{ node.kind === 'core' ? '关系中心' : node.role }}</p>
                <p class="desc">{{ node.description }}</p>

                <div v-if="node.kind !== 'core' && relatedText(node.id).length" class="relation-tags">
                  <span v-for="item in relatedText(node.id)" :key="item">
                    {{ item }}
                  </span>
                </div>
              </div>
            </article>
          </div>
        </div>

        <div class="relationship-summary">
          <h2>关系说明</h2>

          <div class="summary-grid">
            <article
              v-for="item in tree.relationships"
              :key="item.id"
              class="summary-card"
            >
              <div class="summary-title">
                <strong>{{ nodeName(item.from_node_id) }}</strong>
                <span>{{ item.relation_name || item.relation_type }}</span>
                <strong>{{ nodeName(item.to_node_id) }}</strong>
              </div>

              <p v-if="item.remark">{{ item.remark }}</p>
              <p v-else>公开关系树中的一条人物关系。</p>
            </article>
          </div>
        </div>

        <div class="privacy-note">
          <strong>数据说明：</strong>
          本页面展示的是官网公开关系树，仅用于历史人物关系示意，不读取、不展示用户真实家庭树数据。
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api } from '../api'

interface PublicTreeConfig {
  title: string
  subtitle: string
  description: string
}

interface PublicTreeNode {
  id: number
  name: string
  role: string
  generation: number
  description: string
  sort: number
  status: number
}

interface PublicTreeRelationship {
  id: number
  from_node_id: number
  to_node_id: number
  relation_type: string
  relation_name: string
  fromNodeId?: number
  toNodeId?: number
  relationType?: string
  relationName?: string
  remark: string
  sort: number
  status: number
}

interface PublicTreeResult {
  config: PublicTreeConfig
  nodes: PublicTreeNode[]
  relationships: PublicTreeRelationship[]
  generation_keys: number[]
  generations: Record<string, PublicTreeNode[]>
}

interface PositionedNode extends PublicTreeNode {
  x: number
  y: number
  kind: string
}

interface LineItem {
  id: number
  path: string
  pathD: string
  d: string
  curve: string
  label: string
  text: string
  relation_name: string
  relationName: string
  relation_type: string
  relationType?: string
  className: string
  class: string
  labelX: number
  labelY: number
  midX: number
  midY: number
  x: number
  y: number
  type: string
}

const loading = ref(false)

const isMobile = ref(false)
const mobileWidth = ref(375)
const mobileHeight = ref(667)

function updateMobileState() {
  if (typeof window === 'undefined') return

  isMobile.value = window.innerWidth <= 768
  mobileWidth.value = Math.max(320, Math.min(window.innerWidth - 20, 430))
  mobileHeight.value = window.innerHeight || 667
}

const tree = reactive<PublicTreeResult>({
  config: {
    title: '',
    subtitle: '',
    description: ''
  },
  nodes: [],
  relationships: [],
  generation_keys: [],
  generations: {}
})

const cardW = 232
const cardH = 112
const coreW = 285
const coreH = 132

const coreNode = computed(() => {
  return findCoreNode(tree.nodes)
})

const displayTitle = computed(() => {
  return tree.config.title || '汉武帝重要亲属关系示意'
})

const displayDescription = computed(() => {
  return tree.config.description || '本关系树仅用于官网公开展示，不读取用户真实家庭树数据。'
})

const groupedOuterNodes = computed(() => {
  const core = coreNode.value || tree.nodes[0]
  const groups = {
    elders: [] as PublicTreeNode[],
    children: [] as PublicTreeNode[],
    spouses: [] as PublicTreeNode[],
    relatives: [] as PublicTreeNode[]
  }

  tree.nodes
    .filter((node) => core && node.id !== core.id)
    .sort((a, b) => {
      if (a.generation !== b.generation) return a.generation - b.generation
      return Number(a.sort || 0) - Number(b.sort || 0)
    })
    .forEach((node) => {
      const kind = getNodeKind(node, core)

      if (kind === 'elder') groups.elders.push(node)
      else if (kind === 'child') groups.children.push(node)
      else if (kind === 'spouse') groups.spouses.push(node)
      else groups.relatives.push(node)
    })

  return groups
})

const board = computed(() => {
  if (isMobile.value) {
    return {
      width: mobileWidth.value,
      height: Math.max(470, Math.min(mobileHeight.value - 215, 620))
    }
  }

  const groups = groupedOuterNodes.value
  const topCount = Math.max(groups.elders.length, groups.children.length, 2)
  const sideCount = Math.max(groups.spouses.length + groups.relatives.length, 2)

  return {
    width: Math.max(1220, topCount * 280 + 360),
    height: Math.max(760, Math.ceil(sideCount / 2) * 165 + 430)
  }
})

const center = computed(() => {
  return {
    x: board.value.width / 2,
    y: board.value.height / 2
  }
})

const positionedNodes = computed<PositionedNode[]>(() => {
  if (!tree.nodes.length) return []

  if (isMobile.value) {
    return buildMobilePositionedNodes()
  }

  const core = coreNode.value || tree.nodes[0]
  const groups = groupedOuterNodes.value
  const result: PositionedNode[] = []

  result.push({
    ...core,
    x: center.value.x - coreW / 2,
    y: center.value.y - coreH / 2,
    kind: 'core'
  })

  placeHorizontal(result, groups.elders, 'elder', 70)
  placeHorizontal(result, groups.children, 'child', board.value.height - cardH - 70)

  const sideNodes = [
    ...groups.spouses.map((node) => ({ node, kind: 'spouse' })),
    ...groups.relatives.map((node) => ({ node, kind: 'relative' }))
  ]

  const left = sideNodes.filter((_, index) => index % 2 === 0)
  const right = sideNodes.filter((_, index) => index % 2 === 1)

  placeVertical(result, left, 'left')
  placeVertical(result, right, 'right')

  return result
})

const lines = computed<LineItem[]>(() => {
  const nodes = positionedNodes.value
  const map = new Map(nodes.map((node) => [Number(node.id), node]))

  const mobileCardW = 78
  const mobileCardH = 44
  const mobileCoreW = 116
  const mobileCoreH = 64

  function getSize(node: PositionedNode) {
    if (isMobile.value) {
      return node.kind === 'core'
        ? { w: mobileCoreW, h: mobileCoreH }
        : { w: mobileCardW, h: mobileCardH }
    }

    return node.kind === 'core'
      ? { w: coreW, h: coreH }
      : { w: cardW, h: cardH }
  }

  function clamp(value: number, min: number, max: number) {
    return Math.max(min, Math.min(max, value))
  }

  const items: Array<LineItem | null> = tree.relationships
    .map((rel, index) => {
      const fromId = Number(rel.from_node_id ?? rel.fromNodeId)
      const toId = Number(rel.to_node_id ?? rel.toNodeId)

      const from = map.get(fromId)
      const to = map.get(toId)

      if (!from || !to) return null

      const fromSize = getSize(from)
      const toSize = getSize(to)

      const x1 = from.x + fromSize.w / 2
      const y1 = from.y + fromSize.h / 2
      const x2 = to.x + toSize.w / 2
      const y2 = to.y + toSize.h / 2

      const dx = x2 - x1
      const dy = y2 - y1
      const len = Math.max(1, Math.sqrt(dx * dx + dy * dy))

      const nx = -dy / len
      const ny = dx / len

      const offset = isMobile.value
        ? ((index % 3) - 1) * 18
        : (index % 2 === 0 ? -20 : 20)

      const c1x = x1 + dx * 0.35 + nx * offset
      const c1y = y1 + dy * 0.35 + ny * offset
      const c2x = x1 + dx * 0.65 + nx * offset
      const c2y = y1 + dy * 0.65 + ny * offset

      const path = 'M ' + x1 + ' ' + y1 +
        ' C ' + c1x + ' ' + c1y +
        ', ' + c2x + ' ' + c2y +
        ', ' + x2 + ' ' + y2

      let t = 0.5

      if (isMobile.value) {
        if (from.kind === 'core') t = 0.72
        else if (to.kind === 'core') t = 0.28
        else t = index % 2 === 0 ? 0.42 : 0.58
      }

      let labelX = x1 + dx * t + nx * (offset + (index % 2 === 0 ? 24 : -24))
      let labelY = y1 + dy * t + ny * (offset + (index % 2 === 0 ? 18 : -18))

      if (isMobile.value) {
        const core = nodes.find((node) => node.kind === 'core')

        if (core) {
          const coreLeft = core.x - 34
          const coreRight = core.x + mobileCoreW + 34
          const coreTop = core.y - 28
          const coreBottom = core.y + mobileCoreH + 28

          const nearCore =
            labelX >= coreLeft &&
            labelX <= coreRight &&
            labelY >= coreTop &&
            labelY <= coreBottom

          if (nearCore) {
            const coreCx = core.x + mobileCoreW / 2
            const coreCy = core.y + mobileCoreH / 2

            labelX += labelX < coreCx ? -46 : 46
            labelY += labelY < coreCy ? -30 : 30
          }
        }

        labelX = clamp(labelX, 30, board.value.width - 30)
        labelY = clamp(labelY, 20, board.value.height - 20)
      }

      const label = rel.relation_name || rel.relationName || rel.relation_type || rel.relationType || '关系'

      return {
        id: rel.id || index + 1,
        path,
        pathD: path,
        d: path,
        curve: path,
        label,
        text: label,
        relation_name: label,
        relationName: label,
        labelX,
        labelY,
        midX: labelX,
        midY: labelY,
        x: labelX,
        y: labelY,
        relation_type: rel.relation_type || rel.relationType || '',
        relationType: rel.relationType || rel.relation_type || '',
        className: rel.relation_type || rel.relationType || '',
        class: rel.relation_type || rel.relationType || '',
        type: rel.relation_type || rel.relationType || ''
      }
    })

  return items.filter((item): item is LineItem => item !== null)
})

onMounted(() => {
  updateMobileState()
  window.addEventListener('resize', updateMobileState)
  load()
})

async function load() {
  loading.value = true

  try {
    applyTreePayload(await api.publicTree())
  } catch {
    tree.nodes = []
    tree.relationships = []
  } finally {
    loading.value = false
  }
}

function applyTreePayload(payload: PublicTreeResult) {
  tree.config = {
    title: payload.config?.title || tree.config.title,
    subtitle: payload.config?.subtitle || tree.config.subtitle,
    description: payload.config?.description || tree.config.description
  }

  tree.nodes = Array.isArray(payload.nodes) ? payload.nodes : []
  tree.relationships = Array.isArray(payload.relationships) ? payload.relationships : []
  tree.generation_keys = Array.isArray(payload.generation_keys) ? payload.generation_keys : []
  tree.generations = payload.generations || {}
}


function buildMobilePositionedNodes() {
  const core = coreNode.value || tree.nodes[0]

  const mobileCardW = 78
  const mobileCardH = 44
  const mobileCoreW = 116
  const mobileCoreH = 64

  const others = tree.nodes
    .filter((node) => node.id !== core.id)
    .sort((a, b) => {
      const kindA = getNodeKind(a, core)
      const kindB = getNodeKind(b, core)
      const order = ['elder', 'spouse', 'relative', 'child']
      const diff = order.indexOf(kindA) - order.indexOf(kindB)

      if (diff !== 0) return diff
      return Number(a.sort || 0) - Number(b.sort || 0)
    })

  const result = []

  const w = board.value.width
  const h = board.value.height
  const cx = w / 2
  const cy = h / 2

  result.push({
    ...core,
    x: cx - mobileCoreW / 2,
    y: cy - mobileCoreH / 2,
    kind: 'core'
  })

  const leftX = 12
  const rightX = w - mobileCardW - 12
  const midX = cx - mobileCardW / 2

  const slots = [
    { x: midX, y: 18 },
    { x: leftX, y: Math.round(h * 0.19) },
    { x: rightX, y: Math.round(h * 0.19) },
    { x: leftX, y: Math.round(h * 0.40) },
    { x: rightX, y: Math.round(h * 0.40) },
    { x: leftX, y: Math.round(h * 0.64) },
    { x: rightX, y: Math.round(h * 0.64) },
    { x: midX, y: h - mobileCardH - 18 },
    { x: Math.round(w * 0.27 - mobileCardW / 2), y: Math.round(h * 0.10) },
    { x: Math.round(w * 0.73 - mobileCardW / 2), y: Math.round(h * 0.10) },
    { x: Math.round(w * 0.27 - mobileCardW / 2), y: Math.round(h * 0.82) },
    { x: Math.round(w * 0.73 - mobileCardW / 2), y: Math.round(h * 0.82) }
  ]

  others.forEach((node, index) => {
    const slot = slots[index]
    if (!slot) return

    result.push({
      ...node,
      x: Math.max(6, Math.min(w - mobileCardW - 6, slot.x)),
      y: Math.max(8, Math.min(h - mobileCardH - 8, slot.y)),
      kind: getNodeKind(node, core)
    })
  })

  return result
}

function placeHorizontal(
  result: PositionedNode[],
  nodes: PublicTreeNode[],
  kind: string,
  y: number
) {
  if (!nodes.length) return

  const left = 80
  const right = board.value.width - cardW - 80
  const available = right - left

  nodes.forEach((node, index) => {
    const x = nodes.length === 1
      ? center.value.x - cardW / 2
      : left + (available / Math.max(nodes.length - 1, 1)) * index

    result.push({
      ...node,
      x,
      y,
      kind
    })
  })
}

function placeVertical(
  result: PositionedNode[],
  items: Array<{ node: PublicTreeNode; kind: string }>,
  side: 'left' | 'right'
) {
  if (!items.length) return

  const x = side === 'left' ? 82 : board.value.width - cardW - 82
  const gap = 148
  const totalHeight = items.length * cardH + Math.max(0, items.length - 1) * (gap - cardH)
  const startY = center.value.y - totalHeight / 2

  items.forEach((item, index) => {
    result.push({
      ...item.node,
      x,
      y: startY + index * gap,
      kind: item.kind
    })
  })
}

function findCoreNode(nodes: PublicTreeNode[]) {
  if (!nodes.length) return null

  const exact = nodes.find((node) => node.name.includes('汉武帝') || node.name.includes('刘彻'))
  if (exact) return exact

  const roleCore = nodes.find((node) => {
    return node.role.includes('核心') || node.role.includes('中心')
  })

  if (roleCore) return roleCore

  const sorted = nodes
    .map((node) => {
      const count = tree.relationships.filter((rel) => {
        return rel.from_node_id === node.id || rel.to_node_id === node.id
      }).length

      return { node, count }
    })
    .sort((a, b) => b.count - a.count)

  return sorted[0]?.node || nodes[0]
}

function getNodeKind(node: PublicTreeNode, core: PublicTreeNode) {
  const directRelations = tree.relationships.filter((rel) => {
    return rel.from_node_id === node.id || rel.to_node_id === node.id
  })

  if (directRelations.some((rel) => rel.relation_type === 'spouse')) return 'spouse'
  if (node.generation < core.generation) return 'elder'
  if (node.generation > core.generation) return 'child'

  return 'relative'
}

function edgePoint(from: PositionedNode, to: PositionedNode) {
  const fromW = from.kind === 'core' ? coreW : cardW
  const fromH = from.kind === 'core' ? coreH : cardH
  const toW = to.kind === 'core' ? coreW : cardW
  const toH = to.kind === 'core' ? coreH : cardH

  const fromCenterX = from.x + fromW / 2
  const fromCenterY = from.y + fromH / 2
  const toCenterX = to.x + toW / 2
  const toCenterY = to.y + toH / 2

  const dx = toCenterX - fromCenterX
  const dy = toCenterY - fromCenterY

  if (Math.abs(dx) > Math.abs(dy)) {
    return {
      x: dx > 0 ? from.x + fromW : from.x,
      y: fromCenterY
    }
  }

  return {
    x: fromCenterX,
    y: dy > 0 ? from.y + fromH : from.y
  }
}

function relatedText(nodeId: number) {
  return tree.relationships
    .filter((rel) => rel.from_node_id === nodeId || rel.to_node_id === nodeId)
    .slice(0, 1)
    .map((rel) => {
      const otherId = rel.from_node_id === nodeId ? rel.to_node_id : rel.from_node_id
      return (rel.relation_name || rel.relation_type) + '：' + nodeName(otherId)
    })
}

function nodeName(id: number) {
  return tree.nodes.find((item) => item.id === id)?.name || '-'
}

function firstChar(name: string) {
  return name ? name.slice(0, 1) : '?'
}

function safeClass(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]/g, '')
}
</script>

<style scoped>
.family-tree-page {
  min-height: 100vh;
  background:
    radial-gradient(circle at 12% 8%, rgba(185, 28, 28, 0.08), transparent 28%),
    linear-gradient(180deg, #fff7ed 0%, #f8efe3 100%);
}

.intro-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 260px;
  gap: 24px;
  align-items: center;
  padding: 34px;
  border-radius: 28px;
  background:
    radial-gradient(circle at top right, rgba(251, 191, 36, 0.16), transparent 34%),
    linear-gradient(135deg, #fffaf3, #fff4e5);
  border: 1px solid #ead8c0;
  box-shadow: 0 24px 70px rgba(80, 45, 20, 0.08);
}

.eyebrow {
  margin: 0 0 10px;
  color: #b91c1c;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.intro-card h1 {
  margin: 0;
  color: #3f2418;
  font-size: clamp(30px, 4vw, 48px);
}

.intro-card p {
  margin: 14px 0 0;
  color: #7c4a2a;
  line-height: 1.85;
}

.intro-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.intro-stats div {
  padding: 18px;
  border-radius: 20px;
  background: #7f1d1d;
  color: #fde68a;
  text-align: center;
}

.intro-stats strong {
  display: block;
  font-size: 36px;
  line-height: 1;
}

.intro-stats span {
  display: block;
  margin-top: 8px;
  font-weight: 800;
  font-size: 14px;
}

.empty {
  margin-top: 24px;
  padding: 42px;
  border-radius: 24px;
  background: #fffaf3;
  border: 1px solid #ead8c0;
  color: #7c4a2a;
  text-align: center;
  font-weight: 800;
}

.tree-shell {
  margin-top: 24px;
  display: grid;
  gap: 22px;
}

.tree-header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: center;
  padding: 24px 28px;
  border-radius: 24px;
  background: #fffaf3;
  border: 1px solid #ead8c0;
  box-shadow: 0 18px 56px rgba(80, 45, 20, 0.06);
}

.tree-header h2 {
  margin: 0;
  color: #3f2418;
}

.tree-header p {
  margin: 8px 0 0;
  color: #7c4a2a;
  line-height: 1.7;
}

.legend {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.legend span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 8px 12px;
  border-radius: 999px;
  background: #fff4e5;
  color: #7c4a2a;
  font-weight: 800;
  font-size: 13px;
}

.line-dot {
  width: 20px;
  height: 3px;
  border-radius: 999px;
  background: #b45309;
}

.line-dot.spouse {
  background: #be123c;
}

.line-dot.relative {
  background: #2563eb;
}

.tree-scroll {
  overflow-x: auto;
  padding: 14px;
  border-radius: 28px;
  background:
    linear-gradient(90deg, rgba(127, 29, 29, 0.04) 1px, transparent 1px),
    linear-gradient(180deg, rgba(127, 29, 29, 0.04) 1px, transparent 1px),
    #f8efe3;
  background-size: 32px 32px;
  border: 1px solid #ead8c0;
}

.tree-board {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  background:
    radial-gradient(circle at 50% 50%, rgba(251, 191, 36, 0.16), transparent 20%),
    linear-gradient(135deg, #fffdf7, #fff4e5);
  border: 1px solid #ead8c0;
}

.center-halo {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 390px;
  height: 390px;
  transform: translate(-50%, -50%);
  border-radius: 50%;
  border: 1px dashed rgba(127, 29, 29, 0.16);
  background: radial-gradient(circle, rgba(185, 28, 28, 0.05), transparent 65%);
  z-index: 0;
}

.relation-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
}

.relation-line {
  stroke: #b45309;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-dasharray: 6 6;
  opacity: 0.76;
}

.relation-line.spouse {
  stroke: #be123c;
  stroke-width: 2.1;
}

.relation-line.in_law,
.relation-line.uncle_nephew {
  stroke: #2563eb;
}

.line-label {
  height: 24px;
  line-height: 24px;
  text-align: center;
  border-radius: 999px;
  background: rgba(127, 29, 29, 0.94);
  color: #fde68a;
  font-size: 11px;
  font-weight: 900;
  box-shadow: 0 6px 14px rgba(127, 29, 29, 0.14);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.person-card {
  position: absolute;
  z-index: 2;
  width: 232px;
  min-height: 112px;
  display: grid;
  grid-template-columns: 44px 1fr;
  gap: 11px;
  padding: 13px;
  border-radius: 18px;
  background: rgba(255, 250, 243, 0.98);
  border: 1px solid #ead8c0;
  box-shadow: 0 12px 28px rgba(80, 45, 20, 0.08);
}

.person-card.core {
  width: 285px;
  min-height: 132px;
  grid-template-columns: 54px 1fr;
  padding: 18px;
  background:
    radial-gradient(circle at top right, rgba(251, 191, 36, 0.2), transparent 32%),
    linear-gradient(135deg, #7f1d1d, #b91c1c);
  color: #fff7ed;
  border-color: rgba(253, 230, 138, 0.65);
  box-shadow:
    0 24px 60px rgba(127, 29, 29, 0.2),
    0 0 0 8px rgba(185, 28, 28, 0.05);
}

.person-card.elder {
  border-top: 4px solid #b45309;
}

.person-card.spouse {
  border-top: 4px solid #be123c;
}

.person-card.child {
  border-top: 4px solid #15803d;
}

.person-card.relative {
  border-top: 4px solid #2563eb;
}

.avatar {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #b91c1c, #7f1d1d);
  color: #fde68a;
  font-size: 20px;
  font-weight: 900;
}

.core .avatar {
  width: 50px;
  height: 50px;
  border-radius: 17px;
  background: #fde68a;
  color: #7f1d1d;
  font-size: 24px;
}

.person-info h3 {
  margin: 0 0 4px;
  color: #3f2418;
  font-size: 18px;
  line-height: 1.2;
}

.core .person-info h3 {
  color: #fff7ed;
  font-size: 22px;
}

.person-info p {
  margin: 4px 0;
}

.role {
  color: #7f1d1d;
  font-weight: 900;
  font-size: 13px;
}

.core .role {
  color: #fde68a;
  font-size: 14px;
}

.desc {
  color: #7c4a2a;
  line-height: 1.55;
  font-size: 13px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.core .desc {
  color: #ffedd5;
  font-size: 13px;
  -webkit-line-clamp: 2;
}

.relation-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-top: 7px;
}

.relation-tags span {
  padding: 3px 7px;
  border-radius: 999px;
  background: #fff4e5;
  color: #7c4a2a;
  font-size: 11px;
  font-weight: 800;
  max-width: 150px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.core .relation-tags span {
  background: rgba(253, 230, 138, 0.16);
  color: #fde68a;
}

.relationship-summary {
  padding: 26px;
  border-radius: 24px;
  background: #fffaf3;
  border: 1px solid #ead8c0;
  box-shadow: 0 18px 56px rgba(80, 45, 20, 0.06);
}

.relationship-summary h2 {
  margin: 0 0 18px;
  color: #3f2418;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
}

.summary-card {
  padding: 16px;
  border-radius: 18px;
  background: #fff7ed;
  border: 1px solid #ead8c0;
}

.summary-title {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.summary-title strong {
  color: #3f2418;
}

.summary-title span {
  padding: 4px 9px;
  border-radius: 999px;
  background: #7f1d1d;
  color: #fde68a;
  font-size: 12px;
  font-weight: 900;
}

.summary-card p {
  margin: 10px 0 0;
  color: #7c4a2a;
  line-height: 1.7;
}

.privacy-note {
  padding: 18px 22px;
  border-radius: 20px;
  background: #fff4e5;
  border: 1px solid #ead8c0;
  color: #7c4a2a;
  line-height: 1.8;
}

.privacy-note strong {
  color: #7f1d1d;
}

@media (max-width: 900px) {
  .intro-card,
  .tree-header {
    grid-template-columns: 1fr;
    display: grid;
  }

  .intro-stats,
  .summary-grid {
    grid-template-columns: 1fr;
  }

  .legend {
    justify-content: flex-start;
  }
}













/* H5 safe relation network */
@media (max-width: 768px) {
  .family-tree-page {
    min-height: auto !important;
    padding-bottom: 8px !important;
  }

  .family-tree-page .container {
    width: 100% !important;
    max-width: 100% !important;
    padding: 0 10px !important;
  }

  .intro-card {
    grid-template-columns: 1fr auto !important;
    gap: 8px !important;
    padding: 12px !important;
    border-radius: 16px !important;
    margin-bottom: 8px !important;
  }

  .intro-card h1 {
    font-size: 20px !important;
    line-height: 1.22 !important;
  }

  .intro-card p {
    font-size: 11px !important;
    line-height: 1.45 !important;
  }

  .intro-stats {
    display: flex !important;
    gap: 5px !important;
  }

  .intro-stats div {
    width: 48px !important;
    min-width: 48px !important;
    padding: 7px 3px !important;
    border-radius: 11px !important;
  }

  .intro-stats strong {
    font-size: 19px !important;
  }

  .intro-stats span {
    font-size: 9px !important;
  }

  .tree-shell {
    margin-top: 8px !important;
    gap: 7px !important;
  }

  .tree-header {
    padding: 9px 11px !important;
    border-radius: 15px !important;
  }

  .tree-header h2 {
    font-size: 14px !important;
  }

  .tree-header p,
  .legend,
  .relationship-summary,
  .privacy-note {
    display: none !important;
  }

  .tree-scroll {
    overflow: hidden !important;
    padding: 5px !important;
    border-radius: 18px !important;
  }

  .tree-board {
    position: relative !important;
    display: block !important;
    width: 100% !important;
    height: clamp(470px, calc(100dvh - 215px), 620px) !important;
    min-height: 470px !important;
    max-height: 620px !important;
    overflow: hidden !important;
    padding: 0 !important;
    border-radius: 18px !important;
    background:
      radial-gradient(circle at 50% 50%, rgba(251, 191, 36, 0.13), transparent 28%),
      linear-gradient(90deg, rgba(127, 29, 29, 0.035) 1px, transparent 1px),
      linear-gradient(180deg, rgba(127, 29, 29, 0.035) 1px, transparent 1px),
      #fffaf3 !important;
    background-size: auto, 22px 22px, 22px 22px, auto !important;
  }

  .relation-layer {
    display: block !important;
    position: absolute !important;
    inset: 0 !important;
    width: 100% !important;
    height: 100% !important;
    z-index: 1 !important;
    pointer-events: none !important;
    overflow: visible !important;
  }

  .center-halo {
    display: block !important;
    width: 150px !important;
    height: 150px !important;
    opacity: 0.45 !important;
    z-index: 0 !important;
  }

  .relation-line {
    display: block !important;
    stroke-width: 1px !important;
    stroke-dasharray: 4 5 !important;
    opacity: 0.5 !important;
  }

  .relation-line.spouse {
    stroke-width: 1.2px !important;
    opacity: 0.66 !important;
  }

  .line-label,
  .line-label * {
    display: block !important;
    font-size: 7px !important;
    font-weight: 900 !important;
    line-height: 1 !important;
    color: #7f1d1d !important;
    fill: #7f1d1d !important;
    stroke: #fffaf3 !important;
    stroke-width: 3px !important;
    paint-order: stroke !important;
    background: transparent !important;
    border: 0 !important;
    border-radius: 0 !important;
    box-shadow: none !important;
    padding: 0 !important;
    min-width: 0 !important;
    width: auto !important;
    height: auto !important;
    white-space: nowrap !important;
    pointer-events: none !important;
  }

  .person-card {
    position: absolute !important;
    transform: none !important;
    width: 78px !important;
    min-width: 78px !important;
    max-width: 78px !important;
    min-height: 44px !important;
    padding: 5px !important;
    border-radius: 10px !important;
    display: grid !important;
    grid-template-columns: 19px minmax(0, 1fr) !important;
    gap: 4px !important;
    align-items: center !important;
    box-shadow: 0 7px 16px rgba(80, 45, 20, 0.08) !important;
    z-index: 2 !important;
  }

  .person-card.core {
    width: 116px !important;
    min-width: 116px !important;
    max-width: 116px !important;
    min-height: 64px !important;
    padding: 7px !important;
    border-radius: 13px !important;
    grid-template-columns: 28px minmax(0, 1fr) !important;
    gap: 6px !important;
    z-index: 3 !important;
    box-shadow: 0 12px 26px rgba(127, 29, 29, 0.18) !important;
  }

  .avatar {
    width: 19px !important;
    height: 19px !important;
    border-radius: 6px !important;
    font-size: 10px !important;
  }

  .core .avatar {
    width: 28px !important;
    height: 28px !important;
    border-radius: 9px !important;
    font-size: 15px !important;
  }

  .person-info {
    min-width: 0 !important;
  }

  .person-info h3 {
    max-width: 48px !important;
    margin: 0 0 2px !important;
    font-size: 9px !important;
    line-height: 1.12 !important;
    white-space: nowrap !important;
    overflow: hidden !important;
    text-overflow: ellipsis !important;
  }

  .core .person-info h3 {
    max-width: 70px !important;
    font-size: 11px !important;
  }

  .role {
    max-width: 48px !important;
    margin: 0 !important;
    font-size: 7px !important;
    line-height: 1.15 !important;
    white-space: nowrap !important;
    overflow: hidden !important;
    text-overflow: ellipsis !important;
  }

  .core .role {
    max-width: 70px !important;
    font-size: 8px !important;
  }

  .desc,
  .relation-tags {
    display: none !important;
  }
}

@media (max-width: 380px) {
  .tree-board {
    height: clamp(450px, calc(100dvh - 208px), 590px) !important;
    min-height: 450px !important;
  }

  .person-card {
    width: 74px !important;
    min-width: 74px !important;
    max-width: 74px !important;
    min-height: 42px !important;
    padding: 4px !important;
  }

  .person-card.core {
    width: 110px !important;
    min-width: 110px !important;
    max-width: 110px !important;
    min-height: 60px !important;
  }

  .person-info h3,
  .role {
    max-width: 44px !important;
  }

  .core .person-info h3,
  .core .role {
    max-width: 64px !important;
  }

  .line-label,
  .line-label * {
    font-size: 6.5px !important;
  }
}


/* relation labels on path */
.relation-layer .line-label-on-path {
  display: block !important;
  font-size: 11px;
  font-weight: 900;
  fill: #7f1d1d;
  stroke: #fffaf3;
  stroke-width: 4px;
  paint-order: stroke;
  letter-spacing: 0.02em;
  pointer-events: none;
}

.relation-layer .line-label-on-path textPath {
  fill: #7f1d1d;
}

@media (max-width: 768px) {
  .relation-layer .line-label-on-path {
    display: block !important;
    font-size: 7px !important;
    stroke-width: 3px !important;
  }

  .relation-layer .line-label {
    display: block !important;
  }
}

</style>
