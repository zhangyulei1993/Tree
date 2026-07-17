import assert from 'node:assert/strict'
import { createRequire } from 'node:module'
import fs from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { after, before, test } from 'node:test'

import {
  buildCapabilityLines
} from './tree-quota-e2e-lib.mjs'
import {
  createE2EHarness,
  getApiOrigin,
  teardownSessionArtifact
} from './tree-quota-e2e-harness.mjs'

const require = createRequire(join(dirname(fileURLToPath(import.meta.url)), 'package.json'))
const harness = createE2EHarness()

before(async () => {
  if (process.env.TREE_E2E_SKIP === '1') return
  await harness.setup()
})

after(async () => {
  if (process.env.TREE_E2E_SKIP === '1') return
  await harness.teardown()
})

test('fresh quota session creates new user token and admin-aligned capabilities', { timeout: 120000 }, async () => {
  if (process.env.TREE_E2E_SKIP === '1') {
    return
  }

  const adminConfigs = await harness.fetchQuotaConfigs()
  const wechatConfig = adminConfigs.get('WECHAT_ONLY')
  assert.ok(wechatConfig, 'admin WECHAT_ONLY config missing')

  const { session, outFile } = await harness.createSession('fq-int')
  try {
    const stat = fs.statSync(outFile)
    assert.equal(stat.mode & 0o777, 0o600)
    assert.ok(session.user.id > 0, 'expected positive user id')
    assert.ok(session.token.length >= 20, 'expected non-empty jwt token')
    assert.equal(session.capabilities.trustTier, 'WECHAT_ONLY')
    assert.equal(session.capabilities.limits.maxOwnedFamilies, wechatConfig.maxOwnedFamilies)
    assert.equal(session.capabilities.limits.maxMembersPerOwnedFamily, wechatConfig.maxMembersPerOwnedFamily)
    assert.equal(session.capabilities.limits.maxJoinedFamilies, wechatConfig.maxJoinedFamilies)
    assert.equal(session.capabilities.limits.supportsGenerationNaming, wechatConfig.supportsGenerationNaming)
    assert.equal(session.capabilities.usage.ownedFamilies, 0)
    assert.equal(session.capabilities.usage.joinedFamilies, 0)
  } finally {
    await teardownSessionArtifact(session, outFile)
  }
})

test('account security page shows admin-driven quota lines in Playwright', { timeout: 240000 }, async () => {
  if (process.env.TREE_E2E_SKIP === '1') {
    return
  }

  await harness.forceBuildH5()
  const h5Base = await harness.startH5Preview()

  const adminConfigs = await harness.fetchQuotaConfigs()
  const wechatConfig = adminConfigs.get('WECHAT_ONLY')
  assert.ok(wechatConfig)

  const { session, outFile } = await harness.createSession('fq-pw')
  try {
    const expectedLines = buildCapabilityLines({
      trustTier: 'WECHAT_ONLY',
      limits: {
        maxOwnedFamilies: wechatConfig.maxOwnedFamilies,
        maxMembersPerOwnedFamily: wechatConfig.maxMembersPerOwnedFamily,
        maxJoinedFamilies: wechatConfig.maxJoinedFamilies,
        supportsGenerationNaming: wechatConfig.supportsGenerationNaming
      },
      usage: session.capabilities.usage
    })

    const { chromium } = require('playwright')
    const browser = await chromium.launch({ headless: true })
    try {
      const context = await browser.newContext()
      const apiOrigin = getApiOrigin()
      await context.route('**/api/**', async (route) => {
        const requestUrl = new URL(route.request().url())
        const proxied = `${apiOrigin}${requestUrl.pathname}${requestUrl.search}`
        const response = await route.fetch({ url: proxied })
        await route.fulfill({ response })
      })
      const page = await context.newPage()
      await page.addInitScript((initScript) => {
        // eslint-disable-next-line no-eval
        eval(initScript)
      }, session.playwrightInitScript)

      await page.goto(`${h5Base}/#/pages/me/account-security`, { waitUntil: 'networkidle' })
      await page.waitForSelector('.quota-line', { timeout: 20000 })

      const lines = await page.locator('.quota-line').allTextContents()
      assert.deepEqual(lines, expectedLines)
      assert.equal(lines[0], `可创建家庭：0/${wechatConfig.maxOwnedFamilies}`)
      assert.equal(lines[2], `每个家庭成员上限：${wechatConfig.maxMembersPerOwnedFamily}`)
    } finally {
      await browser.close()
    }
  } finally {
    await teardownSessionArtifact(session, outFile)
  }
})
