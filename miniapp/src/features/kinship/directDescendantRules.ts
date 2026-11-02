import type { Gender } from './types'

const DESCENDANT_BASE_BY_DEPTH: Record<number, string> = {
  3: '曾孙',
  4: '玄孙',
  5: '来孙'
}

function genderSuffix(gender: Gender): string | null {
  if (gender === 'male') return ''
  if (gender === 'female') return '女'
  return null
}

function isPaternalDescendantPath(genders: Gender[]): boolean {
  if (genders.length === 0 || genders[0] !== 'male') return false
  return genders.slice(1, -1).every((gender) => gender === 'male')
}

/**
 * 直系晚辈的统一口径：第一层通过儿子为内系，第一层通过女儿为外系；
 * 后续任一非末端节点通过女儿延续，也转为外系。只推导到五层。
 */
export function resolveDirectDescendantTitle(genders: Gender[]): string | null {
  const depth = genders.length
  const targetSuffix = genderSuffix(genders[depth - 1])
  if (targetSuffix === null || depth < 1 || depth > 5) return null

  if (depth === 1) {
    if (genders[0] === 'male') return '儿子'
    if (genders[0] === 'female') return '女儿'
    return null
  }

  if (depth === 2) {
    if (genders[0] === 'male') return genders[depth - 1] === 'male' ? '孙子' : '孙女'
    return genders[depth - 1] === 'male' ? '外孙' : '外孙女'
  }

  const base = DESCENDANT_BASE_BY_DEPTH[depth]
  if (!base) return null
  const prefix = isPaternalDescendantPath(genders) ? '' : '外'
  return `${prefix}${base}${targetSuffix}`
}

/** 直系晚辈配偶称谓，要求血缘链不超过五层。 */
export function resolveDirectDescendantSpouseTitle(
  bloodGenders: Gender[],
  spouseGender: Gender
): string | null {
  const bloodTitle = resolveDirectDescendantTitle(bloodGenders)
  if (!bloodTitle || bloodGenders.length === 0) return null

  const bloodGender = bloodGenders[bloodGenders.length - 1]
  if (bloodGender === 'male' && spouseGender === 'female') {
    if (bloodGenders.length === 1) return '儿媳'
    if (bloodGenders.length === 2) return `${bloodTitle.replace(/子$/, '')}媳`
    return `${bloodTitle}${bloodTitle.startsWith('外') ? '媳' : '媳妇'}`
  }
  if (bloodGender === 'female' && spouseGender === 'male') {
    if (bloodGenders.length === 1) return '女婿'
    return `${bloodTitle}婿`
  }
  return null
}
