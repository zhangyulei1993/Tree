export function maskPhone(phone?: string | null) {
  const value = (phone || '').trim()
  if (!value) return ''
  if (value.length < 7) return value
  return `${value.slice(0, 3)}****${value.slice(-4)}`
}
