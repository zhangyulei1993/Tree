import { layoutFamilyTree } from './layoutFamilyTree'
import { getExportLineageDisplay, type ExportLineageScope } from './exportLineagePresentation'
import type { ExportLineageTree } from './exportLineageTree'
import { resolveAvatarAgeBand } from './avatarAgeBand'
import {
  buildPortraitAppearanceMap,
  resolvePortraitAvatarStyle,
  resolveRegionalPortraitAsset,
  resolveRegionalPortraitGlassesAsset,
  resolveSelectedPortraitAsset
} from './portraitCatalog'
import { buildSpouseRelationLabelMap } from './spouseRelationLabel'
import type { FamilyTreeResult, TreeNode } from '@/types/api'

export interface ExportTreePosterPage {
  width: number
  height: number
  familyName: string
  subtitle: string
  scope: ExportLineageScope
  scale: number
  layout: ReturnType<typeof layoutFamilyTree>
  displayByMemberId: Map<number, ReturnType<typeof getExportLineageDisplay>>
  avatarUrlByMemberId: Map<number, string>
  relationBadgeByMemberId: Record<number, string>
}

const PAGE_WIDTH = 1600
const PAGE_MARGIN = 72
const HEADER_HEIGHT = 180
const FOOTER_HEIGHT = 78
const MIN_SCALE = 0.58

function flattenLineage(lineage: ExportLineageTree) {
  return lineage.generations.flatMap((generation) => generation.members)
}

function buildPosterDimensions(layout: ReturnType<typeof layoutFamilyTree>) {
  const rawWidth = Math.max(layout.canvasWidth, 320)
  const scale = Math.max(MIN_SCALE, Math.min(1, (PAGE_WIDTH - PAGE_MARGIN * 2) / rawWidth))
  return {
    scale,
    width: Math.ceil(rawWidth * scale + PAGE_MARGIN * 2),
    height: Math.ceil(Math.max(layout.canvasHeight, 260) * scale + HEADER_HEIGHT + FOOTER_HEIGHT)
  }
}

function deceasedYearText(node: TreeNode): string {
  if (node.isLiving !== false) return ''
  const birth = node.birthDate?.slice(0, 4)
  const death = node.deathDate?.slice(0, 4)
  if (!death) return '已故'
  return birth ? `${birth}-${death}` : `卒：${death}`
}

/** Mirrors the live tree's portrait choice, while keeping export rendering independent from Vue. */
function resolveExportAvatarUrl(node: TreeNode, appearanceVariant?: number): string {
  const gender = node.gender === 'MALE' || node.gender === 'FEMALE' ? node.gender.toLowerCase() as 'male' | 'female' : null
  if (!gender) return '/static/tree/avatars/avatar-neutral.png'

  const ageBand = resolveAvatarAgeBand(node.birthDate, node.isLiving === false ? node.deathDate : undefined)
  const selected = resolveSelectedPortraitAsset({ gender, ageBand, style: node.portraitAvatarStyle })
  if (selected) return selected

  const variant = appearanceVariant
  if (node.portraitGlassesStyle === 'ON') {
    return resolveRegionalPortraitGlassesAsset({ theme: 'north', gender, ageBand, memberId: node.memberId, variant })
  }
  return resolveRegionalPortraitAsset({ theme: 'north', gender, ageBand, memberId: node.memberId, variant })
}

/** Builds a wide, readable tree poster instead of reducing a live tree screenshot. */
export function buildExportTreePosterPage(
  tree: FamilyTreeResult,
  lineage: ExportLineageTree,
  familyName: string,
  subtitle: string,
  scope: ExportLineageScope
): ExportTreePosterPage {
  const lineageNodes = flattenLineage(lineage)
  const nodeIds = new Set(lineageNodes.map((item) => item.member.memberId))
  const visibleTree: FamilyTreeResult = {
    ...tree,
    nodes: tree.nodes.filter((node) => nodeIds.has(node.memberId)),
    edges: tree.edges.filter((edge) =>
      edge.relationshipType === 'PARENT_CHILD' &&
      nodeIds.has(edge.fromMemberId) &&
      nodeIds.has(edge.toMemberId)
    ),
    tree: tree.tree.filter((item) => nodeIds.has(item.memberId))
  }
  const layout = layoutFamilyTree({ ...visibleTree, viewerMemberId: null })
  const { scale, width, height } = buildPosterDimensions(layout)
  const displayByMemberId = new Map(
    lineageNodes.map((item) => [item.member.memberId, getExportLineageDisplay(item, scope)])
  )
  const appearanceByMemberId = buildPortraitAppearanceMap(visibleTree.nodes)
  const avatarUrlByMemberId = new Map(
    visibleTree.nodes.map((node) => [node.memberId, resolveExportAvatarUrl(node, appearanceByMemberId[node.memberId])])
  )

  return {
    width, height, familyName, subtitle, scope, scale, layout, displayByMemberId, avatarUrlByMemberId,
    relationBadgeByMemberId: buildSpouseRelationLabelMap(visibleTree)
  }
}

/** Builds a full-family tree image without applying long-scroll member or privacy filters. */
export function buildFamilyTreePosterPage(
  tree: FamilyTreeResult,
  familyName: string,
  subtitle = '完整家庭树 · 内部留存'
): ExportTreePosterPage {
  const visibleTree = tree
  const layout = layoutFamilyTree({ ...visibleTree, viewerMemberId: null })
  const { scale, width, height } = buildPosterDimensions(layout)
  const displayByMemberId = new Map(
    visibleTree.nodes.map((node) => [node.memberId, {
      name: node.displayName || '未命名',
      yearText: deceasedYearText(node),
      parentLabels: [],
      spouseLabels: []
    }])
  )
  const appearanceByMemberId = buildPortraitAppearanceMap(visibleTree.nodes)
  const avatarUrlByMemberId = new Map(
    visibleTree.nodes.map((node) => [node.memberId, resolveExportAvatarUrl(node, appearanceByMemberId[node.memberId])])
  )

  return {
    width, height, familyName, subtitle, scope: 'INTERNAL', scale, layout, displayByMemberId, avatarUrlByMemberId,
    relationBadgeByMemberId: buildSpouseRelationLabelMap(visibleTree)
  }
}

/** Canvas exports need local image paths; remote image URLs are omitted by WeChat canvas. */
export async function resolveExportTreePortraitPaths(urlByMemberId: ReadonlyMap<number, string>) {
  const sourcePaths = new Map<string, string>()
  const sources = [...new Set(urlByMemberId.values())]
  const concurrency = 4
  let nextIndex = 0

  async function loadSource(src: string) {
    try {
      const image = await new Promise<UniApp.GetImageInfoSuccessData>((resolve, reject) => {
        uni.getImageInfo({ src, success: resolve, fail: reject })
      })
      sourcePaths.set(src, image.path)
    } catch {
      // A missing portrait must not prevent the rest of the family tree from exporting.
    }
  }

  async function worker() {
    while (nextIndex < sources.length) {
      const src = sources[nextIndex]
      nextIndex += 1
      await loadSource(src)
    }
  }

  await Promise.all(Array.from({ length: Math.min(concurrency, sources.length) }, () => worker()))
  return new Map(
    [...urlByMemberId.entries()]
      .map(([memberId, src]) => {
        const path = sourcePaths.get(src)
        return path ? [memberId, path] as const : null
      })
      .filter((entry): entry is readonly [number, string] => Boolean(entry))
  )
}

function fitText(ctx: UniApp.CanvasContext, text: string, maxWidth: number): string {
  if (ctx.measureText(text).width <= maxWidth) return text
  let result = text
  while (result.length > 1 && ctx.measureText(`${result}…`).width > maxWidth) result = result.slice(0, -1)
  return `${result}…`
}

/** A death marker must not depend on whether the optional death date was recorded. */
export function isExportTreeMemberDeceased(node: Pick<TreeNode, 'isLiving'>): boolean {
  return node.isLiving === false
}

function drawMember(
  ctx: UniApp.CanvasContext,
  node: TreeNode,
  x: number,
  y: number,
  width: number,
  height: number,
  display: ReturnType<typeof getExportLineageDisplay> | undefined,
  avatarPath?: string,
  relationBadge?: string
) {
  const female = node.gender === 'FEMALE'
  const deceased = isExportTreeMemberDeceased(node)
  ctx.setFillStyle(deceased ? '#f1eee7' : '#fffaf0')
  ctx.fillRect(x, y, width, height)
  ctx.setStrokeStyle(deceased ? '#8a8d91' : female ? '#c7786d' : '#4b6683')
  ctx.setLineWidth(2)
  ctx.strokeRect(x, y, width, height)
  const portraitSize = Math.min(50, Math.max(28, width * 0.42), Math.max(28, height * 0.48))
  const portraitX = x + (width - portraitSize) / 2
  const portraitY = y + 7
  ctx.setFillStyle(female ? '#f3dfd3' : '#dbe3e7')
  ctx.beginPath()
  ctx.arc(portraitX + portraitSize / 2, portraitY + portraitSize / 2, portraitSize / 2, 0, Math.PI * 2)
  ctx.fill()
  if (avatarPath) {
    ctx.save()
    ctx.beginPath()
    ctx.arc(portraitX + portraitSize / 2, portraitY + portraitSize / 2, portraitSize / 2, 0, Math.PI * 2)
    ctx.clip()
    ctx.drawImage(avatarPath, portraitX, portraitY, portraitSize, portraitSize)
    if (deceased) {
      ctx.setFillStyle('rgba(232, 231, 225, 0.58)')
      ctx.fillRect(portraitX, portraitY, portraitSize, portraitSize)
    }
    ctx.restore()
  }
  ctx.setStrokeStyle(deceased ? '#8a8d91' : female ? '#c7786d' : '#4b6683')
  ctx.setLineWidth(1)
  ctx.beginPath()
  ctx.arc(portraitX + portraitSize / 2, portraitY + portraitSize / 2, portraitSize / 2, 0, Math.PI * 2)
  ctx.stroke()
  if (relationBadge) {
    ctx.setFillStyle('#fff7e8')
    ctx.fillRect(x + 5, y + 5, 34, 17)
    ctx.setStrokeStyle('#b85a4c')
    ctx.setLineWidth(1)
    ctx.strokeRect(x + 5, y + 5, 34, 17)
    ctx.setFillStyle('#a45045')
    ctx.setTextAlign('center')
    ctx.setFontSize(10)
    ctx.fillText(relationBadge, x + 22, y + 17)
  }
  ctx.setFillStyle(deceased ? '#656b70' : female ? '#a45045' : '#193856')
  ctx.setTextAlign('center')
  ctx.setFontSize(22)
  const nameY = Math.min(y + height - 38, portraitY + portraitSize + 23)
  ctx.fillText(fitText(ctx, display?.name || node.displayName || '未命名', width - 18), x + width / 2, nameY)
  if (deceased) {
    ctx.setFillStyle('#6d625b')
    ctx.setFontSize(10)
    ctx.setTextAlign('right')
    ctx.fillText('已故', x + width - 7, y + 15)
    ctx.setTextAlign('center')
  }
  if (display?.yearText) {
    ctx.setFillStyle('#718096')
    ctx.setFontSize(13)
    ctx.fillText(fitText(ctx, display.yearText, width - 18), x + width / 2, Math.min(y + height - 22, nameY + 20))
  }
  if (display?.spouseLabels?.length) {
    ctx.setFillStyle('#855f4e')
    ctx.setFontSize(12)
    ctx.fillText(fitText(ctx, `（${display.spouseLabels.join('、')}）`, width - 18), x + width / 2, y + height - 14)
  }
}

export function drawExportTreePoster(
  ctx: UniApp.CanvasContext,
  page: ExportTreePosterPage,
  avatarPathByMemberId: ReadonlyMap<number, string> = new Map()
) {
  ctx.setFillStyle('#f8f0df')
  ctx.fillRect(0, 0, page.width, page.height)
  ctx.setStrokeStyle('#b88a52')
  ctx.setLineWidth(2)
  ctx.strokeRect(18, 18, page.width - 36, page.height - 36)
  ctx.setTextAlign('center')
  ctx.setFillStyle('#9e4c3b')
  ctx.setFontSize(16)
  ctx.fillText('FAMILY TREE', page.width / 2, 58)
  ctx.setFillStyle('#193856')
  ctx.setFontSize(34)
  ctx.fillText(fitText(ctx, page.familyName, page.width - PAGE_MARGIN * 2), page.width / 2, 104)
  ctx.setFillStyle('#718096')
  ctx.setFontSize(16)
  ctx.fillText(fitText(ctx, page.subtitle, page.width - PAGE_MARGIN * 2), page.width / 2, 135)
  ctx.setStrokeStyle('#c9a56b')
  ctx.setLineWidth(1)
  ctx.beginPath()
  ctx.moveTo(PAGE_MARGIN, 154)
  ctx.lineTo(page.width - PAGE_MARGIN, 154)
  ctx.stroke()

  const offsetY = HEADER_HEIGHT
  const scale = page.scale
  for (const link of page.layout.renderLinks) {
    ctx.setStrokeStyle(link.highlighted ? '#a45045' : '#a78b68')
    ctx.setLineWidth(link.highlighted ? 3 : 2)
    for (const segment of link.segments) {
      ctx.beginPath()
      if (segment.orientation === 'h') {
        ctx.moveTo(PAGE_MARGIN + segment.x * scale, offsetY + segment.y * scale)
        ctx.lineTo(PAGE_MARGIN + (segment.x + segment.width) * scale, offsetY + segment.y * scale)
      } else {
        ctx.moveTo(PAGE_MARGIN + segment.x * scale, offsetY + segment.y * scale)
        ctx.lineTo(PAGE_MARGIN + segment.x * scale, offsetY + (segment.y + segment.height) * scale)
      }
      ctx.stroke()
    }
  }

  for (const group of page.layout.renderNodes) {
    const count = Math.max(group.parents.length, 1)
    const gap = 10 * scale
    const memberWidth = (group.width * scale - gap * (count - 1)) / count
    group.parents.forEach((member, index) => {
      drawMember(
        ctx,
        member,
        PAGE_MARGIN + group.x * scale + index * (memberWidth + gap),
        offsetY + group.y * scale,
        memberWidth,
        group.height * scale,
        page.displayByMemberId.get(member.memberId),
        avatarPathByMemberId.get(member.memberId),
        page.relationBadgeByMemberId[member.memberId]
      )
    })
  }

  ctx.setStrokeStyle('#c9a56b')
  ctx.beginPath()
  ctx.moveTo(PAGE_MARGIN, page.height - 52)
  ctx.lineTo(page.width - PAGE_MARGIN, page.height - 52)
  ctx.stroke()
  ctx.setFillStyle('#718096')
  ctx.setFontSize(14)
  ctx.setTextAlign('left')
  ctx.fillText('Tree · 家谱记录与家人协作', PAGE_MARGIN, page.height - 24)
  ctx.setTextAlign('right')
  ctx.fillText('树形家谱图', page.width - PAGE_MARGIN, page.height - 24)
}
