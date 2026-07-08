export type TextFieldKind = 'surname' | 'name' | 'singleLine' | 'multiLine'

export interface TextFieldValidation {
  value: string | null | undefined
  label: string
  kind?: TextFieldKind
  required?: boolean
  maxLength?: number
}

const surnamePattern = /^[\u3400-\u9fff\uf900-\ufaffA-Za-z·•．._\-—]+$/u
const namePattern = /^[\u3400-\u9fff\uf900-\ufaffA-Za-z0-9·•．._\-—（）() 　]+$/u
const datePattern = /^\d{4}-\d{2}-\d{2}$/
const forbiddenSingleLinePattern = /[<>`{}\r\n\t]|[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/u
const forbiddenMultiLinePattern = /[<>`{}]|[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/u

export function normalizeText(value: string | null | undefined) {
  return String(value || '').trim()
}

export function optionalText(value: string | null | undefined) {
  const normalized = normalizeText(value)
  return normalized || undefined
}

export function validateTextField(field: TextFieldValidation) {
  const value = normalizeText(field.value)
  const kind = field.kind || 'singleLine'
  if (field.required && !value) return `请填写${field.label}。`
  if (!value) return ''
  if (field.maxLength && value.length > field.maxLength) {
    return `${field.label}不能超过 ${field.maxLength} 个字符。`
  }
  if (kind === 'surname' && !surnamePattern.test(value)) {
    return `${field.label}仅支持中文、英文字母、间隔点或短横线。`
  }
  if (kind === 'name' && !namePattern.test(value)) {
    return `${field.label}仅支持中文、英文字母、数字、空格、间隔点或短横线。`
  }
  const forbiddenPattern = kind === 'multiLine' ? forbiddenMultiLinePattern : forbiddenSingleLinePattern
  if (forbiddenPattern.test(value)) {
    return `${field.label}不能包含 <、>、反引号、花括号或控制字符。`
  }
  return ''
}

export function validateTextFields(fields: TextFieldValidation[]) {
  for (const field of fields) {
    const message = validateTextField(field)
    if (message) return message
  }
  return ''
}

export function validateDateInput(value: string | null | undefined, label: string) {
  const normalized = normalizeText(value)
  if (!normalized) return ''
  if (forbiddenSingleLinePattern.test(normalized)) {
    return `${label}不能包含特殊字符或控制字符。`
  }
  if (!datePattern.test(normalized)) {
    return `${label}格式请使用 YYYY-MM-DD。`
  }
  return ''
}
