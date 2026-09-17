export const CHOICE_TITLE_MAX_LENGTH = 20
export const CHOICE_OPTION_MAX_LENGTH = 30
export const MAX_CHOICE_OPTIONS = 20
export const MAX_CHOICE_OPTIONS_TEXT_LENGTH = 500

export function parseChoiceOptions(input: string) {
  const uniqueOptions: string[] = []
  const parts = input.split(/[\r\n,，、;；]+/)
  for (const part of parts) {
    const option = part.trim().replace(/[\u0000-\u001f\u007f]/g, '')
    if (!option || uniqueOptions.includes(option)) continue
    uniqueOptions.push(option)
    if (uniqueOptions.length >= MAX_CHOICE_OPTIONS) break
  }
  return uniqueOptions
}

export function pickRandomChoice(
  options: readonly string[],
  random = Math.random,
  previousOption = ''
) {
  if (options.length === 0) return ''
  const candidates = previousOption && options.length > 1
    ? options.filter((option) => option !== previousOption)
    : [...options]
  const safeRandom = Math.max(0, Math.min(0.999999, random()))
  return candidates[Math.floor(safeRandom * candidates.length)] || candidates[0] || ''
}
