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
    fields: '微信登录凭证 code（服务端换取 openid/unionid）、昵称、头像（可选）',
    purpose: '微信登录、身份识别、资料展示与账号注销重新认证',
    method: '你授权微信登录；昵称与头像由你主动填写或选择',
    necessary: '使用家庭协作、加入申请、邀请绑定等功能需登录并设置昵称；浏览公开家庭主页、使用本地称谓工具可不登录',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '手机号登录凭证',
    fields: '手机号、登录密码单向哈希值（不存储明文密码）',
    purpose: '设置或使用手机密码登录、提升账号权益等级',
    method: '你在完善昵称后主动填写手机号并设置密码；登录时输入手机号与密码',
    necessary: '可选；不设置仍可使用微信登录及基础功能；设置后可使用手机号密码登录并获得更高权益',
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
    category: '访客留言',
    fields: '留言称呼、留言内容、联系电话、微信号（后两者不公开显示）',
    purpose: '游客向公开家庭留言，供家庭管理员审核与联系',
    method: '你在公开家庭主页主动填写并提交',
    necessary: '提交留言所必需；联系方式为可选',
    retention: LEGAL_RETENTION_TEXT
  },
  {
    category: '公开家庭展示',
    fields: '家庭名称、姓氏、地区、简介、公开联系方式、经审核的公开树字段',
    purpose: '公开展示家庭主页与公开树',
    method: '家庭管理员申请并经平台审核后展示',
    necessary: '公开展示所必需；未申请公开前不对游客展示私有成员详情',
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
  {
    apiOrScene: '用户主动填写（手机号与密码）',
    information: '手机号、登录密码（服务端仅保存密码单向哈希，不存明文）',
    purpose: '可选备用登录方式；不验证号码所有权，不采集短信验证码',
    usedInCode: 'pages/me/profile.vue、pages/auth/phone-login.vue'
  }
]

/** 代码中未使用、不得在隐私指引中声明的能力 */
export const NOT_COLLECTED_IN_APP = [
  '短信验证码',
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
    name: LEGAL_OPERATOR.serverProvider,
    role: '服务器与数据存储托管',
    region: LEGAL_OPERATOR.serverRegion
  }
] as const
