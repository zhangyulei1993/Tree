import assert from 'node:assert/strict'
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
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

function walkRuntimeFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name)
    const stat = statSync(path)
    if (stat.isDirectory()) return walkRuntimeFiles(path)
    if (!/\.(vue|ts|js)$/.test(name)) return []
    if (/\.(test|contract|smoke)\./.test(name)) return []
    return [path]
  })
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

test('tab bar is Tree, my family, reading and profile', () => {
  const pagesJson = JSON.parse(read(join(miniappSrc, 'pages.json'))) as {
    tabBar: { list: Array<{ pagePath: string; text: string }> }
  }
  const tabs = pagesJson.tabBar.list.map((item) => item.text)
  assert.deepEqual(tabs, ['Tree', '我的家庭', '阅读', '我的'])
  const tabPaths = pagesJson.tabBar.list.map((item) => item.pagePath)
  assert.deepEqual(tabPaths, [
    'pages/home/index',
    'pages/family/my',
    'pages/content/index',
    'pages/me/index'
  ])
})

test('runtime code does not navigateTo or redirectTo tabBar pages', () => {
  const pagesJson = JSON.parse(read(join(miniappSrc, 'pages.json'))) as {
    tabBar: { list: Array<{ pagePath: string }> }
  }
  const tabPaths = pagesJson.tabBar.list.map((item) => `/${item.pagePath}`)
  const runtimeFiles = walkRuntimeFiles(miniappSrc)
  const violations: string[] = []

  for (const file of runtimeFiles) {
    const source = read(file)
    for (const tabPath of tabPaths) {
      const escapedPath = tabPath.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
      const invalidNavigation = new RegExp(
        `uni\\.(?:navigateTo|redirectTo)\\s*\\(\\s*\\{\\s*url\\s*:\\s*['"\`]${escapedPath}(?:[?#][^'"\`]*)?['"\`]`
      )
      if (invalidNavigation.test(source)) {
        violations.push(`${file.replace(`${repoRoot}/`, '')} -> ${tabPath}`)
      }
    }
  }

  assert.deepEqual(violations, [])
})

test('notification routes switch to tabBar pages and navigate to detail pages', () => {
  const source = read(join(miniappSrc, 'pages', 'me', 'notifications.vue'))
  assert.match(source, /const tabBarRoutes = new Set\(\['\/pages\/home\/index', '\/pages\/family\/my', '\/pages\/content\/index', '\/pages\/me\/index'\]\)/)
  assert.match(source, /if \(tabBarRoutes\.has\(basePath\)\) \{\s*uni\.switchTab\(\{ url: basePath, success: onSuccess/)
  assert.match(source, /uni\.navigateTo\(\{ url: normalized, success: onSuccess/)
  assert.match(source, /openRoute\(item\.route, \(\) => \{ void markRead\(item\) \}\)/)
  assert.match(source, /while \(filter\.value === 'UNREAD' && !visibleItems\.value\.length && hasMore\.value && !errorMessage\.value\)/)
})

test('private tree blocks member additions when capacity cannot be confirmed', () => {
  const source = read(join(miniappSrc, 'pages', 'family', 'private-tree.vue'))
  assert.match(source, /title: '容量信息暂不可用'/)
  assert.match(source, /暂时无法确认当前家庭容量，请稍后重试。/)
  assert.match(source, /return false\s*\}\s*const usage = capabilities\.usage\.membersPerOwnedFamily/)
})

test('real-api tool config failure fails closed instead of using stale enabled defaults', () => {
  const source = read(join(miniappSrc, 'api', 'toolConfigs.ts'))
  assert.match(source, /function disableToolsAfterConfigFailure\(\)/)
  assert.match(source, /state\.visible = false\s*state\.enabled = false/)
  assert.match(source, /disableToolsAfterConfigFailure\(\)/)
})

test('MiniButton stops native tap bubbling instead of component click modifiers', () => {
  const buttonSource = read(join(miniappSrc, 'components', 'base', 'MiniButton.vue'))
  assert.match(buttonSource, /@tap\.stop="handleTap"/)

  const runtimeFiles = walkRuntimeFiles(miniappSrc)
  const violations: string[] = []

  for (const file of runtimeFiles) {
    const source = read(file)
    if (/<MiniButton[\s\S]*?@click\.stop=/.test(source)) {
      violations.push(file.replace(`${repoRoot}/`, ''))
    }
  }

  assert.deepEqual(violations, [])
})

test('family management pages guard missing familyId before API loading', () => {
  const requiredGuards = [
    {
      file: join(miniappSrc, 'pages', 'family', 'manage.vue'),
      routeSnippet: 'const route = `/pages/family/manage',
      apiSnippet: 'getFamilyDetail(familyId.value)'
    },
    {
      file: join(miniappSrc, 'pages', 'family', 'settings.vue'),
      routeSnippet: 'const route = `/pages/family/settings',
      apiSnippet: 'getFamilyDetail(familyId.value)'
    }
  ]

  for (const item of requiredGuards) {
    const source = read(item.file)
    const guardIndex = source.indexOf('if (!familyId.value)')
    const routeIndex = source.indexOf(item.routeSnippet)
    const apiIndex = source.indexOf(item.apiSnippet)
    assert.notEqual(guardIndex, -1)
    assert.ok(guardIndex < routeIndex)
    assert.ok(guardIndex < apiIndex)
  }
})

test('private tree loads family invitations only for managers', () => {
  const privateTree = read(join(miniappSrc, 'pages', 'family', 'private-tree.vue'))
  const familyIndex = privateTree.indexOf('const familyResult = await getFamilyDetail(familyId.value)')
  const canManageIndex = privateTree.indexOf("const canManage = familyResult.role === 'FOUNDER' || familyResult.role === 'FAMILY_ADMIN'")
  const invitationIndex = privateTree.indexOf('canManage ? listFamilyInvitations(familyId.value) : Promise.resolve([])')

  assert.notEqual(familyIndex, -1)
  assert.notEqual(canManageIndex, -1)
  assert.notEqual(invitationIndex, -1)
  assert.ok(familyIndex < canManageIndex)
  assert.ok(canManageIndex < invitationIndex)
})

test('all family members can switch tree perspective without gaining management actions', () => {
  const privateTree = read(join(miniappSrc, 'pages', 'family', 'private-tree.vue'))
  assert.match(privateTree, /:interactive="true"/)
  assert.match(privateTree, /actions\.push\(\{ label: '查看资料', primary: true/)
  assert.match(privateTree, /if \(canSwitchViewer\(node\)\) actions\.push\(\{ label: '切换到此视角'/)
  assert.match(privateTree, /if \(canManageFamily\.value\) \{[\s\S]*label: '邀请本人绑定'/)
})

test('private tree gates deceased memorial by family gray access', () => {
  const privateTree = read(join(miniappSrc, 'pages', 'family', 'private-tree.vue'))
  assert.match(privateTree, /:selectable-deceased="grayAccessAvailable"/)
  assert.match(privateTree, /grayAccessAvailable\.value = Boolean\(treeResult\.grayAccessEnabled\)/)
  assert.match(privateTree, /if \(node\.isLiving === false && grayAccessAvailable\.value\) \{[\s\S]*showMemorial\(node\)[\s\S]*return/)
  assert.match(privateTree, /class="memorial-action secondary" @click="switchMemorialViewer"/)
  assert.match(privateTree, /class="memorial-action primary" @click="viewMemorialMember"/)
  assert.match(privateTree, /v-if="canDeleteMemorialMember" class="memorial-action danger[^"]*" @click="deleteMemorialMember"/)
  assert.match(privateTree, /v-else-if="memorialDeleteRestriction" class="memorial-delete-restriction"/)
  assert.match(privateTree, /getFamilyMember\(familyId\.value, node\.memberId\)/)
  assert.match(privateTree, /亲人简介/)
})

test('private tree forwards deceased selection through located and unlocated cards', () => {
  const treeGraph = read(join(miniappSrc, 'components', 'family', 'FamilyTreeGraph.vue'))
  assert.match(treeGraph, /:selectable="isNodeSelectable\(parent\)"/)
  assert.match(treeGraph, /:selectable="isUnlocatedNodeSelectable\(node\)"/)
})

test('private tree deceased node panel keeps all actions accessible', () => {
  const privateTree = read(join(miniappSrc, 'pages', 'family', 'private-tree.vue'))
  const treeGraph = read(join(miniappSrc, 'components', 'family', 'FamilyTreeGraph.vue'))
  assert.equal(privateTree.includes('编辑亲人简介'), false)
  assert.equal(privateTree.includes("'查看亲人简介'"), false)
  assert.equal(privateTree.includes('<picker :range="viewerPickerLabels"'), false)
  assert.match(privateTree, /selectable-all/)
  assert.match(treeGraph, /props\.selectableAll \|\| props\.interactive \|\| \(props\.selectableDeceased && node\.isLiving === false\)/)
  assert.match(treeGraph, /function isUnlocatedNodeSelectable\(node: FamilyTreeResult\['nodes'\]\[number\]\)/)
  assert.match(treeGraph, /props\.interactive \|\| \(props\.selectableDeceased && node\.isLiving === false\)/)
  assert.match(privateTree, /function canSwitchViewer\(node: TreeNode\)/)
  assert.match(privateTree, /class="memorial-action secondary" @click="switchMemorialViewer"/)
  assert.match(privateTree, /class="memorial-action primary" @click="viewMemorialMember"/)
  assert.match(privateTree, /v-if="canDeleteMemorialMember" class="memorial-action danger/)
  assert.match(privateTree, /@click="deleteMemorialMember"/)
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
  assert.match(showcase, /优质家庭/)
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
  assert.match(home, /帝王世家/)
  assert.match(home, /查看全部/)
  assert.match(home, /listPublishedShowcaseFamilies/)
  assert.doesNotMatch(home, /阅读精选/)
  assert.match(home, /创建我的家庭/)
  assert.match(home, /openHomePrimaryEntry/)
  assert.match(home, /pages\/family\/create/)
})

test('my family tab keeps create entry when showing public showcase', () => {
  const myFamily = read(join(miniappSrc, 'pages', 'family', 'my.vue'))
  assert.match(myFamily, /优质家庭/)
  assert.match(myFamily, /创建家庭/)
  assert.match(myFamily, /openCreateFamily/)
  assert.match(myFamily, /pages\/family\/create/)
})

test('public profile page has no join apply entry', () => {
  const publicProfile = read(join(miniappSrc, 'pages', 'family', 'public-profile.vue'))
  const forbidden = ['申请加入家庭', 'openJoinApply', 'pages/join/apply']
  for (const snippet of forbidden) {
    assert.equal(publicProfile.includes(snippet), false, `public-profile must not include: ${snippet}`)
  }
})

test('family settings gates quality-family application by eligibility', () => {
  const settings = read(join(miniappSrc, 'pages', 'family', 'settings.vue'))
  const profile = read(join(miniappSrc, 'pages', 'family', 'profile.vue'))
  const detail = read(join(miniappSrc, 'pages', 'family', 'detail.vue'))
  const manageCenter = read(join(miniappSrc, 'pages', 'family', 'manage-center.vue'))
  assert.match(settings, /getQualityEligibility/)
  assert.match(settings, /const customerPublicDisplayAvailable = computed\(\(\) => Boolean\(qualityEligibility\.value\?\.eligible\)\)/)
  assert.match(settings, /const settingsSection = ref<[^>]+>\('roles'\)/)
  assert.match(settings, /v-if="publicSettingsVisible"[\s\S]*?<text>展示<\/text>/)
  assert.match(settings, /visibleOperationLogs/)
  const forbidden = ['暂缓开放', '普通家庭公开展示暂缓开放', '公开展示摘要', '公开状态与联系方式', '公开信息、权限角色', '公开展示会重置', '公开审核', '搜索开关']
  for (const source of [settings, profile, detail, manageCenter]) {
    for (const phrase of forbidden) {
      assert.equal(source.includes(phrase), false, `family management copy must not include: ${phrase}`)
    }
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
  assert.equal(router.includes('families.POST("/:familyId/join-requests", joinHandler.Create)'), false)
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
  const catalog = read(join(miniappSrc, 'features', 'legal', 'privacyCatalog.ts'))
  assert.match(legal, /历史世系示例/)
  assert.match(catalog, /历史世系示例与分享浏览/)
  const forbidden = ['搜索公开家庭', '寻找家庭', '浏览公开家庭列表', '公开家庭展示申请', '申请公开展示', '家庭管理员申请并经平台审核']
  for (const phrase of forbidden) {
    assert.equal(legal.includes(phrase), false, `legal content must not include: ${phrase}`)
    assert.equal(catalog.includes(phrase), false, `privacy catalog must not include: ${phrase}`)
  }
  const meta = read(join(miniappSrc, 'features', 'legal', 'legalMeta.ts'))
  assert.match(meta, /1\.5\.0/)
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
