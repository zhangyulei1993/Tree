import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

function source(relativePath: string) {
  return readFileSync(join(miniappSrc, relativePath), 'utf8')
}

test('tree cards expose relation-specific actions and no standalone tree editor entry', () => {
  const tree = source('pages/family/private-tree.vue')

  assert.match(tree, /availableRelationAdditions/)
  assert.match(tree, /if \(!hasFather\) additions\.push\(\{ label: '添加父亲'/)
  assert.match(tree, /if \(parents\.length > 0\) additions\.push\(\{ label: '添加兄弟姐妹'/)
  assert.doesNotMatch(tree, /function openManage\(/)
  assert.doesNotMatch(tree, /<view v-if="canManageFamily" class="genealogy-tools">/)
})

test('member quota differentiates phone upgrade from a final bound-phone limit', () => {
  const tree = source('pages/family/private-tree.vue')
  const manage = source('pages/family/manage.vue')
  const dialog = source('components/base/MiniPhoneUpgradeDialog.vue')

  assert.match(tree, /!session\.user\?\.phoneVerified/)
  assert.match(tree, /phoneUpgradeVisible\.value = true/)
  assert.match(tree, /购买扩容/)
  assert.match(tree, /const page = result\.tapIndex === 0 \? 'capacity' : 'invitation-reward'/)
  assert.match(tree, /pages\/family\/\$\{page\}\?familyId=/)
  assert.match(manage, /isMemberLimitError/)
  assert.match(dialog, /open-type="getPhoneNumber"/)
  assert.match(dialog, /绑定手机号可升级权益/)
})

test('temporary member creation and placement are absent from the management surface', () => {
  const manage = source('pages/family/manage.vue')

  assert.doesNotMatch(manage, /添加暂存|接入暂存|暂存成员|submitUnlocated|submitPlacement|placeExistingMember/)
})

test('family tree is the single member operation surface', () => {
  const tree = source('pages/family/private-tree.vue')
  const memberView = source('pages/family/member-view.vue')

  assert.match(tree, /function openNodeActions\(node: TreeNode\)/)
  assert.match(tree, /const selectedNode = ref<TreeNode \| null>\(null\)/)
  assert.match(tree, /class="node-action-panel archive-form-panel"/)
  assert.match(tree, /position: fixed;/)
  assert.match(tree, /label: '查看资料', primary: true/)
  assert.match(tree, /label: '邀请本人绑定', primary: true/)
  assert.doesNotMatch(tree, /function openMembers\(/)
  assert.match(tree, /requestedViewerMemberId/)
  assert.match(tree, /member-view\?familyId=.*&returnTo=tree/)
  assert.match(memberView, /returnTo\.value === 'tree' \? '返回家庭树' : returnTo\.value === 'health' \? '返回家庭健康' : '返回家庭详情'/)
  assert.match(memberView, /pages\/family\/private-tree\?familyId=/)
})

test('tree keeps the family context while handling member and family affairs', () => {
  const tree = source('pages/family/private-tree.vue')
  const invitation = source('pages/invite/sent.vue')
  const manage = source('pages/family/manage.vue')
  const center = source('pages/family/manage-center.vue')
  const health = source('pages/family/health.vue')

  assert.match(tree, /class="tree-affairs archive-form-panel" @tap="openTreeAffairs"/)
  assert.match(tree, /function openTreeAffairs\(\)/)
  assert.match(tree, /&returnTo=tree/)
  assert.match(tree, /label: '补充资料', primary: true/)
  assert.match(invitation, /returnTo\.value === 'tree' \? '返回家庭树' : returnTo\.value === 'health' \? '返回家庭健康' : '返回家庭详情'/)
  assert.match(manage, /returnTo\.value === 'tree' \? '返回家庭树' : '返回家庭详情'/)
  assert.match(center, /const treeReturn = returnTo\.value === 'tree'/)
  assert.match(health, /returnTo\.value === 'tree' \? '返回家庭树' : '返回家庭详情'/)
})

test('explicit return targets reuse the preceding page instead of duplicating it in the stack', () => {
  const backHome = source('components/base/MiniBackHome.vue')

  assert.match(backHome, /const targetPath = props\.returnUrl\.split\('\?'\)\[0\]/)
  assert.match(backHome, /\.slice\(0, -1\)/)
  assert.match(backHome, /\.lastIndexOf\(targetPath\)/)
  assert.match(backHome, /uni\.navigateBack\(\{ delta: pages\.length - 1 - targetPageIndex \}\)/)
})

test('family detail keeps actions with their real operational surface', () => {
  const detail = source('pages/family/detail.vue')
  const health = source('pages/family/health.vue')
  const capacity = source('pages/family/capacity.vue')

  assert.match(detail, /编辑家庭资料/)
  assert.match(detail, /class="archive-row" @click="openFamilyTree"/)
  assert.match(detail, /查看成员、编辑关系和邀请本人绑定/)
  assert.match(detail, /家庭健康与档案导出/)
  assert.match(detail, /成员名额与订单/)
  assert.doesNotMatch(detail, /我的成员资料|家庭档案<\/text>|openSentInvitations|openCapacityOrders/)
  assert.match(health, /const exportScoreThreshold = 90/)
  assert.match(health, /完成健康检查后可导出/)
  assert.match(capacity, /成员名额与订单/)
  assert.doesNotMatch(detail, /detail-quick-actions|常用操作/)
  assert.match(detail, /openRecovery/)
  assert.match(detail, /最近删除可恢复/)
})

test('family management keeps settings in one page instead of splitting identical destinations', () => {
  const center = source('pages/family/manage-center.vue')
  const detail = source('pages/family/detail.vue')

  assert.match(center, /家庭设置/)
  assert.match(center, /展示、成员权限提醒、创建者转让、解散与操作记录/)
  assert.match(center, /function openSettings\(\)/)
  assert.doesNotMatch(center, /openSettings\('roles'\)|openSettings\('transfer'\)|openSettings\('dissolution'\)/)
  assert.doesNotMatch(center, /function openOperationLogs\(/)
  assert.match(detail, /navigateOnce\(`\/pages\/family\/manage-center\?familyId=/)
  assert.doesNotMatch(center, /function openSentInvitations\(/)
  assert.match(center, /完善家庭资料与关系/)
  assert.match(center, /继续支付 \$\{item\.productName\}/)
  assert.match(center, /继续邀请 \$\{item\.targetMemberName\}/)
  assert.match(center, /pages\/invite\/sent\?familyId=.*&mode=node/)
  assert.doesNotMatch(center, /function openRecovery\(/)
})

test('family management presents an empty task list as a completed state, not a blank page', () => {
  const center = source('pages/family/manage-center.vue')

  assert.match(center, /class="manage-clear-state"/)
  assert.match(center, /事项已处理/)
  assert.match(center, /class="manage-directory-head"/)
  assert.match(center, /需要时可查看记录或调整家庭设置/)
  assert.doesNotMatch(center, /<MiniEmptyState v-else symbol="好" title="当前没有待办"/)
})

test('family health presents a clear completion result instead of an oversized empty state', () => {
  const health = source('pages/family/health.vue')

  assert.match(health, /class="health-clear-state"/)
  assert.match(health, /基础检查已通过/)
  assert.match(health, /当前没有需要优先补充的成员资料或关系问题。/)
  assert.doesNotMatch(health, /<MiniEmptyState v-else symbol="好" title="家庭资料很完整"/)
})

test('tree node relation additions are grouped under a secondary action menu', () => {
  const tree = source('pages/family/private-tree.vue')

  assert.match(tree, /type NodeActionMenu = 'MAIN' \| 'RELATION' \| 'MORE'/)
  assert.match(tree, /const nodeActionMenu = ref<NodeActionMenu>\('MAIN'\)/)
  assert.match(tree, /const selectedNodeRelationAdditions = computed\(\(\) => \{/)
  assert.match(tree, /label: '添加关系', primary: true, keepPanelOpen: true, run: \(\) => openNodeActionMenu\('RELATION'\)/)
  assert.match(tree, /v-else-if="nodeActionMenu === 'RELATION'"/)
  assert.match(tree, /v-else class="node-action-grid"/)
  assert.match(tree, /label: '更多操作', keepPanelOpen: true, run: \(\) => openNodeActionMenu\('MORE'\)/)
  assert.match(tree, /function openNodeActionMenu\(menu: NodeActionMenu\)/)
  assert.match(tree, /if \(!action\.keepPanelOpen\) closeNodeActions\(\)/)
  assert.match(tree, /function openSelectedNodeRelation\(addType: RelationshipAddType\)/)
  assert.doesNotMatch(tree, /actions\.push\(\.\.\.relationAdditions\.map/)
  assert.match(tree, /label: '调整已有关系'/)
  assert.doesNotMatch(tree, /添加亲属或调整关系/)
  assert.doesNotMatch(tree, /function openRelationActions/)
})

test('tree hides no-op or blocked destructive operations before the user taps them', () => {
  const tree = source('pages/family/private-tree.vue')

  assert.match(tree, /v-if="canResetViewerToSelf" class="viewer-action" @tap="resetViewerToSelf">回到我的视角/)
  assert.match(tree, /v-if="viewerMemberId !== null" class="viewer-action" @tap="showAllTree">查看全树/)
  assert.match(tree, /const canResetViewerToSelf = computed\(\(\) =>/)
  assert.match(tree, /function deleteRestriction\(node: TreeNode \| null \| undefined\)/)
  assert.match(tree, /该成员已有子女，不能删除。/)
  assert.match(tree, /当前创建人需先完成转让，不能删除。/)
  assert.match(tree, /if \(canDeleteTreeNode\(node\)\) actions\.push\(\{ label: '删除节点'/)
  assert.match(tree, /v-if="canDeleteMemorialMember" class="memorial-action danger memorial-delete-action"/)
  assert.match(tree, /v-else-if="memorialDeleteRestriction" class="memorial-delete-restriction"/)
  assert.match(tree, /当前创建人需先完成转让，不能删除。/)
  assert.match(tree, /该成员已有子女，不能删除。/)
})

test('self-maintained profiles do not require a redundant confirmation tap', () => {
  const memberView = source('pages/family/member-view.vue')

  assert.match(memberView, /const canReviewOtherMember = computed\(\(\) => canManageFamily\.value && !isCurrentUserMember\.value\)/)
  assert.match(memberView, /v-if="canReviewOtherMember" class="archive-form-panel verification-panel"/)
  assert.match(memberView, /确认资料无误/)
  assert.doesNotMatch(memberView, /确认是我的资料/)
})

test('member editing preserves whether the user came from the tree or family detail', () => {
  const memberEdit = source('pages/family/member-edit.vue')

  assert.match(memberEdit, /const memberViewUrl = computed\(\(\) => \{/)
  assert.match(memberEdit, /memberViewUrl\.value/)
  assert.match(memberEdit, /const shouldKeepReturnTarget = computed\(\(\) => \['tree', 'health'\]\.includes\(returnTo\.value\)\)/)
  assert.doesNotMatch(memberEdit, /member-view\?familyId=.*&returnTo=tree/)
})

test('health issue member flows preserve the health-check context through view and edit', () => {
  const health = source('pages/family/health.vue')
  const memberView = source('pages/family/member-view.vue')
  const memberEdit = source('pages/family/member-edit.vue')
  const privateTree = source('pages/family/private-tree.vue')
  const sentInvitation = source('pages/invite/sent.vue')

  assert.match(health, /member-view\?familyId=.*&returnTo=health/)
  assert.match(health, /private-tree\?familyId=.*&returnTo=health/)
  assert.match(health, /invite\/sent\?familyId=.*&returnTo=health/)
  assert.match(memberView, /returnTo\.value === 'health' \? '返回家庭健康'/)
  assert.match(memberView, /\['tree', 'health'\]\.includes\(returnTo\.value\)/)
  assert.match(memberEdit, /returnTo\.value === 'health' \? '返回家庭健康'/)
  assert.match(memberEdit, /const shouldKeepReturnTarget = computed\(\(\) => \['tree', 'health'\]\.includes\(returnTo\.value\)\)/)
  assert.match(privateTree, /returnTo\.value === 'health' \? '返回家庭健康' : '返回家庭详情'/)
  assert.match(sentInvitation, /returnTo\.value === 'health' \? '返回家庭健康' : '返回家庭详情'/)
})

test('pending invitations open their action page without a tree-page detour', () => {
  const center = source('pages/family/manage-center.vue')
  const health = source('pages/family/health.vue')
  const detail = source('pages/family/detail.vue')

  assert.match(center, /pages\/invite\/sent\?familyId=.*&mode=node/)
  assert.match(health, /pages\/invite\/sent\?familyId=.*&mode=node/)
  assert.match(detail, /pages\/invite\/sent\?familyId=.*&mode=node/)
})

test('personal and family management destinations have distinct names and notifications acknowledge missing destinations', () => {
  const personalAffairs = source('pages/me/family-affairs.vue')
  const familyAffairs = source('pages/family/manage-center.vue')
  const notifications = source('pages/me/notifications.vue')

  assert.match(personalAffairs, /我的邀请与申请/)
  assert.match(familyAffairs, /家庭管理/)
  assert.match(notifications, /暂无可打开的关联页面/)
})

test('notification pagination restores the current page when loading more fails', () => {
  const notifications = source('pages/me/notifications.vue')

  assert.match(notifications, /const previousPage = page\.value/)
  assert.match(notifications, /if \(!await load\(\)\) page\.value = previousPage/)
  assert.ok(
    notifications.indexOf('if (loading.value || loadingMore.value) return false') < notifications.indexOf('if (reset) { page.value = 1; items.value = [] }')
  )
})

test('every visible tree node can open the perspective action, including unlocated members', () => {
  const graph = source('components/family/FamilyTreeGraph.vue')
  const tree = source('pages/family/private-tree.vue')

  assert.match(graph, /props\.selectableAll \|\| props\.interactive/)
  assert.match(tree, /tree\.value\?\.nodes\.some\(\(item\) => item\.memberId === node\.memberId\)/)
  assert.match(tree, /label: '切换到此视角'/)
  assert.match(tree, /if \(node\.userBindingState === 'INVITING'\)/)
  assert.match(tree, /label: '继续邀请', primary: true/)
})

test('home tree entry opens the only member-management surface', () => {
  const home = source('pages/home/index.vue')
  const pages = source('pages.json')

  assert.ok(home.includes('`/pages/family/private-tree?familyId=${familyId}&guide=1`'))
  assert.ok(!pages.includes('"pages/family/members"'))
})

test('tree layout dimensions and tap targets match the portrait node footprint', () => {
  const layout = source('features/family-tree/layoutFamilyTree.ts')
  const graph = source('components/family/FamilyTreeGraph.vue')

  assert.match(layout, /const SINGLE_NODE_W = 160/)
  assert.match(layout, /const COUPLE_ROW_H = 178/)
  assert.match(graph, /class="tree-person-hit" @tap\.stop="handleNodeSelect\(parent\)"/)
})

test('tree removes legacy node side borders without removing real relationship lines', () => {
  const graph = source('components/family/FamilyTreeGraph.vue')
  const chip = source('components/family/FamilyTreeMemberChip.vue')
  const poster = source('features/family-tree/exportTreePoster.ts')

  assert.match(graph, /v-for="\(segment, index\) in link\.segments"/)
  assert.doesNotMatch(chip, /lineage-PATRILINEAL \{\s*border-left/)
  assert.doesNotMatch(chip, /lineage-NATAL_FEMALE \{\s*border-left/)
  assert.match(poster, /if \(segment\.orientation === 'h'\)/)
})

test('tree keeps active lineage nodes transparent instead of restoring legacy role cards', () => {
  const chip = source('components/family/FamilyTreeMemberChip.vue')

  assert.match(chip, /\.tree-member-chip\.lineage-tag\.lineage-PATRILINEAL,\s*\.tree-member-chip\.lineage-tag\.lineage-NATAL_FEMALE,/)
  assert.match(chip, /background: transparent;/)
})

test('export preview loads source data independently from export-record sync', () => {
  const exportPage = source('pages/family/export.vue')
  const treePage = source('pages/family/private-tree.vue')

  assert.match(exportPage, /const \[currentTree, currentFamily, currentMembers\] = await Promise\.all/)
  assert.match(exportPage, /await exportFamilyArchive\(familyId\.value, scope\.value\)/)
  assert.match(exportPage, /导出记录暂未同步，不影响图片使用/)
  assert.match(exportPage, /暂不能生成长卷/)
  assert.doesNotMatch(exportPage, /树形家谱图/)
  assert.match(treePage, /导出家庭树/)
  assert.match(treePage, /buildFamilyTreePosterPage\(/)
  assert.match(treePage, /canvas-id="privateTreePosterCanvas"/)
})

test('MiniButton consumers use the component click event instead of a native tap listener', () => {
  const health = source('pages/family/health.vue')
  const notifications = source('pages/me/notifications.vue')
  const capacity = source('pages/family/capacity.vue')

  assert.match(health, /<MiniButton :disabled="!canExport" @click="openExport">/)
  assert.doesNotMatch(health, /<MiniButton[^>]*@tap=/)
  assert.doesNotMatch(notifications, /<MiniButton[^>]*@tap=/)
  assert.doesNotMatch(capacity, /<MiniButton[^>]*@tap=/)
})

test('tree member cards keep component and footprint click paths with one selection guard', () => {
  const graph = source('components/family/FamilyTreeGraph.vue')

  assert.match(graph, /@tap\.stop="handleNodeSelect\(parent\)"/)
  assert.match(graph, /@select="handleNodeSelect"/)
  assert.match(graph, /const selectingMemberId = ref<number \| null>\(null\)/)
  assert.match(graph, /if \(selectingMemberId\.value === node\.memberId\) return/)
  assert.match(graph, /Promise\.resolve\(\)\.then\(/)
  assert.match(graph, /watch\(\(\) => props\.tree, \(\) => \{\s*selectingMemberId\.value = null/)
})

test('relationship preview remounts when the selected member or relation changes', () => {
  const manage = source('pages/family/manage.vue')

  assert.match(manage, /:key="`\$\{selectedBaseMember\?\.memberId \|\| 'member'\}-\$\{relationValues\[relationIndex\]\}`"/)
  assert.match(manage, /title="关系预览"/)
})

test('pending invitations distinguish a valid invitation from a locally shareable link', () => {
  const sent = source('pages/invite/sent.vue')

  assert.match(sent, /邀请仍有效，但本机没有原链接/)
  assert.match(sent, /邀请记录未丢失。出于安全考虑，链接只保存在首次生成的设备/)
  assert.match(sent, /return hasStoredInviteToken\(item\) \? '可继续发送' : '邀请有效，需重新生成链接'/)
  assert.match(sent, /function hasStoredInviteToken\(item: Invitation\)/)
})

test('relationship health can take an administrator directly to adding a spouse', () => {
  const health = source('pages/family/health.vue')
  const tree = source('pages/family/private-tree.vue')

  assert.match(health, /issue\.key === 'spouse' \? '&task=ADD_SPOUSE'/)
  assert.match(tree, /requestedRelationAddType/)
  assert.match(tree, /openAddRelation\(target, addType\)/)
})

test('family settings contains operation records and tree avoids reloading the current member list', () => {
  const center = source('pages/family/manage-center.vue')
  const settings = source('pages/family/settings.vue')
  const tree = source('pages/family/private-tree.vue')

  assert.match(center, /家庭设置/)
  assert.match(settings, /操作记录/)
  assert.match(settings, /settingsSection === 'logs'/)
  assert.match(tree, /session\.user\?\.id && currentUserMemberId\.value == null/)
})

test('recovery is a dedicated reminder and is not counted as a management task', () => {
  const detail = source('pages/family/detail.vue')
  const center = source('pages/family/manage-center.vue')

  assert.match(detail, /最近删除可恢复/)
  assert.doesNotMatch(detail, /dissolutionCount \+ recoverableMemberCount\.value \+ healthTaskCount/)
  assert.doesNotMatch(detail, /if \(recoverableMemberCount\.value > 0\) return .*\/pages\/family\/recovery/)
  assert.doesNotMatch(center, /key: 'recovery'/)
  assert.doesNotMatch(center, /listRecoverableFamilyMembers/)
})

test('family detail shows a health count using the shared health summary', () => {
  const detail = source('pages/family/detail.vue')

  assert.match(detail, /summarizeFamilyHealth/)
  assert.match(detail, /pendingHealthCount/)
  assert.match(detail, /class="affair-badge health-badge"/)
})

test('health member chips take profile issues to the member view before editing', () => {
  const health = source('pages/family/health.vue')

  assert.match(health, /\['gender', 'birth', 'birth-order'\]\.includes\(issue\.key\)/)
  assert.match(health, /pages\/family\/member-view\?familyId=/)
})

test('family status pages retain loaded content and expose refresh failures', () => {
  const health = source('pages/family/health.vue')
  const detail = source('pages/family/detail.vue')
  const affairs = source('pages/family/manage-center.vue')

  assert.match(health, /loading\.value=!family\.value/)
  assert.match(health, /title="刷新未完成"/)
  assert.match(detail, /class="detail-refresh-error"/)
  assert.match(affairs, /class="manage-refresh-error"/)
})

test('family health prevents duplicate refreshes and duplicate navigation', () => {
  const health = source('pages/family/health.vue')

  assert.match(health, /loading\.value \|\| refreshing\.value/)
  assert.match(health, /function navigateOnce\(url: string\)/)
  assert.match(health, /function openIssue\(issue: Issue\) \{ navigateOnce\(issue\.route\) \}/)
})

test('received invitations lock refreshes and acknowledge accept or reject outcomes', () => {
  const received = source('pages/invite/my.vue')

  assert.match(received, /loading\.value \|\| refreshing\.value/)
  assert.match(received, /title: '已接受邀请'/)
  assert.match(received, /title: '已拒绝邀请'/)
})

test('member editing exposes a read-only state before a non-manager can submit another member profile', () => {
  const edit = source('pages/family/member-edit.vue')

  assert.match(edit, /const canEditCurrentMember = computed\(\(\) => canManageFamily\.value \|\| isCurrentUserMember\.value\)/)
  assert.match(edit, /v-if="canEditCurrentMember" class="archive-form-panel"/)
  assert.match(edit, /当前为只读查看/)
})

test('spouse display identity is edited from the member profile, not the relationship action list', () => {
  const memberEditor = readFileSync(new URL('../../pages/family/member-edit.vue', import.meta.url), 'utf8')
  const relationshipManager = readFileSync(new URL('../../pages/family/manage.vue', import.meta.url), 'utf8')

  assert.match(memberEditor, /婚姻展示身份/)
  assert.match(memberEditor, /spouseDisplayOptions/)
  assert.match(memberEditor, /updateRelationship\(familyId\.value, spouseRelationship\.value\.relationshipId/)
  assert.doesNotMatch(relationshipManager, /编辑展示身份/)
})

test('adding a child lets the manager select the recorded co-parent instead of inferring a spouse', () => {
  const manage = source('pages/family/manage.vue')

  assert.match(manage, /另一位家长/)
  assert.match(manage, /前任、现任配偶均可选/)
  assert.match(manage, /coParentMemberId: childCoParentIndex\.value > 0/)
})

test('adding a spouse captures the display identity before the node is created', () => {
  const manage = source('pages/family/manage.vue')

  assert.match(manage, /婚姻身份/)
  assert.match(manage, /前妻、前夫仅记录历史婚姻；现任妻、现任夫用于当前婚姻展示/)
  assert.match(manage, /relationNoteType: addType === 'ADD_SPOUSE'/)
})

test('child editing can complete a missing parent using a recorded former or current spouse', () => {
  const editor = source('pages/family/member-edit.vue')

  assert.match(editor, /另一位实际家长/)
  assert.match(editor, /前任、现任配偶中选择/)
  assert.match(editor, /coParentMemberId: selectedCoParent\.memberId/)
  assert.match(editor, /placeExistingMember\(familyId\.value/)
})

test('member verification is separated from editing and begins on a read-only member view', () => {
  const editor = source('pages/family/member-edit.vue')
  const memberView = source('pages/family/member-view.vue')
  const tree = source('pages/family/private-tree.vue')

  assert.doesNotMatch(editor, /确认资料无误|标记待完善|confirmFamilyMemberProfile/)
  assert.match(memberView, /核实操作只更新资料状态，不会修改当前档案内容/)
  assert.match(memberView, /确认资料无误/)
  assert.match(memberView, /标记待完善/)
  assert.match(memberView, /编辑资料/)
  assert.match(memberView, /<MiniButton @click="openEditor">编辑资料<\/MiniButton>/)
  assert.doesNotMatch(memberView, /<MiniButton @tap="openEditor">编辑资料<\/MiniButton>/)
  assert.match(tree, /label: '查看资料', primary: true/)
  assert.doesNotMatch(tree, /label: '编辑资料'/)
})

test('capacity order actions share one in-flight lock across payment and refund operations', () => {
  const orders = source('pages/family/capacity-orders.vue')

  assert.match(orders, /const orderActionBusy = computed\(\(\) => Boolean\(refundingOrderNo\.value \|\| reconcilingOrderNo\.value \|\| payingOrderNo\.value\)\)/)
  assert.match(orders, /if \(orderActionBusy\.value \|\| !canResumePayment\(order\)\) return/)
  assert.match(orders, /if \(orderActionBusy\.value \|\| !isRefundAvailable\(order\)\) return/)
})

test('public showcase keeps loaded cases visible when a later page fails', () => {
  const showcase = source('pages/family/showcase.vue')

  assert.match(showcase, /v-else-if="errorMessage && !items\.length"/)
  assert.match(showcase, /const loadMoreError = ref\(''\)/)
  assert.match(showcase, /<MiniNotice tone="warm" title="更多案例加载失败">/)
  assert.match(showcase, /function retryLoadMore\(\)/)
})

test('legacy public-join routes guide users to showcase browsing instead of a stranger join form', () => {
  const apply = source('pages/join/apply.vue')
  const history = source('pages/join/my.vue')

  assert.match(apply, /当前不支持公开申请加入家庭/)
  assert.match(apply, /pages\/family\/showcase/)
  assert.match(history, /历史加入申请/)
  assert.match(history, /当前家庭仅支持节点邀请绑定/)
})
