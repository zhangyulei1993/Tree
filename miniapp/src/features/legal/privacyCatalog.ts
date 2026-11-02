import { LEGAL_OPERATOR } from './legalMeta'

/**
 * 与小程序代码、后端接口实际一致的信息收集清单。
 * 用于隐私政策正文与微信后台《用户隐私保护指引》对照。
 */

export interface CollectedInfoItem {
  category: string
  fields: string
  purpose: string
  method: string
  necessary: string
  retention: string
}

export interface WechatPrivacyDeclaration {
  apiOrScene: string
  information: string
  purpose: string
  usedInCode: string
}

const LEGAL_RETENTION_TEXT = LEGAL_OPERATOR.dataRetention

/** 仅在本机处理、不上传服务端的信息 */
export const LOCAL_ONLY_INFO = [
  {
    scene: '亲属称谓工具',
    fields: '本人性别、本人出生日期或年龄（可选）、关系路径中各步性别与长幼（可选）',
    note: '仅在设备本地用于称谓推测，不调用家庭或账号相关接口，不写入服务端。'
  }
] as const

export const COLLECTED_INFO_ITEMS: CollectedInfoItem[] = [
  {
    category: '账号与身份',
    fields: '微信登录凭证 code（服务端换取 openid/unionid）、微信手机号授权 code（服务端换取手机号）、昵称、头像（可选）',
    purpose: '微信登录、手机号授权绑定、身份识别、资料展示与账号注销重新认证',
    method: '你授权微信登录或手机号授权绑定；昵称与头像由你主动填写或选择',
    necessary: '使用家庭协作、加入申请、邀请绑定等功能需登录并设置昵称；浏览展示家庭页面、经分享链接访问公开家庭主页、使用本地称谓工具可不登录',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '手机号绑定信息',
    fields: '微信手机号授权 code（服务端换取手机号）',
    purpose: '绑定手机号、提升账号权益等级，并在发起邀请等敏感操作前完成校验',
    method: '你在敏感操作前或账号资料中主动授权微信手机号绑定',
    necessary: '可选；不授权手机号仍可使用微信登录及基础功能；绑定后可获得更高权益并可发起邀请',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '家庭成员节点资料',
    fields: '成员姓名、性别、出生年份、去世年份、是否健在、成员简介、成员类型（本家成员/配偶/外系成员）',
    purpose: '建立家谱节点、展示亲属关系、权限判断与家庭管理',
    method: '家庭创建者或管理员录入；部分字段由加入申请人填写',
    necessary: '维护家谱与家庭协作所必需；你应确保已取得被录入成员的合法授权',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '亲属关系',
    fields: '父母子女关系、配偶关系、关系备注',
    purpose: '生成家谱结构、计算权限与展示亲属联结',
    method: '家庭管理员在小程序内创建或维护',
    necessary: '家谱核心功能所必需',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '邀请与加入',
    fields: '邀请说明、邀请状态、加入申请姓名/性别/说明/关系意向、处理结果',
    purpose: '邀请成员绑定账号、审核加入申请、联系申请人',
    method: '你或家庭管理员主动提交；申请人填写申请表单',
    necessary: '邀请与加入流程所必需',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '家庭容量订单与退款',
    fields: '订单号、商品名称、购买名额、金额、支付交易标识、订单状态、退款申请理由和审核结果',
    purpose: '完成家庭容量购买、退款审核、原路退款、对账和争议处理',
    method: '你主动购买容量或提交退款申请时由服务端记录；支付由微信支付处理',
    necessary: '购买、退款与售后处理所必需；我们不直接保存完整支付账户信息',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '历史世系示例与分享浏览',
    fields: '平台整理的历史世系示例字段、经家庭成员主动分享链接访问时所需的脱敏展示字段',
    purpose: '展示平台整理的历史示例，或支持通过家庭成员分享链接浏览必要信息',
    method: '平台整理发布，或由家庭成员主动分享链接触发访问',
    necessary: '示例浏览和分享访问所必需；私有家庭资料默认不进入平台示例页面',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '安全与运行日志',
    fields: 'IP 地址、User-Agent、登录时间、重要操作类型与时间（不含完整 token、完整 openid/unionid）',
    purpose: '保障账号与家庭数据安全、满足审计与纠纷核查',
    method: '你使用服务时由服务端自动记录',
    necessary: '网络安全与合规所必需',
    retention: LEGAL_RETENTION_TEXT
  }
]

export const WECHAT_PRIVACY_DECLARATIONS: WechatPrivacyDeclaration[] = [
  {
    apiOrScene: 'wx.login / uni.login（微信登录）',
    information: '微信登录凭证；服务端获取 openid、unionid（如可用）',
    purpose: '创建或识别平台账号、完成微信登录与注销重新认证',
    usedInCode: 'pages/auth/wechat-login.vue、pages/account/cancel.vue、stores/session.ts'
  },
  {
    apiOrScene: 'getPhoneNumber（手机号授权绑定）',
    information: '手机号授权 code；服务端获取手机号',
    purpose: '绑定手机号以提升账号权益并支持敏感操作前身份校验',
    usedInCode: 'pages/me/account-security.vue、pages/invite/sent.vue、api/auth.ts、stores/session.ts'
  },
  {
    apiOrScene: 'chooseAvatar（选择头像）',
    information: '头像图片',
    purpose: '设置账号头像，便于家庭成员识别',
    usedInCode: 'pages/me/profile.vue'
  },
  {
    apiOrScene: 'input type=nickname（昵称填写）',
    information: '昵称',
    purpose: '设置或更新账号昵称',
    usedInCode: 'pages/me/profile.vue'
  },
]

/** 代码中未使用、不得在隐私指引中声明的能力 */
export const NOT_COLLECTED_IN_APP = [
  '短信验证码',
  '通讯录、地理位置、麦克风、摄像头（除头像选择外）、蓝牙、日历、运动数据',
  '完整支付账号、银行卡号或支付密码（支付由微信支付处理）'
] as const

export const THIRD_PARTY_PROCESSORS = [
  {
    name: '深圳市腾讯计算机系统有限公司',
    role: '微信开放平台与微信支付（登录、手机号授权、头像选择、支付与退款等微信能力）',
    region: '中国境内'
  },
  {
    name: LEGAL_OPERATOR.serverProvider,
    role: '服务器与数据存储托管',
    region: LEGAL_OPERATOR.serverRegion
  }
] as const
