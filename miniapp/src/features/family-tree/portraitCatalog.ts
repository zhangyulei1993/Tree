export const portraitAgeBands = [
  'toddler',
  'child',
  'teen',
  'youth',
  'adult',
  'mature',
  'senior',
  'elder',
  'centenarian'
] as const

export const regionalPortraitThemes = ['north', 'south'] as const

export type RegionalPortraitTheme = typeof regionalPortraitThemes[number]
export type PortraitGender = 'male' | 'female'
export type PortraitHairStyle = 'AUTO' | 'SHORT' | 'BOB' | 'BRAIDS' | 'LOW_BUN' | 'WAVY' | 'LONG'
export type PortraitAvatarStyle = 'AUTO' | `BASE_${number}` | `GLASSES_${number}`
export type PortraitAgeBand = typeof portraitAgeBands[number]
export type PortraitAppearanceMember = {
  memberId: number
  gender?: string | null
  birthDate?: string | null
  deathDate?: string | null
  isLiving?: boolean | null
}

const northRegionPattern = /北京|天津|河北|山西|山东|河南|辽宁|吉林|黑龙江|内蒙古|陕西|甘肃|宁夏|青海|新疆/

/**
 * Family hometowns now resolve to two intentionally broad visual languages.
 * The division follows the north/south convention rather than administrative regions.
 */
export function resolveRegionalPortraitTheme(
  regionCode?: string | null,
  regionText?: string | null
): RegionalPortraitTheme {
  const code = regionCode || ''
  if (
    code.startsWith('CN-NORTH-') ||
    code.startsWith('CN-NORTHEAST-') ||
    code.startsWith('CN-NORTHWEST-')
  ) return 'north'

  const region = (regionText || '').replace(/\s/g, '')
  return northRegionPattern.test(region) ? 'north' : 'south'
}

const portraitAppearanceCount = 12
const portraitAssetVariantCount = 12
// These optional illustration sets are hosted with the existing avatar static
// service so every mini program install does not carry the full portrait atlas.
export const portraitAssetBaseURL = 'https://tapi.bigbigboy.cn/api/static/avatars/tree-portraits'
const portraitBaseAssetRevision = 'age-only-editorial-20261014-112'

/**
 * The visual treatment has twelve stable combinations, so a sizeable sibling
 * group does not start repeating faces after four people.
 */
export function resolvePortraitAppearanceVariant(memberId?: number | null) {
  const value = Math.abs(Number(memberId) || 0)
  return (value % portraitAppearanceCount) + 1
}

/** Selects one of the twelve independent illustration bases for an appearance. */
export function resolvePortraitVariant(memberId?: number | null) {
  return ((resolvePortraitAppearanceVariant(memberId) - 1) % portraitAssetVariantCount) + 1
}

type FemaleHairVariantMap = Record<Exclude<PortraitHairStyle, 'AUTO'>, number>

/**
 * The twelve portrait files in each age group are illustrations, not a shared
 * hairstyle index. Keep this table explicit so a selected style always points
 * to the matching drawing for that age group.
 */
const femaleHairVariantsByAge: Record<PortraitAgeBand, FemaleHairVariantMap> = {
  toddler: { SHORT: 4, BOB: 4, BRAIDS: 3, LOW_BUN: 2, WAVY: 1, LONG: 6 },
  child: { SHORT: 4, BOB: 4, BRAIDS: 5, LOW_BUN: 2, WAVY: 6, LONG: 3 },
  teen: { SHORT: 2, BOB: 2, BRAIDS: 1, LOW_BUN: 4, WAVY: 3, LONG: 5 },
  youth: { SHORT: 3, BOB: 3, BRAIDS: 5, LOW_BUN: 1, WAVY: 4, LONG: 2 },
  adult: { SHORT: 6, BOB: 2, BRAIDS: 1, LOW_BUN: 4, WAVY: 3, LONG: 5 },
  mature: { SHORT: 6, BOB: 3, BRAIDS: 5, LOW_BUN: 1, WAVY: 4, LONG: 5 },
  senior: { SHORT: 6, BOB: 6, BRAIDS: 5, LOW_BUN: 3, WAVY: 5, LONG: 2 },
  elder: { SHORT: 1, BOB: 1, BRAIDS: 5, LOW_BUN: 3, WAVY: 4, LONG: 6 },
  centenarian: { SHORT: 5, BOB: 5, BRAIDS: 6, LOW_BUN: 1, WAVY: 4, LONG: 6 }
}

export function resolveFemalePortraitVariant(
  memberId?: number | null,
  hairStyle?: string | null,
  ageBand: PortraitAgeBand = 'adult'
) {
  const style = String(hairStyle || 'AUTO').toUpperCase() as PortraitHairStyle
  const variants = femaleHairVariantsByAge[ageBand]
  return style !== 'AUTO' && variants[style]
    ? variants[style]
    : resolvePortraitVariant(memberId)
}

/**
 * Avoid duplicates inside the same gender and visual-age group. Existing
 * members retain their preferred appearance unless it is already occupied.
 */
export function buildPortraitAppearanceMap(members: PortraitAppearanceMember[]) {
  const groups = new Map<string, PortraitAppearanceMember[]>()
  for (const member of members) {
    if (member.gender !== 'MALE' && member.gender !== 'FEMALE') continue
    const ageBand = resolveAvatarAgeBand(
      member.birthDate,
      member.isLiving === false ? member.deathDate : undefined
    )
    const key = `${member.gender}:${ageBand}`
    groups.set(key, [...(groups.get(key) || []), member])
  }

  const appearances: Record<number, number> = {}
  for (const group of groups.values()) {
    const occupied = new Set<number>()
    const occupiedAssets = new Set<number>()
    for (const member of [...group].sort((left, right) => left.memberId - right.memberId)) {
      let appearance = resolvePortraitAppearanceVariant(member.memberId)
      for (
        let attempt = 0;
        (occupied.has(appearance) || occupiedAssets.has(resolvePortraitVariant(appearance - 1))) &&
          attempt < portraitAppearanceCount;
        attempt += 1
      ) {
        appearance = (appearance % portraitAppearanceCount) + 1
      }
      appearances[member.memberId] = appearance
      occupied.add(appearance)
      occupiedAssets.add(resolvePortraitVariant(appearance - 1))
    }
  }
  return appearances
}

export function resolveRegionalPortraitAsset(input: {
  theme: RegionalPortraitTheme
  gender: PortraitGender
  ageBand: PortraitAgeBand
  memberId?: number | null
  variant?: number
}) {
  const variant = input.variant || resolvePortraitVariant(input.memberId)
  return `${portraitAssetBaseURL}/age-only/base/${input.gender}-${input.ageBand}-${variant}.png?v=${portraitBaseAssetRevision}`
}

export function resolveRegionalPortraitGlassesAsset(input: {
  theme: RegionalPortraitTheme
  gender: PortraitGender
  ageBand: PortraitAgeBand
  memberId?: number | null
  variant?: number
}) {
  const variant = input.variant || resolvePortraitVariant(input.memberId)
  return `${portraitAssetBaseURL}/age-only/glasses/${input.gender}-${input.ageBand}-${variant}.png?v=${portraitBaseAssetRevision}`
}

export function resolvePortraitAvatarStyle(style?: string | null) {
  const normalized = String(style || 'AUTO').toUpperCase()
  const matched = /^(BASE|GLASSES)_([1-9]|1[0-2])$/.exec(normalized)
  return matched
    ? { mode: matched[1] === 'GLASSES' ? 'glasses' as const : 'base' as const, variant: Number(matched[2]) }
    : null
}

/** A selected portrait is one complete illustration, never an overlay. */
export function resolveSelectedPortraitAsset(input: {
  gender: PortraitGender
  ageBand: PortraitAgeBand
  style?: string | null
}) {
  const selection = resolvePortraitAvatarStyle(input.style)
  if (!selection) return ''
  const directory = selection.mode === 'glasses' ? 'glasses' : 'base'
  return `${portraitAssetBaseURL}/age-only/${directory}/${input.gender}-${input.ageBand}-${selection.variant}.png?v=${portraitBaseAssetRevision}`
}
import { resolveAvatarAgeBand } from './avatarAgeBand'
