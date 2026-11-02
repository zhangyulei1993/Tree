export type AvatarAgeBand =
  | 'toddler'
  | 'child'
  | 'teen'
  | 'youth'
  | 'adult'
  | 'mature'
  | 'senior'
  | 'elder'
  | 'centenarian'

function parseMemberDate(value?: string | null) {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null
  const date = new Date(`${value}T00:00:00`)
  return Number.isNaN(date.getTime()) ? null : date
}

/** Calculates visual age on the exact anniversary; deceased members use their death date. */
export function resolveAvatarAgeBand(
  birthDate?: string | null,
  referenceDate?: string | null
): AvatarAgeBand {
  const birth = parseMemberDate(birthDate)
  const reference = parseMemberDate(referenceDate) || new Date()
  if (!birth || birth > reference) return 'adult'

  let age = reference.getFullYear() - birth.getFullYear()
  const anniversaryNotReached =
    reference.getMonth() < birth.getMonth() ||
    (reference.getMonth() === birth.getMonth() && reference.getDate() < birth.getDate())
  if (anniversaryNotReached) age -= 1

  if (age <= 6) return 'toddler'
  if (age <= 12) return 'child'
  if (age <= 17) return 'teen'
  if (age <= 29) return 'youth'
  if (age <= 44) return 'adult'
  if (age <= 59) return 'mature'
  if (age <= 74) return 'senior'
  if (age <= 99) return 'elder'
  return 'centenarian'
}
