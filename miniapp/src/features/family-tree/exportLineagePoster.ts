import type { ExportLineageGeneration, ExportLineageNode } from './exportLineageTree'
import { getExportLineageDisplay, type ExportLineageScope } from './exportLineagePresentation'

export interface ExportLineagePosterGroup {
  generation: number
  continued: boolean
  members: ExportLineageNode[]
}

export interface ExportLineagePosterPage {
  width: number
  height: number
  pageNumber: number
  pageCount: number
  familyName: string
  subtitle: string
  scope: ExportLineageScope
  groups: ExportLineagePosterGroup[]
}

const POSTER_WIDTH = 750
const POSTER_MAX_HEIGHT = 2880
const POSTER_HEADER_HEIGHT = 210
const POSTER_FOOTER_HEIGHT = 86
const POSTER_GROUP_HEAD_HEIGHT = 52
const POSTER_CARD_HEIGHT = 132
const POSTER_ROW_GAP = 16
const POSTER_GROUP_GAP = 28
const MEMBERS_PER_GROUP_CHUNK = 10

function groupHeight(memberCount: number): number {
  return POSTER_GROUP_HEAD_HEIGHT + Math.ceil(memberCount / 2) * POSTER_CARD_HEIGHT +
    Math.max(0, Math.ceil(memberCount / 2) - 1) * POSTER_ROW_GAP + POSTER_GROUP_GAP
}

/** 将长卷拆为微信可稳定处理的多张图片，同时不拆开单个成员条目。 */
export function buildExportLineagePosterPages(
  generations: ExportLineageGeneration[],
  familyName: string,
  subtitle: string,
  scope: ExportLineageScope = 'INTERNAL'
): ExportLineagePosterPage[] {
  const groups: ExportLineagePosterGroup[] = []
  for (const generation of generations) {
    for (let start = 0; start < generation.members.length; start += MEMBERS_PER_GROUP_CHUNK) {
      groups.push({
        generation: generation.generation,
        continued: start > 0,
        members: generation.members.slice(start, start + MEMBERS_PER_GROUP_CHUNK)
      })
    }
  }

  const pages: ExportLineagePosterGroup[][] = [[]]
  let usedHeight = POSTER_HEADER_HEIGHT + POSTER_FOOTER_HEIGHT
  for (const group of groups) {
    const height = groupHeight(group.members.length)
    if (pages[pages.length - 1]!.length > 0 && usedHeight + height > POSTER_MAX_HEIGHT) {
      pages.push([])
      usedHeight = POSTER_HEADER_HEIGHT + POSTER_FOOTER_HEIGHT
    }
    pages[pages.length - 1]!.push(group)
    usedHeight += height
  }

  return pages.map((pageGroups, index) => ({
    width: POSTER_WIDTH,
    height: Math.max(POSTER_HEADER_HEIGHT + POSTER_FOOTER_HEIGHT + pageGroups.reduce((sum, group) => sum + groupHeight(group.members.length), 0), 520),
    pageNumber: index + 1,
    pageCount: pages.length,
    familyName,
    subtitle,
    scope,
    groups: pageGroups
  }))
}

function fitText(ctx: UniApp.CanvasContext, text: string, maxWidth: number): string {
  if (ctx.measureText(text).width <= maxWidth) return text
  let result = text
  while (result.length > 1 && ctx.measureText(`${result}…`).width > maxWidth) result = result.slice(0, -1)
  return `${result}…`
}

export function drawExportLineagePoster(ctx: UniApp.CanvasContext, page: ExportLineagePosterPage) {
  const margin = 34
  const contentWidth = page.width - margin * 2
  const gap = 16
  const cardWidth = (contentWidth - gap) / 2
  let y = 0

  ctx.setFillStyle('#f8f0df')
  ctx.fillRect(0, 0, page.width, page.height)
  ctx.setStrokeStyle('#b88a52')
  ctx.setLineWidth(2)
  ctx.strokeRect(18, 18, page.width - 36, page.height - 36)

  y = 62
  ctx.setFillStyle('#9e4c3b')
  ctx.setFontSize(15)
  ctx.setTextAlign('center')
  ctx.fillText('FAMILY GENEALOGY', page.width / 2, y)
  y += 44
  ctx.setFillStyle('#193856')
  ctx.setFontSize(34)
  ctx.fillText(fitText(ctx, page.familyName, contentWidth), page.width / 2, y)
  y += 32
  ctx.setFillStyle('#718096')
  ctx.setFontSize(16)
  ctx.fillText(fitText(ctx, page.subtitle, contentWidth), page.width / 2, y)
  y += 54
  ctx.setStrokeStyle('#c9a56b')
  ctx.setLineWidth(1)
  ctx.beginPath()
  ctx.moveTo(margin, y)
  ctx.lineTo(page.width - margin, y)
  ctx.stroke()
  y += 36

  for (const group of page.groups) {
    ctx.setTextAlign('left')
    ctx.setFillStyle('#9e4c3b')
    ctx.setFontSize(20)
    ctx.fillText(`第 ${group.generation} 代${group.continued ? '（续）' : ''}`, margin, y)
    ctx.setStrokeStyle('#c9a56b')
    ctx.beginPath()
    ctx.moveTo(margin + 112, y - 6)
    ctx.lineTo(page.width - margin, y - 6)
    ctx.stroke()
    y += 20

    group.members.forEach((item, index) => {
      const column = index % 2
      const row = Math.floor(index / 2)
      const x = margin + column * (cardWidth + gap)
      const cardY = y + row * (POSTER_CARD_HEIGHT + POSTER_ROW_GAP)
      ctx.setFillStyle('#fffaf0')
      ctx.fillRect(x, cardY, cardWidth, POSTER_CARD_HEIGHT)
      ctx.setStrokeStyle('#d5b47b')
      ctx.strokeRect(x, cardY, cardWidth, POSTER_CARD_HEIGHT)
      ctx.setFillStyle('#193856')
      ctx.setFontSize(25)
      ctx.setTextAlign('left')
      const display = getExportLineageDisplay(item, page.scope)
      ctx.fillText(fitText(ctx, display.name, cardWidth - 28), x + 14, cardY + 37)
      const yearText = display.yearText
      if (yearText) {
        ctx.setFillStyle('#718096')
        ctx.setFontSize(15)
        ctx.fillText(yearText, x + 14, cardY + 67)
      }
      const parentText = display.parentLabels.join(' · ')
      if (parentText) {
        ctx.setFillStyle('#62768d')
        ctx.setFontSize(14)
        ctx.fillText(fitText(ctx, parentText, cardWidth - 28), x + 14, cardY + 91)
      }
      const spouseText = display.spouseLabels.length ? `（${display.spouseLabels.join('、')}）` : ''
      if (spouseText) {
        ctx.setFillStyle('#855f4e')
        ctx.setFontSize(14)
        ctx.fillText(fitText(ctx, spouseText, cardWidth - 28), x + 14, cardY + (parentText ? 110 : 94))
      }
    })
    y += Math.ceil(group.members.length / 2) * POSTER_CARD_HEIGHT +
      Math.max(0, Math.ceil(group.members.length / 2) - 1) * POSTER_ROW_GAP + POSTER_GROUP_GAP
  }

  ctx.setStrokeStyle('#c9a56b')
  ctx.beginPath()
  ctx.moveTo(margin, page.height - 58)
  ctx.lineTo(page.width - margin, page.height - 58)
  ctx.stroke()
  ctx.setFillStyle('#718096')
  ctx.setFontSize(14)
  ctx.setTextAlign('left')
  ctx.fillText('Tree · 家谱记录与家人协作', margin, page.height - 30)
  ctx.setTextAlign('right')
  ctx.fillText(page.pageCount > 1 ? `${page.pageNumber} / ${page.pageCount}` : '宗支长卷', page.width - margin, page.height - 30)
}
