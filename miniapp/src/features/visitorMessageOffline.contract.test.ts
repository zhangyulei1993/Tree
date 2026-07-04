import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(here, '..', '..', '..')
const miniappSrc = join(repoRoot, 'miniapp', 'src')
const adminSrc = join(repoRoot, 'admin-web', 'src')
const backendApp = join(repoRoot, 'backend', 'internal', 'app')
const backendMigrations = join(repoRoot, 'backend', 'migrations')

function read(path: string): string {
  return readFileSync(path, 'utf8')
}

test('backend router does not register visitor message routes', () => {
  const routerSource = read(join(backendApp, 'router.go'))
  const forbidden = [
    'api.POST("/public/families/:familyId/visitor-messages"',
    'api.GET("/public/families/:familyId/visitor-messages"',
    'admin.GET("/visitor-messages"',
    'admin.POST("/visitor-messages/:messageId/approve"',
    'admin.POST("/visitor-messages/:messageId/reject"',
    'admin.DELETE("/visitor-messages/:messageId"',
    'messagehandler.NewHandler'
  ]
  for (const snippet of forbidden) {
    assert.equal(routerSource.includes(snippet), false, `router.go must not include: ${snippet}`)
  }
})

test('miniapp public family page has no visitor message UI or API', () => {
  const publicProfile = read(join(miniappSrc, 'pages', 'family', 'public-profile.vue'))
  const forbidden = [
    '我要留言',
    '提交留言',
    '公开留言',
    'visitorMessages',
    'listPublicVisitorMessages',
    'createVisitorMessage'
  ]
  for (const snippet of forbidden) {
    assert.equal(publicProfile.includes(snippet), false, `public-profile.vue must not include: ${snippet}`)
  }
  assert.equal(existsSync(join(miniappSrc, 'api', 'visitorMessages.ts')), false)
})

test('invite detail still shows inviter message block', () => {
  const inviteDetail = read(join(miniappSrc, 'pages', 'invite', 'detail.vue'))
  assert.match(inviteDetail, /邀请人留言/)
})

test('admin web removes visitor message route, sidebar and dashboard card', () => {
  const router = read(join(adminSrc, 'router', 'index.ts'))
  const sidebar = read(join(adminSrc, 'components', 'AdminSidebar.vue'))
  const dashboard = read(join(adminSrc, 'views', 'Dashboard.vue'))
  const forbidden = [
    'visitor-messages',
    'VisitorMessages',
    '游客留言',
    'pendingVisitorMessages'
  ]
  for (const snippet of forbidden) {
    assert.equal(router.includes(snippet), false, `admin router must not include: ${snippet}`)
    assert.equal(sidebar.includes(snippet), false, `admin sidebar must not include: ${snippet}`)
    assert.equal(dashboard.includes(snippet), false, `admin dashboard must not include: ${snippet}`)
  }
  assert.equal(existsSync(join(adminSrc, 'views', 'VisitorMessages.vue')), false)
  assert.equal(existsSync(join(adminSrc, 'api', 'visitorMessages.ts')), false)
})

test('visitor_messages table migration is preserved', () => {
  const migrationPath = join(backendMigrations, '000010_create_public_applications_and_messages.up.sql')
  assert.ok(existsSync(migrationPath), 'migration 000010 must remain')
  const migration = read(migrationPath)
  assert.match(migration, /CREATE TABLE visitor_messages/)
})
