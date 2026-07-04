import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(here, '..', '..', '..')
const miniappSrc = join(repoRoot, 'miniapp', 'src')
const backendApp = join(repoRoot, 'backend', 'internal', 'app')
const backendMask = join(repoRoot, 'backend', 'internal', 'common', 'mask')
const backendFamilyCore = join(repoRoot, 'backend', 'internal', 'family', 'core')
const backendTree = join(repoRoot, 'backend', 'internal', 'family', 'tree')
const backendMigrations = join(repoRoot, 'backend', 'migrations')
const adminSrc = join(repoRoot, 'admin-web', 'src')

function read(path: string): string {
  return readFileSync(path, 'utf8')
}

test('pages.json includes showcase, public profile, public tree and content pages', () => {
  const pagesJson = JSON.parse(read(join(miniappSrc, 'pages.json'))) as {
    pages: Array<{ path: string }>
  }
  const paths = pagesJson.pages.map((page) => page.path)
  assert.equal(paths.includes('pages/family/search'), false)
  assert.ok(paths.includes('pages/family/showcase'))
  assert.ok(paths.includes('pages/family/public-profile'))
  assert.ok(paths.includes('pages/family/public-tree'))
  assert.ok(paths.includes('pages/content/index'))
  assert.ok(paths.includes('pages/content/detail'))
})

test('tab bar is home, showcase, reading and profile', () => {
  const pagesJson = JSON.parse(read(join(miniappSrc, 'pages.json'))) as {
    tabBar: { list: Array<{ pagePath: string; text: string }> }
  }
  const tabs = pagesJson.tabBar.list.map((item) => item.text)
  assert.deepEqual(tabs, ['首页', '展示家庭', '阅读', '我的'])
  const tabPaths = pagesJson.tabBar.list.map((item) => item.pagePath)
  assert.deepEqual(tabPaths, [
    'pages/home/index',
    'pages/family/showcase',
    'pages/content/index',
    'pages/me/index'
  ])
})

test('showcase page has no search box, filters or search button', () => {
  const showcase = read(join(miniappSrc, 'pages', 'family', 'showcase.vue'))
  const forbidden = [
    'confirm-type="search"',
    'query.keyword',
    'familySurname?.trim',
    'regionText?.trim',
    'buildPublicFamiliesQuery',
    '搜索',
    '筛选',
    '寻找家庭',
    'listPublicFamilies'
  ]
  for (const snippet of forbidden) {
    assert.equal(showcase.includes(snippet), false, `showcase must not include: ${snippet}`)
  }
  assert.match(showcase, /listPublicFamilyShowcase/)
  assert.match(showcase, /展示家庭/)
  assert.match(showcase, /public-profile/)
})

test('home page has showcase section without search wording', () => {
  const home = read(join(miniappSrc, 'pages', 'home', 'index.vue'))
  const forbidden = [
    'pages/family/search',
    '寻找家庭',
    '浏览公开主页',
    '搜索'
  ]
  for (const snippet of forbidden) {
    assert.equal(home.includes(snippet), false, `home must not include: ${snippet}`)
  }
  assert.match(home, /展示家庭/)
  assert.match(home, /查看更多展示家庭/)
  assert.match(home, /listPublicFamilyShowcase/)
  assert.match(home, /阅读精选/)
})

test('public profile page has no join apply entry', () => {
  const publicProfile = read(join(miniappSrc, 'pages', 'family', 'public-profile.vue'))
  const forbidden = ['申请加入家庭', 'openJoinApply', 'pages/join/apply']
  for (const snippet of forbidden) {
    assert.equal(publicProfile.includes(snippet), false, `public-profile must not include: ${snippet}`)
  }
})

test('client API uses listPublicFamilyShowcase and public families endpoint', () => {
  const familiesApi = read(join(miniappSrc, 'api', 'families.ts'))
  assert.match(familiesApi, /listPublicFamilyShowcase/)
  assert.match(familiesApi, /\/public\/families/)
  assert.equal(familiesApi.includes('listPublicFamilies'), false)
  assert.equal(existsSync(join(miniappSrc, 'pages', 'family', 'search.vue')), false)
})

test('backend registers showcase list route instead of offline search handler', () => {
  const router = read(join(backendApp, 'router.go'))
  assert.match(router, /api\.GET\("\/public\/families", handler\.ListPublicFamilyShowcase\)/)
  assert.equal(router.includes('PublicFamiliesSearchOffline'), false)
  const handler = read(join(backendFamilyCore, 'handler', 'family_handler.go'))
  assert.match(handler, /ListPublicFamilyShowcase/)
  assert.equal(handler.includes('PublicFamiliesSearchOffline'), false)
})

test('showcase service ignores search filters and omits contact fields in tests', () => {
  const testSource = read(join(backendFamilyCore, 'service', 'family_service_public_search_test.go'))
  assert.match(testSource, /TestListPublicFamilyShowcaseIgnoresSearchFilters/)
  assert.match(testSource, /TestListPublicFamilyShowcaseOmitsContactFields/)
  const service = read(join(backendFamilyCore, 'service', 'family_service.go'))
  assert.match(service, /ListPublicFamilyShowcase/)
  assert.match(service, /publicFamilyShowcaseItem/)
})

test('public family detail, public tree and content routes remain registered', () => {
  const router = read(join(backendApp, 'router.go'))
  const required = [
    'api.GET("/families/:familyId/public", handler.PublicDetail)',
    'api.GET("/public/families/:familyId/tree", treeHandler.PublicTree)',
    'content.GET("/categories", handler.PublicCategories)',
    'content.GET("/articles", handler.PublicArticles)',
    'content.GET("/articles/:articleId", handler.PublicArticleDetail)'
  ]
  for (const snippet of required) {
    assert.match(router, new RegExp(snippet.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
})

test('public detail and tree masking remains in place', () => {
  const familyMask = read(join(backendFamilyCore, 'vo', 'public_mask.go'))
  const treeMask = read(join(backendTree, 'service', 'public_mask.go'))
  const treeService = read(join(backendTree, 'service', 'tree_service.go'))
  const maskTest = read(join(backendMask, 'public_test.go'))
  const treeTest = read(join(backendTree, 'service', 'tree_service_test.go'))
  assert.match(familyMask, /mask\.DisplayName/)
  assert.match(treeMask, /GenerationCharacter = nil/)
  assert.match(treeMask, /RelationNote = nil/)
  assert.match(treeService, /maskPublicTreeResult\(result\), nil/)
  assert.match(maskTest, /张建国.*张\*国/)
  assert.match(treeTest, /TestPublicThenPrivateTreeUsesUnmaskedCache/)
})

test('legal text allows showcase browsing and forbids public search wording', () => {
  const legal = read(join(miniappSrc, 'features', 'legal', 'legalContent.ts'))
  assert.match(legal, /展示家庭/)
  assert.match(legal, /脱敏/)
  const forbidden = ['搜索公开家庭', '寻找家庭', '浏览公开家庭列表']
  for (const phrase of forbidden) {
    assert.equal(legal.includes(phrase), false, `legal content must not include: ${phrase}`)
  }
  const meta = read(join(miniappSrc, 'features', 'legal', 'legalMeta.ts'))
  assert.match(meta, /1\.4\.0/)
})

test('historical search service code and migrations remain preserved', () => {
  const migrationPath = join(backendMigrations, '000010_create_public_applications_and_messages.up.sql')
  assert.ok(existsSync(migrationPath))
  assert.match(read(join(backendFamilyCore, 'service', 'family_service.go')), /ListPublicFamilies/)
  assert.ok(existsSync(join(backendFamilyCore, 'dto', 'public_search.go')))
})

test('admin content center route remains available', () => {
  const adminRouter = read(join(adminSrc, 'router', 'index.ts'))
  assert.match(adminRouter, /content/)
})
