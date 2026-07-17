import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'url'
import { test } from 'node:test'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const adminView = readFileSync(join(root, 'admin-web/src/views/AccountQuotaConfigs.vue'), 'utf8')
const adminApi = readFileSync(join(root, 'admin-web/src/api/accountQuota.ts'), 'utf8')

test('admin quota page loads configs from backend on mount', () => {
  assert.match(adminView, /listAccountQuotaConfigs/)
  assert.match(adminView, /loadConfigs\(\)/)
  assert.match(adminApi, /\/admin\/account-quota-configs/)
})

test('admin quota page saves via PUT and re-reads saved values', () => {
  assert.match(adminView, /updateAccountQuotaConfig/)
  assert.match(adminView, /Object\.assign\(item, snapshot\(saved\)\)/)
  assert.match(adminApi, /apiClient\.put\(`\/admin\/account-quota-configs\/\$\{tier\}`/)
})

test('admin quota page previews impact before save', () => {
  assert.match(adminView, /previewAccountQuotaImpact/)
  assert.match(adminApi, /\/account-quota-configs\/\$\{tier\}\/impact-preview/)
})

test('admin quota page confirms lowering limits', () => {
  assert.match(adminView, /isLowering/)
  assert.match(adminView, /确认降低额度/)
})

test('admin quota page manages feature preview privileged users', () => {
  assert.match(adminView, /特色功能特权体验/)
  assert.match(adminView, /FEATURE_PREVIEW/)
  assert.match(adminView, /listAccountFeatureOverrides/)
  assert.match(adminView, /updateAccountFeatureOverrides/)
  assert.match(adminView, /removeFeaturePreviewOverride/)
  assert.match(adminView, /closable/)
  assert.match(adminView, /clearFeaturePreviewOverrides/)
  assert.match(adminView, /清空名单/)
  assert.match(adminApi, /\/admin\/account-feature-overrides\/\$\{featureKey\}/)
  assert.match(adminApi, /apiClient\.delete\(`\/admin\/account-feature-overrides\/\$\{featureKey\}\/\$\{overrideId\}`/)
})

test('admin sidebar restricts quota config to root and super admins', () => {
  const sidebar = readFileSync(join(root, 'admin-web/src/components/AdminSidebar.vue'), 'utf8')
  assert.match(sidebar, /account-quota-configs/)
  assert.match(sidebar, /账号权益配置.*roles: \['ROOT_ADMIN', 'SUPER_ADMIN'\]/)
})
