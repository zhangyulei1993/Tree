/** 运营主体信息 */
export const LEGAL_OPERATOR = {
  name: '张玉磊',
  privacyEmail: 'zhangyulei1993@gmail.com',
  phone: '通过隐私联系邮箱联系',
  serverProvider: '阿里云计算有限公司',
  serverRegion: '中国浙江省杭州市',
  dataRetention:
    '账号存续期间保存；账号注销完成后 30 日内删除或匿名化账号相关个人信息；家庭历史节点数据在家庭解散或法律要求的期限内保留；安全与操作日志保存不超过 6 个月'
} as const

export const LEGAL_DOCUMENT_VERSION = '1.2.0'
export const LEGAL_EFFECTIVE_DATE = '2026-06-11'
export const LEGAL_UPDATED_DATE = '2026-07-04'
export const LEGAL_PRODUCT_NAME = 'Tree（家脉亲缘）'

export const LEGAL_META_LINE = `版本 ${LEGAL_DOCUMENT_VERSION} · 生效日期 ${LEGAL_EFFECTIVE_DATE} · 更新日期 ${LEGAL_UPDATED_DATE}`
