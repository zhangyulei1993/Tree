import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import {
  COLLECTED_INFO_ITEMS,
  LOCAL_ONLY_INFO,
  NOT_COLLECTED_IN_APP,
  WECHAT_PRIVACY_DECLARATIONS
} from './privacyCatalog'
import { PRIVACY_POLICY_SECTIONS, USER_AGREEMENT_SECTIONS } from './legalContent'
import { LEGAL_DOCUMENT_VERSION, LEGAL_EFFECTIVE_DATE, LEGAL_UPDATED_DATE } from './legalMeta'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

const FORBIDDEN_PHRASES = [
  '占位',
  '最终法律文本',
  '最终文本为准',
  '上线前更新',
  'TODO',
  '待补充'
]

function readLegalSourceFiles(): string[] {
  const files = [
    join(miniappSrc, 'features', 'legal', 'legalMeta.ts'),
    join(miniappSrc, 'features', 'legal', 'legalContent.ts'),
    join(miniappSrc, 'features', 'legal', 'privacyCatalog.ts'),
    join(miniappSrc, 'features', 'legal', 'privacyConsent.ts'),
    join(miniappSrc, 'components', 'legal', 'LegalDocumentView.vue'),
    join(miniappSrc, 'components', 'legal', 'PrivacyConsentModal.vue'),
    join(miniappSrc, 'components', 'legal', 'AuthLegalConsent.vue'),
    join(miniappSrc, 'pages', 'legal', 'user-agreement.vue'),
    join(miniappSrc, 'pages', 'legal', 'privacy-policy.vue')
  ]
  return files.filter((file) => existsSync(file)).map((file) => readFileSync(file, 'utf8'))
}

test('legal documents do not contain placeholder or draft-only phrases', () => {
  const combined = readLegalSourceFiles().join('\n')
  for (const phrase of FORBIDDEN_PHRASES) {
    assert.equal(
      combined.includes(phrase),
      false,
      `legal sources must not include forbidden phrase: ${phrase}`
    )
  }
})

test('user agreement covers required topics', () => {
  const titles = USER_AGREEMENT_SECTIONS.map((section) => section.title).join('\n')
  const body = USER_AGREEMENT_SECTIONS.flatMap((section) => section.paragraphs).join('\n')
  const requiredSnippets = [
    '服务内容与运营者',
    '账号注册、登录与未成年人',
    '授权保证',
    '家庭角色与权限',
    '家谱节点与平台账号相互独立',
    '退出家庭、注销账号与解散家庭',
    '在世成员',
    '用户行为规范',
    '服务变更与责任限制',
    '投诉、联系与争议解决'
  ]
  for (const snippet of requiredSnippets) {
    assert.match(titles + body, new RegExp(snippet))
  }
  assert.equal(LEGAL_DOCUMENT_VERSION, '1.1.0')
  assert.equal(LEGAL_UPDATED_DATE, '2026-07-04')
  assert.match(LEGAL_EFFECTIVE_DATE, /^\d{4}-\d{2}-\d{2}$/)
  assert.match(LEGAL_UPDATED_DATE, /^\d{4}-\d{2}-\d{2}$/)
})

test('privacy policy covers collection categories and rights', () => {
  const titles = PRIVACY_POLICY_SECTIONS.map((section) => section.title).join('\n')
  const body = PRIVACY_POLICY_SECTIONS.flatMap((section) => section.paragraphs).join('\n')
  const combined = titles + '\n' + body
  const requiredSnippets = [
    'openid/unionid',
    '手机号',
    '登录密码',
    '不使用短信验证码',
    '访客留言',
    'IP、User-Agent',
    '仅在本地处理、不上传',
    '未满 14 周岁',
    '注销不会自动删除',
    '在世成员'
  ]
  for (const snippet of requiredSnippets) {
    assert.match(combined, new RegExp(snippet.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
  assert.ok(COLLECTED_INFO_ITEMS.length >= 8)
  assert.ok(LOCAL_ONLY_INFO.length >= 1)
})

test('wechat privacy declarations reference existing source files', () => {
  for (const item of WECHAT_PRIVACY_DECLARATIONS) {
    const paths = item.usedInCode
      .split('、')
      .map((part) => part.replace(/（[^）]*）/g, '').trim())
    for (const rel of paths) {
      const full = join(miniappSrc, rel)
      assert.ok(existsSync(full), `missing code path for ${item.apiOrScene}: ${rel}`)
    }
  }
})

test('not collected list excludes capabilities used in code', () => {
  const profilePath = join(miniappSrc, 'pages', 'me', 'profile.vue')
  const profileSource = readFileSync(profilePath, 'utf8')
  assert.match(profileSource, /chooseAvatar/)
  assert.ok(NOT_COLLECTED_IN_APP.some((item) => item.includes('getPhoneNumber')))
})

test('kinship page documents local-only processing', () => {
  const kinshipPath = join(miniappSrc, 'pages', 'tools', 'kinship.vue')
  const source = readFileSync(kinshipPath, 'utf8')
  assert.match(source, /不上传|不上送|本地/)
})

test('app mounts privacy consent modal and triggers prompt on launch/show', () => {
  const appPath = join(miniappSrc, 'App.vue')
  const source = readFileSync(appPath, 'utf8')
  assert.match(source, /PrivacyConsentModal/)
  assert.match(source, /onLaunch/)
  assert.match(source, /onShow/)
  assert.match(source, /promptPrivacyConsentIfNeeded/)
})

test('manifest enables privacy check without invalid chooseAvatar private-info declaration', () => {
  const manifestPath = join(miniappSrc, 'manifest.json')
  const manifest = JSON.parse(readFileSync(manifestPath, 'utf8')) as {
    'mp-weixin'?: {
      __usePrivacyCheck__?: boolean
      requiredPrivateInfos?: string[]
    }
  }
  assert.equal(manifest['mp-weixin']?.__usePrivacyCheck__, true)
  assert.ok(!manifest['mp-weixin']?.requiredPrivateInfos?.includes('chooseAvatar'))
})
