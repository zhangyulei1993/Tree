import { hierarchy, tree } from 'd3-hierarchy'

import { buildFamilyGraph } from './buildFamilyGraph'
import { collectDirectLineMemberIds, isParentChildLink } from './directLine'
import { buildMemberGraph } from './graph'
import type {
  CoupleHierarchyDatum,
  FamilyTreeBranchNode,
  FamilyTreeBuildInput,
  FamilyTreeLayoutResult,
  RenderCoupleNode,
  RenderLinkSegment,
  RenderTreeLink
} from './types'

/** 布局尺寸（rpx）— 与 FamilyTreeMemberChip compact 样式对齐 */
const SINGLE_CARD_W = 112
const CARD_GAP = 7
const COUPLE_ROW_H = 112
const NODE_GAP_X = 14
const LEVEL_GAP_Y = 150
const CANVAS_PADDING = 14
const LINK_STROKE = 2

function coupleWidth(parentCount: number): number {
  if (parentCount <= 0) return SINGLE_CARD_W
  return parentCount * SINGLE_CARD_W + (parentCount - 1) * CARD_GAP
}

function branchToDatum(branch: FamilyTreeBranchNode): CoupleHierarchyDatum {
  return {
    id: branch.id,
    parents: branch.parents,
    coupleWidth: coupleWidth(branch.parents.length),
    children: branch.childBranches.length > 0 ? branch.childBranches.map(branchToDatum) : undefined
  }
}

function maxCoupleWidth(datum: CoupleHierarchyDatum): number {
  let max = datum.coupleWidth
  for (const child of datum.children || []) {
    max = Math.max(max, maxCoupleWidth(child))
  }
  return max
}

function buildOrthogonalLink(
  source: RenderCoupleNode,
  target: RenderCoupleNode
): RenderLinkSegment[] {
  const x1 = source.x + source.width / 2
  const y1 = source.y + source.height
  const x2 = target.x + target.width / 2
  const y2 = target.y
  const midY = y1 + (y2 - y1) / 2

  const segments: RenderLinkSegment[] = [
    {
      x: x1 - LINK_STROKE / 2,
      y: y1,
      width: LINK_STROKE,
      height: Math.max(midY - y1, LINK_STROKE),
      orientation: 'v'
    }
  ]

  if (Math.abs(x2 - x1) > LINK_STROKE) {
    segments.push({
      x: Math.min(x1, x2),
      y: midY - LINK_STROKE / 2,
      width: Math.abs(x2 - x1),
      height: LINK_STROKE,
      orientation: 'h'
    })
  }

  segments.push({
    x: x2 - LINK_STROKE / 2,
    y: midY,
    width: LINK_STROKE,
    height: Math.max(y2 - midY, LINK_STROKE),
    orientation: 'v'
  })

  return segments
}

export function layoutFamilyTree(input: FamilyTreeBuildInput): FamilyTreeLayoutResult {
  let memberCount = 0
  let renderedCount = 0

  try {
    const safeInput: FamilyTreeBuildInput = {
      nodes: Array.isArray(input.nodes) ? input.nodes : [],
      edges: Array.isArray(input.edges) ? input.edges : [],
      tree: Array.isArray(input.tree) ? input.tree : [],
      viewerMemberId: input.viewerMemberId
    }

    const graph = buildFamilyGraph(safeInput)
    memberCount = graph.memberCount
    renderedCount = graph.renderedCount
    if (!graph.success) {
      return {
        success: false,
        renderNodes: [],
        renderLinks: [],
        canvasWidth: 0,
        canvasHeight: 0,
        memberCount,
        renderedCount,
        scrollIntoViewId: undefined,
        unlocatedMembers: [],
        message: graph.message
      }
    }

    const memberGraph = buildMemberGraph(safeInput)
    const directLine =
      safeInput.viewerMemberId != null
        ? collectDirectLineMemberIds(safeInput.viewerMemberId, memberGraph)
        : new Set<number>()

    const branchData = graph.roots.map(branchToDatum)
    const unlocatedMembers = graph.unlocatedMemberIds
      .map((memberId) => memberGraph.nodeMap.get(memberId))
      .filter((node): node is import('@/types/api').TreeNode => Boolean(node))
    if (branchData.length === 0) {
      return {
        success: false,
        renderNodes: [],
        renderLinks: [],
        canvasWidth: 0,
        canvasHeight: 0,
        memberCount,
        renderedCount,
        scrollIntoViewId: undefined,
        unlocatedMembers,
        message: '家谱结构暂时无法生成，可查看关系明细'
      }
    }

    const rootDatum: CoupleHierarchyDatum =
      branchData.length === 1
        ? branchData[0]
        : {
            id: '__virtual_root__',
            virtual: true,
            parents: [],
            coupleWidth: 0,
            children: branchData
          }

    const maxWidth = Math.max(
      ...branchData.map((branch) => maxCoupleWidth(branch)),
      SINGLE_CARD_W
    )

    const root = hierarchy(rootDatum, (datum) => datum.children)
    const layout = tree<CoupleHierarchyDatum>()
      .nodeSize([maxWidth + NODE_GAP_X, LEVEL_GAP_Y])
      .separation((a, b) => {
        if (a.parent === b.parent) {
          return (a.data.coupleWidth + b.data.coupleWidth) / (2 * maxWidth) + 0.4
        }
        return 1.15
      })

    layout(root)

    let minX = Number.POSITIVE_INFINITY
    let maxX = Number.NEGATIVE_INFINITY
    root.each((node) => {
      if (node.data.virtual) return
      const half = node.data.coupleWidth / 2
      minX = Math.min(minX, node.x - half)
      maxX = Math.max(maxX, node.x + half)
    })

    const shiftX = CANVAS_PADDING - minX
    const shiftY = CANVAS_PADDING

    const renderNodes: RenderCoupleNode[] = []
    const nodeById = new Map<string, RenderCoupleNode>()
    let scrollIntoViewId: string | undefined
    let rootAnchorAssigned = false

    root.each((node) => {
      if (node.data.virtual) return
      const width = node.data.coupleWidth
      const scrollAnchorId = rootAnchorAssigned ? undefined : 'tree-root-anchor'
      if (scrollAnchorId) {
        rootAnchorAssigned = true
        scrollIntoViewId = scrollAnchorId
      }

      const renderNode: RenderCoupleNode = {
        id: node.data.id,
        parents: node.data.parents,
        x: node.x - width / 2 + shiftX,
        y: node.y + shiftY,
        width,
        height: COUPLE_ROW_H,
        scrollAnchorId
      }
      renderNodes.push(renderNode)
      nodeById.set(renderNode.id, renderNode)
    })

    const renderLinks: RenderTreeLink[] = []
    for (const link of root.links()) {
      if (link.source.data.virtual || link.target.data.virtual) continue
      const source = nodeById.get(link.source.data.id)
      const target = nodeById.get(link.target.data.id)
      if (!source || !target) continue

      let highlighted = false
      if (directLine.size > 0) {
        for (const parent of source.parents) {
          for (const child of target.parents) {
            if (
              isParentChildLink(parent.memberId, child.memberId, memberGraph) &&
              directLine.has(parent.memberId) &&
              directLine.has(child.memberId)
            ) {
              highlighted = true
              break
            }
          }
          if (highlighted) break
        }
      }

      renderLinks.push({
        id: `${source.id}->${target.id}`,
        segments: buildOrthogonalLink(source, target),
        highlighted
      })
    }

    let canvasWidth = CANVAS_PADDING
    let canvasHeight = CANVAS_PADDING
    for (const node of renderNodes) {
      canvasWidth = Math.max(canvasWidth, node.x + node.width + CANVAS_PADDING)
      canvasHeight = Math.max(canvasHeight, node.y + node.height + CANVAS_PADDING)
    }

    return {
      success: renderNodes.length > 0,
      renderNodes,
      renderLinks,
      canvasWidth,
      canvasHeight,
      memberCount,
      renderedCount,
      scrollIntoViewId,
      unlocatedMembers,
      message: renderNodes.length === 0 ? '家谱结构暂时无法生成，可查看关系明细' : undefined
    }
  } catch {
    return {
      success: false,
      renderNodes: [],
      renderLinks: [],
      canvasWidth: 0,
      canvasHeight: 0,
      memberCount,
      renderedCount,
      scrollIntoViewId: undefined,
      unlocatedMembers: [],
      message: '家谱结构暂时无法生成，可查看关系明细'
    }
  }
}
