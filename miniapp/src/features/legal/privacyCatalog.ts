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

const LEGAL_RETENTION_PLACEHOLDER = `保存至账号注销、家庭解散或法律要求的期限；具体期限见运营方确认的「${LEGAL_OPERATOR.dataRetention}」`

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
    fields: '手机号、登录密码（哈希保存）、短信验证码（哈希保存）、微信登录凭证 code（服务端换取 openid/unionid）、昵称、头像',
    purpose: '注册登录、绑定手机号、找回与变更联系方式、账号注销、身份识别与资料展示',
    method: '你主动填写或授权；微信登录时由微信向服务端提供 openid/unionid（小程序前端不展示完整值）',
    necessary: '使用家庭协作、加入申请、邀请绑定、留言等需登录且绑定手机号的功能所必需；浏览公开家庭主页、使用本地称谓工具可不提供',
    retention: LEGAL_RETENTION_PLACEHOLDER
  },
  {
    category: '家庭成员节点资料',
    fields: '成员姓名、性别、出生年份、去世年份、是否健在、成员简介、成员类型（本家成员/配偶/外系成员）',
    purpose: '建立家谱节点、展示亲属关系、权限判断与家庭管理',
    method: '家庭创建者或管理员录入；部分字段由加入申请人填写',
    necessary: '维护家谱与家庭协作所必需；你应确保已取得被录入成员的合法授权',
    retention: LEGAL_RETENTION_PLACEHOLDER
  },
  {
    category: '亲属关系',
    fields: '父母子女关系、配偶关系、关系备注',
    purpose: '生成家谱结构、计算权限与展示亲属联结',
    method: '家庭管理员在小程序内创建或维护',
    necessary: '家谱核心功能所必需',
    retention: LEGAL_RETENTION_PLACEHOLDER
  },
  {
    category: '邀请与加入',
    fields: '邀请说明、邀请状态、加入申请姓名/性别/说明/关系意向、处理结果',
    purpose: '邀请成员绑定账号、审核加入申请、联系申请人',
    method: '你或家庭管理员主动提交；申请人填写申请表单',
    necessary: '邀请与加入流程所必需',
    retention: LEGAL_RETENTION_PLACEHOLDER
  },
  {
    category: '访客留言',
    fields: '留言称呼、留言内容、联系电话、微信号（后两者不公开显示）',
    purpose: '游客向公开家庭留言，供家庭管理员审核与联系',
    method: '你在公开家庭主页主动填写并提交',
    necessary: '提交留言所必需；联系方式为可选',
    retention: LEGAL_RETENTION_PLACEHOLDER
  },
  {
    category: '公开家庭展示',
    fields: '家庭名称、姓氏、地区、简介、公开联系方式、经审核的公开树字段',
    purpose: '公开展示家庭主页与公开树',
    method: '家庭管理员申请并经平台审核后展示',
    necessary: '公开展示所必需；未申请公开前不对游客展示私有成员详情',
    retention: LEGAL_RETENTION_PLACEHOLDER
  },
  {
    category: '安全与运行日志',
    fields: 'IP 地址、User-Agent、登录时间、重要操作类型与时间（不含明文密码、验证码、完整 token、完整 openid/unionid）',
    purpose: '保障账号与家庭数据安全、满足审计与纠纷核查',
    method: '你使用服务时由服务端自动记录',
    necessary: '网络安全与合规所必需',
    retention: LEGAL_RETENTION_PLACEHOLDER
  }
]

export const WECHAT_PRIVACY_DECLARATIONS: WechatPrivacyDeclaration[] = [
  {
    apiOrScene: 'wx.login / uni.login（微信登录）',
    information: '微信登录凭证；服务端获取 openid、unionid（如可用）',
    purpose: '创建或识别平台账号、完成微信登录',
    usedInCode: 'pages/auth/wechat-login.vue、stores/session.ts'
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
    usedInCode: 'pages/me/profile.vue、pages/auth/register-phone.vue（可选）'
  },
  {
    apiOrScene: '用户主动输入手机号 + 短信验证码',
    information: '手机号、验证码',
    purpose: '注册、登录、绑定手机号、变更手机号、注销账号',
    usedInCode: 'pages/auth/phone-login.vue、pages/auth/bind-phone.vue、pages/auth/register-phone.vue、pages/account/cancel.vue、pages/account/change-phone.vue'
  }
]

/** 代码中未使用、不得在隐私指引中声明的能力 */
export const NOT_COLLECTED_IN_APP = [
  '微信手机号一键获取（getPhoneNumber）',
  '通讯录、地理位置、麦克风、摄像头（除头像选择外）、蓝牙、日历、运动数据',
  '支付账号与交易信息（小程序当前无支付功能）'
] as const

export const THIRD_PARTY_PROCESSORS = [
  {
    name: '深圳市腾讯计算机系统有限公司',
    role: '微信开放平台（登录、头像选择等微信能力）',
    region: '中国境内'
  },
  {
    name: LEGAL_OPERATOR.smsProvider,
    role: '短信验证码发送',
    region: LEGAL_OPERATOR.serverRegion
  },
  {
    name: LEGAL_OPERATOR.serverProvider,
    role: '服务器与数据存储托管',
    region: LEGAL_OPERATOR.serverRegion
  }
] as const
