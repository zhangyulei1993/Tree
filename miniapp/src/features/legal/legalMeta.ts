/** 运营主体信息：以下字段须由运营方确认后替换，不得编造。 */
export const LEGAL_OPERATOR = {
  name: '[填写真实姓名]',
  privacyEmail: '[填写邮箱]',
  phone: '[填写或注明仅通过邮箱联系]',
  serverProvider: '[填写]',
  serverRegion: '[填写，例如中国大陆]',
  smsProvider: '[填写]',
  dataRetention: '[填写]'
} as const

export const LEGAL_DOCUMENT_VERSION = '1.0.0'
export const LEGAL_EFFECTIVE_DATE = '2026-06-11'
export const LEGAL_UPDATED_DATE = '2026-06-11'
export const LEGAL_PRODUCT_NAME = 'Tree（家脉亲缘）'

export const LEGAL_META_LINE = `版本 ${LEGAL_DOCUMENT_VERSION} · 生效日期 ${LEGAL_EFFECTIVE_DATE} · 更新日期 ${LEGAL_UPDATED_DATE}`
