import chinaAreaData from './china-area-data.json'
import chinaTownshipData from './china-township-data.json'

export interface GeographicZone {
  code: string
  label: string
  provinces: string[]
}

export interface NativePlaceSelection {
  province: string
  city?: string
  district?: string
}

export interface NativePlaceResolution {
  nativePlace: string
  regionCode: string
  regionText: string
  zone: GeographicZone
  province: string
  city: string
  district: string
  township: string
  kind: 'MUNICIPALITY_DISTRICT' | 'PROVINCE_CITY' | 'PROVINCE_CITY_COUNTY' | 'PROVINCE_CITY_DISTRICT'
}

export const geographicZones: GeographicZone[] = [
  { code: 'NORTH', label: '华北地区', provinces: ['北京市', '天津市', '河北省', '山西省', '内蒙古自治区'] },
  { code: 'NORTHEAST', label: '东北地区', provinces: ['辽宁省', '吉林省', '黑龙江省'] },
  { code: 'EAST', label: '华东地区', provinces: ['上海市', '江苏省', '浙江省', '安徽省', '福建省', '江西省', '山东省', '台湾省'] },
  { code: 'CENTRAL', label: '华中地区', provinces: ['河南省', '湖北省', '湖南省'] },
  { code: 'SOUTH', label: '华南地区', provinces: ['广东省', '广西壮族自治区', '海南省', '香港特别行政区', '澳门特别行政区'] },
  { code: 'SOUTHWEST', label: '西南地区', provinces: ['重庆市', '四川省', '贵州省', '云南省', '西藏自治区'] },
  { code: 'NORTHWEST', label: '西北地区', provinces: ['陕西省', '甘肃省', '青海省', '宁夏回族自治区', '新疆维吾尔自治区'] }
]

const municipalities = new Set(['北京市', '天津市', '上海市', '重庆市', '香港特别行政区', '澳门特别行政区'])
type AreaMap = Record<string, Record<string, string>>

interface AreaRecord {
  province: string
  city: string
  district: string
}

interface TownshipRecord extends AreaRecord {
  township: string
}

const areaMap = chinaAreaData as AreaMap
const townshipMap = chinaTownshipData as Record<string, Record<string, Record<string, string[]>>>
const areaRecords: AreaRecord[] = []
const townshipRecords: TownshipRecord[] = []
const townshipRecordsByExactName = new Map<string, TownshipRecord[]>()
const townshipRecordsByShortName = new Map<string, TownshipRecord[]>()
const areaRecordKeys = new Set<string>()
const searchablePlaceRecords: Array<{ record: AreaRecord; township: string }> = []

function addAreaRecord(record: AreaRecord) {
  const key = [record.province, record.city, record.district].join('|')
  if (!areaRecordKeys.has(key)) {
    areaRecordKeys.add(key)
    areaRecords.push(record)
  }
}

Object.entries(areaMap['86'] || {}).forEach(([provinceCode, province]) => {
  const cities = areaMap[provinceCode] || {}
  addAreaRecord({ province, city: '', district: '' })
  Object.entries(cities).forEach(([cityCode, cityName]) => {
    const city = cityName === '市辖区' || cityName === '县' ? province : cityName
    addAreaRecord({ province, city, district: '' })
    Object.values(areaMap[cityCode] || {}).forEach((district) => {
      addAreaRecord({ province, city, district })
    })
  })
})

Object.entries(townshipMap).forEach(([province, cities]) => {
  Object.entries(cities).forEach(([cityName, districts]) => {
    const city = cityName === '市辖区' || cityName === '县' ? province : cityName
    Object.entries(districts).forEach(([district, townships]) => {
      townships.forEach((township) => {
        const record = { province, city, district, township }
        townshipRecords.push(record)
        const sameNameRecords = townshipRecordsByExactName.get(township) || []
        sameNameRecords.push(record)
        townshipRecordsByExactName.set(township, sameNameRecords)
        placeAliases(township).forEach((alias) => {
          const sameAliasRecords = townshipRecordsByShortName.get(alias) || []
          sameAliasRecords.push(record)
          townshipRecordsByShortName.set(alias, sameAliasRecords)
        })
      })
    })
  })
})

areaRecords.forEach((record) => searchablePlaceRecords.push({ record, township: '' }))
townshipRecords.forEach((record) => searchablePlaceRecords.push({ record, township: record.township }))

export function findZoneByProvince(province: string) {
  return geographicZones.find((item) => item.provinces.includes(province))
}

export function isMunicipality(province: string) {
  return municipalities.has(province)
}

function cleanPart(value?: string) {
  return (value || '').replace(/\s/g, '').trim()
}

function normalizeKeywordAliases(value: string) {
  return value.replaceAll('簰洲湾', '牌洲湾')
}

function hasAdministrativeLevelSuffix(value: string) {
  return /(特别行政区|维吾尔自治区|壮族自治区|回族自治区|自治区|自治州|自治县|林区|特区|省|市|县|区|旗|地区|盟)$/.test(value)
}

function isCountyLevel(value: string) {
  return /(?:自治县|县|旗|林区|特区)$/.test(value)
}

/**
 * Produces the only location representation stored for a family. Native picker
 * values are kept as source data so municipal districts never become a fake
 * "city + city" path.
 */
export function resolveNativePlace(selection: NativePlaceSelection): NativePlaceResolution | null {
  const province = cleanPart(selection.province)
  const city = cleanPart(selection.city)
  const district = cleanPart(selection.district)
  const zone = findZoneByProvince(province)
  if (!zone || !province) return null

  const directMunicipality = isMunicipality(province)
  const normalizedCity = directMunicipality && city === province ? '' : city
  const lowest = district && district !== province && district !== city ? district : ''
  const hierarchy = directMunicipality
    ? [province, lowest].filter(Boolean)
    : [province, normalizedCity, lowest].filter(Boolean)
  const nativePlace = hierarchy.join('')
  const kind = directMunicipality
    ? 'MUNICIPALITY_DISTRICT'
    : !lowest
      ? 'PROVINCE_CITY'
      : isCountyLevel(lowest)
        ? 'PROVINCE_CITY_COUNTY'
        : 'PROVINCE_CITY_DISTRICT'
  const codeParts = [zone.code, province, normalizedCity || (directMunicipality ? province : ''), lowest].filter(Boolean)

  return {
    nativePlace,
    regionCode: `CN-${codeParts.join('-')}`,
    regionText: [zone.label, ...hierarchy].join(' · '),
    zone,
    province,
    city: normalizedCity,
    district: lowest,
    township: '',
    kind
  }
}

function placeAliases(value: string) {
  const compact = value.replace(/(特别行政区|维吾尔自治区|壮族自治区|回族自治区|自治区|自治州|自治县|林区|特区|省|市|县|区|旗|地区|盟|镇|乡|街道|苏木|农场|开发区)$/, '')
  return [...new Set([value, compact].filter(Boolean))]
}

function matchesPlaceKeyword(source: string, value: string) {
  if (source.includes(value)) return true
  // A supplied administrative suffix is meaningful: 朝阳区 must not silently
  // become 朝阳县. Short aliases are only for inputs such as 嘉鱼 or 杭州.
  if (/(特别行政区|维吾尔自治区|壮族自治区|回族自治区|自治区|自治州|自治县|林区|特区|省|市|县|区|旗|地区|盟|镇|乡|街道|苏木|农场|开发区)$/.test(source)) {
    return false
  }
  return placeAliases(value).some((alias) => source.includes(alias))
}

function matchesTownshipKeyword(source: string, township: string) {
  if (source.includes(township)) return true
  if (hasAdministrativeLevelSuffix(source)) {
    return false
  }
  return placeAliases(township).some((alias) => source.endsWith(alias))
}

/**
 * Resolves a place from the bundled national administrative-division index.
 * When the same short name exists in more than one place, it deliberately
 * returns null rather than attaching a family to an arbitrary region.
 */
export function resolveNativePlaceKeyword(keyword: string): NativePlaceResolution | null {
  const candidates = searchNativePlaceKeywords(keyword)
  return candidates.length === 1 ? candidates[0] : null
}

/**
 * Returns every equally plausible administrative path. Callers can present
 * these choices instead of silently assigning a reused township name to the
 * wrong county.
 */
export function searchNativePlaceKeywords(keyword: string): NativePlaceResolution[] {
  const source = normalizeKeywordAliases(cleanPart(keyword))
  if (!source) return []
  const exactTownshipRecords = townshipRecordsByExactName.get(source)
  const shortTownshipRecords = hasAdministrativeLevelSuffix(source) ? undefined : townshipRecordsByShortName.get(source)
  if (exactTownshipRecords?.length || shortTownshipRecords?.length) {
    const exactCandidates = new Map<string, NativePlaceResolution>()
    ;(exactTownshipRecords || shortTownshipRecords || []).forEach((record) => {
      const resolution = resolveNativePlace(record)
      if (!resolution) return
      const candidate = { ...resolution, township: record.township }
      exactCandidates.set([candidate.regionText, candidate.township].join('|'), candidate)
    })
    return [...exactCandidates.values()].sort((left, right) => nativePlaceQueryText(left).localeCompare(nativePlaceQueryText(right), 'zh-CN'))
  }
  const matches = searchablePlaceRecords
    .map(({ record, township }) => {
      const provinceMatched = matchesPlaceKeyword(source, record.province)
      const cityMatched = !!record.city && matchesPlaceKeyword(source, record.city)
      const districtMatched = !!record.district && matchesPlaceKeyword(source, record.district)
      const townshipMatched = !!township && matchesTownshipKeyword(source, township)
      const score = (provinceMatched ? 10_000 : 0) + (cityMatched ? 1_000 : 0) + (districtMatched ? 100 : 0) + (townshipMatched ? 100_000 : 0) + township.length + record.district.length + record.city.length
      return {
        record,
        township,
        score,
        provinceMatched,
        cityMatched,
        districtMatched,
        townshipMatched,
        matched: township ? townshipMatched : provinceMatched || cityMatched || districtMatched
      }
    })
    .filter((item) => item.matched)
    .sort((left, right) => right.score - left.score)

  // A township name alone is often reused across counties. Do not let a
  // longer unrelated administrative path win the score tie-breaker.
  const townshipOnlyMatches = matches.filter(
    (item) => item.townshipMatched && !item.provinceMatched && !item.cityMatched && !item.districtMatched
  )
  const toResolution = (item: typeof matches[number]) => {
    const resolution = resolveNativePlace(item.record)
    return resolution ? { ...resolution, township: item.township } : null
  }
  const uniqueResolutions = (items: typeof matches) => {
    const resolutions = new Map<string, NativePlaceResolution>()
    items.forEach((item) => {
      const resolution = toResolution(item)
      if (!resolution) return
      resolutions.set([resolution.regionText, resolution.township].join('|'), resolution)
    })
    return [...resolutions.values()].sort((left, right) => nativePlaceQueryText(left).localeCompare(nativePlaceQueryText(right), 'zh-CN'))
  }
  const hasTownshipContext = matches.some(
    (item) => item.townshipMatched && (item.provinceMatched || item.cityMatched || item.districtMatched)
  )
  if (!hasTownshipContext) {
    const townshipOnlyResolutions = uniqueResolutions(townshipOnlyMatches)
    if (townshipOnlyResolutions.length > 1) return townshipOnlyResolutions
  }

  const best = matches[0]
  if (!best) return []
  const sameScore = matches.filter((item) => item.score === best.score)
  const bestResolutions = uniqueResolutions(sameScore)
  return bestResolutions.length ? bestResolutions : []
}

export function nativePlaceQueryText(resolution: NativePlaceResolution) {
  return [resolution.regionText, resolution.township].filter(Boolean).join(' · ')
}

/** Formats legacy family-region values for display without changing stored data. */
export function formatFamilyRegionText(input: {
  nativePlace?: string | null
  regionText?: string | null
  regionCode?: string | null
}) {
  const nativePlace = cleanPart(input.nativePlace)
  const regionText = (input.regionText || '').trim()

  const regionCode = cleanPart(input.regionCode)
  const codeParts = regionCode.match(/^CN-([A-Z]+)-(.*)$/)
  if (codeParts) {
    const zone = geographicZones.find((item) => item.code === codeParts[1])
    const levels = codeParts[2].split('-').filter(Boolean)
    if (zone && levels.length) {
      return isMunicipality(levels[0])
        ? [zone.label, levels[0]].join(' · ')
        : [zone.label, ...levels.slice(0, 2)].join(' · ')
    }
  }

  const resolvedNativePlace = nativePlace ? resolveNativePlaceKeyword(nativePlace) : null
  if (resolvedNativePlace) {
    return isMunicipality(resolvedNativePlace.province)
      ? [resolvedNativePlace.zone.label, resolvedNativePlace.province].join(' · ')
      : [resolvedNativePlace.zone.label, resolvedNativePlace.province, resolvedNativePlace.city].filter(Boolean).join(' · ')
  }

  if (regionText.includes('·')) {
    const parts = regionText.split('·').map((item) => item.trim()).filter(Boolean)
    if (parts.length >= 2 && geographicZones.some((item) => item.label === parts[0])) {
      return isMunicipality(parts[1]) ? parts.slice(0, 2).join(' · ') : parts.slice(0, 3).join(' · ')
    }
    return regionText
  }

  // Legacy families sometimes stored only a city name such as "咸宁市".
  // Enrich only when that city belongs to a single province in the bundled data.
  const citySource = nativePlace || regionText
  if (citySource) {
    const cityMatches = new Map<string, AreaRecord>()
    areaRecords.forEach((record) => {
      if (!record.city || !citySource.includes(record.city)) return
      cityMatches.set([record.province, record.city].join('|'), record)
    })
    if (cityMatches.size === 1) {
      const match = [...cityMatches.values()][0]
      const zone = findZoneByProvince(match.province)
      if (zone) return [zone.label, match.province, match.city].join(' · ')
    }
  }

  return regionText || nativePlace || '未填写'
}

export function nativePlaceKindLabel(kind: NativePlaceResolution['kind']) {
  const labels: Record<NativePlaceResolution['kind'], string> = {
    MUNICIPALITY_DISTRICT: '直辖市 / 特别行政区辖区',
    PROVINCE_CITY: '省辖市',
    PROVINCE_CITY_COUNTY: '省辖市县',
    PROVINCE_CITY_DISTRICT: '省辖市辖区'
  }
  return labels[kind]
}
