import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

test('new user guide starts with a sample and uses WeChat login as the next action', () => {
  const home = readFileSync(join(miniappSrc, 'pages', 'home', 'index.vue'), 'utf8')

  assert.match(home, /key: 'SHOWCASE', label: '看示例'/)
  assert.match(home, /key: 'LOGIN', label: '微信登录'/)
  assert.match(home, /key: 'CREATE', label: '创建家谱'/)
  assert.match(home, /先用一棵示例树熟悉家谱/)
})

test('guide routes creators and ordinary members to different next actions', () => {
  const home = readFileSync(join(miniappSrc, 'pages', 'home', 'index.vue'), 'utf8')

  assert.match(home, /guideFamilyCanManage/)
  assert.match(home, /补充成员/)
  assert.match(home, /连好关系/)
  assert.match(home, /看看家庭树/)
  assert.match(home, /帮家庭助力/)
  assert.match(home, /nextJourneyStepKey/)
  assert.match(home, /journeyRequiredMilestones/)
})

test('restarting the guide stays on the home page until the user chooses the first step', () => {
  const home = readFileSync(join(miniappSrc, 'pages', 'home', 'index.vue'), 'utf8')
  const restart = home.slice(home.indexOf('function restartNewUserGuide()'), home.indexOf('function syncNewUserGuideProgress()'))

  assert.match(restart, /resetNewUserGuideProgress\(\)/)
  assert.match(restart, /已重置，请从第一步开始/)
  assert.doesNotMatch(restart, /openHistoricalShowcase/)
})

test('the invitation guide requires phone binding and completes only after an invitation is created', () => {
  const home = readFileSync(join(miniappSrc, 'pages', 'home', 'index.vue'), 'utf8')
  const invite = readFileSync(join(miniappSrc, 'pages', 'invite', 'sent.vue'), 'utf8')

  assert.match(home, /&guide=1/)
  assert.match(invite, /最后一步：完成邀请/)
  assert.match(invite, /绑定手机号后即可发起邀请/)
  assert.match(invite, /:disabled="showPhoneBindGate \|\| Boolean\(actingId\) \|\| actingType === 'create-member'"/)
  assert.match(invite, /markNewUserGuideMilestone\('INVITE_GUIDE_OPENED'\)/)
})

test('guide can be deferred and resumed without resetting completed steps', () => {
  const home = readFileSync(join(miniappSrc, 'pages', 'home', 'index.vue'), 'utf8')
  const guide = readFileSync(join(miniappSrc, 'features', 'onboarding', 'newUserGuide.ts'), 'utf8')
  const me = readFileSync(join(miniappSrc, 'pages', 'me', 'index.vue'), 'utf8')

  assert.match(home, /dismissNewUserGuide\(3\)/)
  assert.match(guide, /function resumeNewUserGuide/)
  assert.match(me, /新手引导/)
  assert.match(me, /resumeNewUserGuide\(\)/)
})
