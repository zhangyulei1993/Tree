#!/usr/bin/env node
/**
 * Tree miniapp H5 real-API E2E data tool (staging only).
 *
 * Usage:
 *   node scripts/tree-h5-real-e2e-setup.mjs audit
 *   node scripts/tree-h5-real-e2e-setup.mjs setup
 *   node scripts/tree-h5-real-e2e-setup.mjs status
 *   node scripts/tree-h5-real-e2e-setup.mjs cleanup
 *   node scripts/tree-h5-real-e2e-setup.mjs repair-tree
 *   node scripts/tree-h5-real-e2e-setup.mjs cleanup-staging
 *   node scripts/tree-h5-real-e2e-setup.mjs setup-admin
 */

import crypto from 'node:crypto'
import fs from 'node:fs'
import process from 'node:process'

const CONFIG = {
  runId: 'tree-h5-real-e2e-v1',
  apiBase: process.env.TREE_E2E_API_BASE || 'https://tapi.bigbigboy.cn/api',
  stateFile: process.env.TREE_E2E_STATE_FILE || '/tmp/tree-h5-real-e2e-state.json',
  credentialsFile:
    process.env.TREE_E2E_CREDENTIALS_FILE || '/tmp/tree-h5-real-e2e-credentials.json',
  sendCodeDelayMs: Number(process.env.TREE_E2E_SEND_CODE_DELAY_MS || 65000),
  clientType: 'WECHAT_MINI_PROGRAM',
  mainFamilyName: '张氏家族·武汉江夏支系',
  mainFamilySurname: '张',
  mainNativePlace: '湖北武汉',
  mainRegionText: '湖北省武汉市江夏区',
  mainDescription:
    'H5 真实接口 E2E 基线家庭。四代结构，覆盖绑定、邀请、加入申请与家谱管理场景。',
  isolationFamilyName: '江夏张氏·流程验证专户',
  isolationDescription: 'H5 E2E 隔离家庭，供转让/解散/公开申请等高风险 Playwright 流程使用。',
  preservedFamilyIds: [12, 20, 21],
  adminCredentialsFile: '/tmp/tree-h5-real-e2e-admin-credentials.json',
  adminEnvFile: '/tmp/tree-h5-e2e-admin.env',
  autoDataMarkers: [
    'regression',
    'Regression',
    'Mock',
    'mock',
    '测试成员',
    'Mock 家庭',
    '用户 A',
    '访客 B',
    'PUBSRCH_',
    'Staging公开验收'
  ],
  accounts: {
    founder: { phone: '16600020101', nickname: '张承志', role: 'FOUNDER', memberName: '张承志' },
    familyAdmin: { phone: '16600020102', nickname: '张承浩', role: 'FAMILY_ADMIN', memberName: '张承浩' },
    member: { phone: '16600020103', nickname: '张承雅', role: 'MEMBER', memberName: '张承雅' },
    outsider: { phone: '16600020104', nickname: '周远航', role: 'OUTSIDER', memberName: null }
  }
}

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

function maskPhone(phone) {
  return `${phone.slice(0, 3)}****${phone.slice(-4)}`
}

function redactToken(token) {
  if (!token) return ''
  return `${token.slice(0, 8)}…(${token.length})`
}

function log(section, message, extra) {
  const prefix = `[tree-h5-e2e:${section}]`
  if (extra !== undefined) {
    console.log(prefix, message, typeof extra === 'string' ? extra : JSON.stringify(extra, null, 2))
    return
  }
  console.log(prefix, message)
}

function readJson(filePath, fallback = null) {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'))
  } catch {
    return fallback
  }
}

function writeJson(filePath, data) {
  fs.writeFileSync(filePath, `${JSON.stringify(data, null, 2)}\n`, 'utf8')
}

async function apiRequest(pathname, { method = 'GET', token, body } = {}) {
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`

  const response = await fetch(`${CONFIG.apiBase}${pathname}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body)
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok || payload.code !== 0) {
    const err = new Error(payload.message || `HTTP ${response.status}`)
    err.code = payload.code
    err.httpStatus = response.status
    err.payload = payload
    throw err
  }
  return payload.data
}

async function sendCode(phone, scene) {
  const data = await apiRequest('/auth/send-code', {
    method: 'POST',
    body: { phone, scene, clientType: CONFIG.clientType }
  })
  if (!data.devCode) throw new Error('staging 未返回 devCode')
  return data.devCode
}

async function registerAccount(accountKey, password) {
  const account = CONFIG.accounts[accountKey]
  const code = await sendCode(account.phone, 'REGISTER')
  const result = await apiRequest('/auth/register-phone', {
    method: 'POST',
    body: {
      phone: account.phone,
      password,
      code,
      nickname: account.nickname,
      clientType: CONFIG.clientType
    }
  })
  log('auth', `registered ${accountKey}`, maskPhone(account.phone))
  return result
}

async function loginAccount(accountKey, password) {
  const account = CONFIG.accounts[accountKey]
  return apiRequest('/auth/login-phone', {
    method: 'POST',
    body: { phone: account.phone, password, clientType: CONFIG.clientType }
  })
}

async function ensureAccount(accountKey, password) {
  try {
    const login = await loginAccount(accountKey, password)
    log('auth', `login ${accountKey}`, maskPhone(CONFIG.accounts[accountKey].phone))
    return login
  } catch (error) {
    if (error.code === 40201 || error.code === 40202 || error.message?.includes('未注册')) {
      await sleep(CONFIG.sendCodeDelayMs)
      return registerAccount(accountKey, password)
    }
    throw error
  }
}

async function ensureAllAccounts(password) {
  const sessions = {}
  for (const key of Object.keys(CONFIG.accounts)) {
    sessions[key] = await ensureAccount(key, password)
  }
  return sessions
}

function findMemberByName(members, name) {
  return members.find((item) => item.name === name)
}

async function addRelative(token, familyId, baseMemberId, addType, newMember) {
  return apiRequest(`/families/${familyId}/relationships`, {
    method: 'POST',
    token,
    body: {
      baseMemberId,
      addType,
      newMember,
      relationship: { parentLinkType: addType === 'ADD_SPOUSE' ? undefined : 'PRIMARY' }
    }
  })
}

async function updateMember(token, familyId, memberId, patch) {
  return apiRequest(`/families/${familyId}/members/${memberId}`, {
    method: 'PUT',
    token,
    body: patch
  })
}

function memberNeedsUpdate(member, patch) {
  for (const [key, value] of Object.entries(patch)) {
    const current = member[key]
    if (key === 'isAlive') {
      const currentAlive = current === true
      if (currentAlive !== value) return true
      continue
    }
    if (current !== value) return true
  }
  return false
}

async function updateMemberIfChanged(token, familyId, member, patch) {
  if (!memberNeedsUpdate(member, patch)) return false
  await updateMember(token, familyId, member.memberId, patch)
  return true
}

async function bindUser(token, familyId, memberId, userId) {
  return apiRequest(`/families/${familyId}/members/${memberId}/bind-user`, {
    method: 'POST',
    token,
    body: { userId: Number(userId) }
  })
}

async function setFamilyAdmin(token, familyId, memberId) {
  return apiRequest(`/families/${familyId}/members/${memberId}/set-admin`, {
    method: 'POST',
    token,
    body: {}
  })
}

async function createUnlocatedMember(token, familyId, name, gender = 'MALE') {
  return apiRequest(`/families/${familyId}/members`, {
    method: 'POST',
    token,
    body: { name, gender, isAlive: true, userBindingPolicy: 'OPTIONAL' }
  })
}

async function createInvitation(token, familyId, memberId) {
  return apiRequest(`/families/${familyId}/members/${memberId}/invite`, {
    method: 'POST',
    token,
    body: {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: 'H5 E2E 基线：邀请绑定父亲节点账号',
      familyRoleAfterAccept: 'MEMBER'
    }
  })
}

async function createJoinRequest(token, familyId, input) {
  return apiRequest(`/families/${familyId}/join-requests`, { method: 'POST', token, body: input })
}

async function listJoinRequests(token, familyId) {
  return apiRequest(`/families/${familyId}/join-requests`, { token })
}

async function cancelJoinRequest(token, familyId, requestId, reason) {
  return apiRequest(`/families/${familyId}/join-requests/${requestId}/cancel`, {
    method: 'POST',
    token,
    body: { cancelReason: reason }
  })
}

async function listFamilyInvitations(token, familyId) {
  return apiRequest(`/families/${familyId}/invitations`, { token })
}

async function cancelInvitation(token, invitationId, reason) {
  return apiRequest(`/invitations/${invitationId}/cancel`, {
    method: 'POST',
    token,
    body: { reason }
  })
}

async function getPrivateTree(token, familyId) {
  return apiRequest(`/families/${familyId}/tree`, { token })
}

async function listMembers(token, familyId) {
  return apiRequest(`/families/${familyId}/members`, { token })
}

async function ensureMainFamily(founderToken) {
  const families = await apiRequest('/families', { token: founderToken })
  const existing = families.find((item) => item.familyName === CONFIG.mainFamilyName)
  if (existing) {
    log('family', 'reuse main family', { familyId: existing.id })
    return existing.id
  }
  const created = await apiRequest('/families', {
    method: 'POST',
    token: founderToken,
    body: {
      surname: CONFIG.mainFamilySurname,
      familyName: CONFIG.mainFamilyName,
      founderGender: 'MALE',
      nativePlace: CONFIG.mainNativePlace,
      regionText: CONFIG.mainRegionText,
      description: CONFIG.mainDescription
    }
  })
  log('family', 'created main family', { familyId: created.id })
  return created.id
}

async function ensureIsolationFamily(founderToken) {
  const families = await apiRequest('/families', { token: founderToken })
  const existing = families.find((item) => item.familyName === CONFIG.isolationFamilyName)
  if (existing) return existing.id
  const created = await apiRequest('/families', {
    method: 'POST',
    token: founderToken,
    body: {
      surname: '张',
      familyName: CONFIG.isolationFamilyName,
      founderGender: 'MALE',
      nativePlace: '湖北武汉',
      regionText: '湖北省武汉市江夏区',
      description: CONFIG.isolationDescription
    }
  })
  log('family', 'created isolation family', { familyId: created.id })
  return created.id
}

async function buildMainFamilyTree(founderToken, familyId) {
  let members = await listMembers(founderToken, familyId)
  const founder = findMemberByName(members, '张承志')
  if (!founder) throw new Error('未找到创建者成员 张承志')
  const founderId = founder.memberId

  const steps = [
    ['张文昌', founderId, 'ADD_FATHER', { name: '张文昌', gender: 'MALE', birthYear: 1968, isAlive: true, userBindingPolicy: 'OPTIONAL' }],
    ['林晓梅', founderId, 'ADD_MOTHER', { name: '林晓梅', gender: 'FEMALE', birthYear: 1970, isAlive: true, userBindingPolicy: 'REQUIRED' }],
    ['陈雨桐', founderId, 'ADD_SPOUSE', { name: '陈雨桐', gender: 'FEMALE', birthYear: 1993, isAlive: true, userBindingPolicy: 'REQUIRED' }],
    ['张思远', founderId, 'ADD_CHILD', { name: '张思远', gender: 'MALE', birthYear: 2018, isAlive: true, userBindingPolicy: 'OPTIONAL' }],
    ['张思琪', founderId, 'ADD_CHILD', { name: '张思琪', gender: 'FEMALE', birthYear: 2020, isAlive: true, userBindingPolicy: 'OPTIONAL' }],
    ['张思涵', founderId, 'ADD_CHILD', { name: '张思涵', gender: 'MALE', birthYear: 2022, isAlive: true, userBindingPolicy: 'OPTIONAL' }],
    ['张承浩', founderId, 'ADD_SIBLING', { name: '张承浩', gender: 'MALE', birthYear: 1994, isAlive: true, userBindingPolicy: 'REQUIRED' }],
    ['张承雅', founderId, 'ADD_SIBLING', { name: '张承雅', gender: 'FEMALE', birthYear: 1997, isAlive: true, userBindingPolicy: 'REQUIRED' }]
  ]

  for (const [name, baseId, addType, member] of steps) {
    members = await listMembers(founderToken, familyId)
    if (findMemberByName(members, name)) continue
    const result = await addRelative(founderToken, familyId, baseId, addType, member)
    log('tree', `added ${name}`, { memberId: result.createdMember.memberId, graphVersion: result.graphVersion })
    await sleep(300)
  }

  members = await listMembers(founderToken, familyId)
  const brotherId = findMemberByName(members, '张承浩')?.memberId
  const fatherId = findMemberByName(members, '张文昌')?.memberId

  if (brotherId && !findMemberByName(members, '孙丽华')) {
    await addRelative(founderToken, familyId, brotherId, 'ADD_SPOUSE', {
      name: '孙丽华', gender: 'FEMALE', birthYear: 1995, isAlive: true, userBindingPolicy: 'OPTIONAL'
    })
    await addRelative(founderToken, familyId, brotherId, 'ADD_CHILD', {
      name: '张乐童', gender: 'MALE', birthYear: 2019, isAlive: true, userBindingPolicy: 'OPTIONAL'
    })
  }

  if (fatherId && !findMemberByName(members, '张启明')) {
    await addRelative(founderToken, familyId, fatherId, 'ADD_FATHER', {
      name: '张启明', gender: 'MALE', birthYear: 1945, isAlive: false, userBindingPolicy: 'NOT_REQUIRED'
    })
    await addRelative(founderToken, familyId, fatherId, 'ADD_MOTHER', {
      name: '赵清秀', gender: 'FEMALE', birthYear: 1948, isAlive: true, userBindingPolicy: 'NOT_REQUIRED'
    })
  }

  members = await listMembers(founderToken, familyId)
  const grandfatherId = findMemberByName(members, '张启明')?.memberId
  if (grandfatherId && !findMemberByName(members, '张礼安')) {
    const great = await addRelative(founderToken, familyId, grandfatherId, 'ADD_FATHER', {
      name: '张礼安', gender: 'MALE', birthYear: 1920, deathYear: 1998, isAlive: false, userBindingPolicy: 'NOT_REQUIRED'
    })
    await addRelative(founderToken, familyId, great.createdMember.memberId, 'ADD_SPOUSE', {
      name: '周淑贞', gender: 'FEMALE', birthYear: 1923, deathYear: 2001, isAlive: false, userBindingPolicy: 'NOT_REQUIRED'
    })
  }

  members = await listMembers(founderToken, familyId)
  const qimingId = findMemberByName(members, '张启明')?.memberId
  if (qimingId && !findMemberByName(members, '张启国')) {
    await addRelative(founderToken, familyId, qimingId, 'ADD_SIBLING', {
      name: '张启国', gender: 'MALE', birthYear: 1947, isAlive: true, userBindingPolicy: 'NOT_REQUIRED'
    })
  }

  if (!findMemberByName(members, '张待安')) {
    await createUnlocatedMember(founderToken, familyId, '张待安', 'MALE')
  }

  members = await listMembers(founderToken, familyId)
  const deceasedPatches = {
    张礼安: { isAlive: false, deathYear: 1998, userBindingPolicy: 'NOT_REQUIRED' },
    周淑贞: { isAlive: false, deathYear: 2001, userBindingPolicy: 'NOT_REQUIRED' },
    张启明: { isAlive: false, deathYear: 2016, userBindingPolicy: 'NOT_REQUIRED' }
  }
  for (const item of members) {
    const patch = deceasedPatches[item.name]
    if (patch) await updateMemberIfChanged(founderToken, familyId, item, patch)
  }

  members = await listMembers(founderToken, familyId)
  return Object.fromEntries(members.map((item) => [item.name, item.memberId]))
}

async function bindRoleAccounts(founderToken, familyId, sessions, memberMap) {
  const adminMemberId = memberMap['张承浩']
  const memberMemberId = memberMap['张承雅']
  const members = await listMembers(founderToken, familyId)
  const adminMember = members.find((item) => item.memberId === adminMemberId)
  const plainMember = members.find((item) => item.memberId === memberMemberId)

  if (adminMemberId && !adminMember?.boundUserId) {
    await bindUser(founderToken, familyId, adminMemberId, sessions.familyAdmin.user.id)
    await setFamilyAdmin(founderToken, familyId, adminMemberId)
  }
  if (memberMemberId && !plainMember?.boundUserId) {
    await bindUser(founderToken, familyId, memberMemberId, sessions.member.user.id)
  }
}

async function ensurePendingInvitation(founderToken, familyId, memberMap) {
  const targetId = memberMap['张文昌']
  if (!targetId) return null
  const invitations = await listFamilyInvitations(founderToken, familyId)
  const pending = invitations.find(
    (item) => String(item.targetMemberId) === String(targetId) && item.status === 'PENDING'
  )
  if (pending) return pending.invitationId
  const created = await createInvitation(founderToken, familyId, targetId)
  return created.invitation.invitationId
}

async function listMyJoinRequests(token) {
  return apiRequest('/users/me/join-requests', { token })
}

async function ensurePendingJoinRequest(outsiderToken, familyId) {
  const requests = await listMyJoinRequests(outsiderToken)
  const pending = requests.find(
    (item) => String(item.familyId) === String(familyId) && item.requestStatus === 'PENDING'
  )
  if (pending) return pending.requestId
  const created = await createJoinRequest(outsiderToken, familyId, {
    applicantRealName: '周远航',
    applicantGender: 'MALE',
    applicantMessage: 'H5 E2E 基线：待审批加入申请，供 Playwright 审批流程使用。'
  })
  return created.requestId
}

async function auditStaging() {
  const version = await apiRequest('/version')
  const publicFamilies = await apiRequest('/public/families?page=1&pageSize=100')
  const flagged = []
  const preserved = []
  for (const item of publicFamilies.items || []) {
    const text = [item.familyName, item.description, item.regionText].filter(Boolean).join(' ')
    const matched = CONFIG.autoDataMarkers.filter((marker) => text.includes(marker))
    if (matched.length > 0) flagged.push({ familyId: item.id, familyName: item.familyName, markers: matched })
    else preserved.push({ familyId: item.id, familyName: item.familyName })
  }
  const report = {
    auditedAt: new Date().toISOString(),
    apiBase: CONFIG.apiBase,
    version,
    publicFamiliesTotal: publicFamilies.total,
    flaggedPublicFamilies: flagged,
    preservedPublicFamilies: preserved,
    retainedNotes: [
      'familyId=12 张氏家族测试数据-M22 为长期演示数据，不在自动清理范围',
      'familyId=8 Staging公开验收189912 需 ROOT_ADMIN 审核下架，脚本未删除',
      '169122942xx 回归账号无密码记录，无法通过 API 安全注销'
    ]
  }
  writeJson('/tmp/tree-h5-real-e2e-audit.json', report)
  log('audit', 'report', '/tmp/tree-h5-real-e2e-audit.json')
  return report
}

async function setup() {
  const password =
    process.env.TREE_E2E_PASSWORD
    || readJson(CONFIG.credentialsFile)?.password
    || `TreeH5E2e@${crypto.randomBytes(3).toString('hex')}`

  await auditStaging()
  const sessions = await ensureAllAccounts(password)
  const founderToken = sessions.founder.accessToken
  const outsiderToken = sessions.outsider.accessToken
  const mainFamilyId = await ensureMainFamily(founderToken)
  const isolationFamilyId = await ensureIsolationFamily(founderToken)
  const memberMap = await buildMainFamilyTree(founderToken, mainFamilyId)
  await bindRoleAccounts(founderToken, mainFamilyId, sessions, memberMap)
  const invitationId = await ensurePendingInvitation(founderToken, mainFamilyId, memberMap)
  const joinRequestId = await ensurePendingJoinRequest(outsiderToken, mainFamilyId)
  const tree = await getPrivateTree(founderToken, mainFamilyId)
  const members = await listMembers(founderToken, mainFamilyId)

  const state = {
    runId: CONFIG.runId,
    apiBase: CONFIG.apiBase,
    updatedAt: new Date().toISOString(),
    mainFamily: {
      familyId: mainFamilyId,
      familyName: CONFIG.mainFamilyName,
      graphVersion: tree.graphVersion,
      memberCount: members.length,
      edgeCount: tree.edges?.length || 0,
      memberMap
    },
    isolationFamily: { familyId: isolationFamilyId, familyName: CONFIG.isolationFamilyName },
    accounts: Object.fromEntries(
      Object.entries(CONFIG.accounts).map(([key, account]) => [
        key,
        {
          phoneMasked: maskPhone(account.phone),
          userId: sessions[key].user.id,
          memberId: account.memberName ? memberMap[account.memberName] || null : null,
          role: account.role
        }
      ])
    ),
    pending: {
      invitationId,
      joinRequestId,
      inviteTargetMemberId: memberMap['张文昌'] || null
    },
    scenarios: {
      inviteTargetMemberName: '张文昌',
      deletableLeafMemberName: '张思涵',
      nonDeletableMemberName: '张承志',
      unlocatedMemberName: '张待安',
      deceasedMemberNames: ['张礼安', '周淑贞', '张启明'],
      notRequiredMemberNames: ['张礼安', '周淑贞', '张启明', '赵清秀', '张启国']
    }
  }

  writeJson(CONFIG.stateFile, state)
  writeJson(CONFIG.credentialsFile, {
    updatedAt: new Date().toISOString(),
    apiBase: CONFIG.apiBase,
    password,
    accounts: Object.fromEntries(
      Object.entries(CONFIG.accounts).map(([key, account]) => [
        key,
        { phone: account.phone, nickname: account.nickname, role: account.role }
      ])
    ),
    tokens: Object.fromEntries(
      Object.entries(sessions).map(([key, session]) => [
        key,
        { tokenPreview: redactToken(session.accessToken), userId: session.user.id }
      ])
    )
  })

  log('setup', 'done', {
    stateFile: CONFIG.stateFile,
    credentialsFile: CONFIG.credentialsFile,
    mainFamilyId,
    memberCount: members.length,
    graphVersion: tree.graphVersion,
    invitationId,
    joinRequestId
  })
  return state
}

function status() {
  const state = readJson(CONFIG.stateFile)
  if (!state) {
    log('status', 'no state; run setup first')
    process.exitCode = 1
    return
  }
  log('status', CONFIG.stateFile, state)
}

async function placeExistingMember(token, familyId, baseMemberId, memberId, addType) {
  return apiRequest(`/families/${familyId}/relationships/place-existing`, {
    method: 'POST',
    token,
    body: {
      baseMemberId,
      memberId,
      addType,
      relationship: { parentLinkType: addType === 'ADD_SPOUSE' ? undefined : 'PRIMARY' }
    }
  })
}

async function deleteRelationship(token, familyId, relationshipId, reason) {
  return apiRequest(`/families/${familyId}/relationships/${relationshipId}`, {
    method: 'DELETE',
    token,
    body: reason ? { reason } : {}
  })
}

async function unbindUser(token, familyId, memberId, reason) {
  return apiRequest(`/families/${familyId}/members/${memberId}/unbind-user`, {
    method: 'POST',
    token,
    body: reason ? { reason } : {}
  })
}

async function rejectJoinRequest(token, familyId, requestId, comment) {
  return apiRequest(`/families/${familyId}/join-requests/${requestId}/reject`, {
    method: 'POST',
    token,
    body: { handleComment: comment }
  })
}

function loadAdminEnv() {
  if (!fs.existsSync(CONFIG.adminEnvFile)) return {}
  const env = {}
  for (const line of fs.readFileSync(CONFIG.adminEnvFile, 'utf8').split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue
    const index = trimmed.indexOf('=')
    if (index <= 0) continue
    env[trimmed.slice(0, index).trim()] = trimmed.slice(index + 1).trim()
  }
  return env
}

async function adminLogin() {
  const cached = readJson(CONFIG.adminCredentialsFile)
  const env = loadAdminEnv()
  const username = process.env.TREE_E2E_ADMIN_USERNAME || env.TREE_E2E_ADMIN_USERNAME || cached?.username
  const password = process.env.TREE_E2E_ADMIN_PASSWORD || env.TREE_E2E_ADMIN_PASSWORD || cached?.password
  if (!username || !password) {
    throw new Error(`缺少管理员凭据，请写入 ${CONFIG.adminEnvFile}`)
  }
  const data = await apiRequest('/admin/auth/login', {
    method: 'POST',
    body: { username, password }
  })
  writeJson(CONFIG.adminCredentialsFile, {
    updatedAt: new Date().toISOString(),
    username,
    password,
    role: data.admin?.role,
    tokenPreview: redactToken(data.accessToken),
    accessToken: data.accessToken
  })
  return data
}

function getParentMemberIds(tree, memberId) {
  return tree.edges
    .filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.toMemberId === memberId)
    .map((edge) => edge.fromMemberId)
}

function membersShareParents(tree, memberIdA, memberIdB) {
  const parentsA = new Set(getParentMemberIds(tree, memberIdA))
  return getParentMemberIds(tree, memberIdB).some((parentId) => parentsA.has(parentId))
}

async function repairMainFamilyTree(founderToken, familyId) {
  const tree = await getPrivateTree(founderToken, familyId)
  const nameById = Object.fromEntries(tree.nodes.map((node) => [node.memberId, node.displayName]))
  const actions = []

  for (const edge of tree.edges) {
    if (edge.relationshipType !== 'PARENT_CHILD') continue
    const from = nameById[edge.fromMemberId]
    const to = nameById[edge.toMemberId]
    const wrongGuoguoChild =
      (from === '张启明' || from === '赵清秀') && to === '张启国'
    const wrongDaiAnChild =
      (from === '张文昌' || from === '林晓梅') && to === '张待安'
    if (wrongGuoguoChild || wrongDaiAnChild) {
      await deleteRelationship(founderToken, familyId, edge.relationshipId, 'H5 E2E repair tree')
      actions.push(`deleted rel ${edge.relationshipId}: ${from} -> ${to}`)
    }
  }

  const members = await listMembers(founderToken, familyId)
  const memberMap = Object.fromEntries(members.map((item) => [item.name, item.memberId]))
  const qimingId = memberMap['张启明']
  const qiguoId = memberMap['张启国']
  if (qimingId && qiguoId) {
    const treeAfter = await getPrivateTree(founderToken, familyId)
    if (!membersShareParents(treeAfter, qimingId, qiguoId)) {
      await placeExistingMember(founderToken, familyId, qimingId, qiguoId, 'ADD_SIBLING')
      actions.push('placed 张启国 as sibling of 张启明')
    }
  } else if (qimingId && !qiguoId) {
    await addRelative(founderToken, familyId, qimingId, 'ADD_SIBLING', {
      name: '张启国',
      gender: 'MALE',
      birthYear: 1947,
      isAlive: true,
      userBindingPolicy: 'NOT_REQUIRED'
    })
    actions.push('recreated 张启国 as sibling of 张启明')
  }

  for (const item of members) {
    if (item.birthYear && item.birthYear > new Date().getFullYear()) {
      actions.push(`WARN future birthYear ${item.name}: ${item.birthYear}`)
    }
  }

  return actions
}

async function refreshMainFamilyState(founderToken, familyId) {
  const state = readJson(CONFIG.stateFile)
  if (!state?.mainFamily) return null
  const tree = await getPrivateTree(founderToken, familyId)
  const members = await listMembers(founderToken, familyId)
  state.updatedAt = new Date().toISOString()
  state.mainFamily.graphVersion = tree.graphVersion
  state.mainFamily.memberCount = members.length
  state.mainFamily.edgeCount = tree.edges?.length || 0
  state.mainFamily.memberMap = Object.fromEntries(members.map((item) => [item.name, item.memberId]))
  writeJson(CONFIG.stateFile, state)
  return tree.graphVersion
}

async function repairTree() {
  const credentials = readJson(CONFIG.credentialsFile)
  if (!credentials?.password) throw new Error('missing user credentials')
  const founder = await loginAccount('founder', credentials.password)
  const familyId = readJson(CONFIG.stateFile)?.mainFamily?.familyId || 20
  const graphVersionBefore = (await getPrivateTree(founder.accessToken, familyId)).graphVersion
  const actions = await repairMainFamilyTree(founder.accessToken, familyId)
  const graphVersionAfter = await refreshMainFamilyState(founder.accessToken, familyId)
  writeJson('/tmp/tree-h5-real-e2e-repair.json', {
    repairedAt: new Date().toISOString(),
    familyId,
    actions,
    graphVersionBefore,
    graphVersionAfter
  })
  log('repair', 'done', { actions, graphVersionBefore, graphVersionAfter })
}

async function setupAdmin() {
  if (!fs.existsSync(CONFIG.adminEnvFile)) {
    fs.writeFileSync(
      CONFIG.adminEnvFile,
      '# staging admin credentials (never commit)\nTREE_E2E_ADMIN_USERNAME=admin\nTREE_E2E_ADMIN_PASSWORD=\n',
      'utf8'
    )
    log('admin', 'created env template', CONFIG.adminEnvFile)
  }
  try {
    const session = await adminLogin()
    log('admin', 'login ok', {
      username: readJson(CONFIG.adminCredentialsFile)?.username,
      role: session.admin?.role || session.role,
      credentialsFile: CONFIG.adminCredentialsFile
    })
  } catch (error) {
    log('admin', 'login skipped', error.message)
  }
}

async function loginPhone(phone, password) {
  return apiRequest('/auth/login-phone', {
    method: 'POST',
    body: { phone, password, clientType: CONFIG.clientType }
  })
}

async function cleanupStaging() {
  const report = { cleanedAt: new Date().toISOString(), actions: [], skipped: [] }
  const m22Password = 'TreeTest@2026'

  try {
    const f12 = await loginPhone('16600010001', m22Password)
    const token12 = f12.accessToken
    const joins = await apiRequest('/families/12/join-requests', { token: token12 })
    for (const item of joins) {
      if (item.requestStatus === 'PENDING' && item.applicantRealName === '张玉磊') {
        await rejectJoinRequest(token12, 12, item.requestId, 'H5 E2E cleanup: 清理遗留待审申请')
        report.actions.push(`rejected family 12 join request ${item.requestId} 张玉磊`)
      }
    }
    const members12 = await listMembers(token12, 12)
    const linSiqi = members12.find((item) => item.name === '林思琪' && item.boundUserId)
    if (linSiqi) {
      await unbindUser(token12, 12, linSiqi.memberId, 'H5 E2E cleanup: 恢复张承雅绑定测试')
      report.actions.push(`unbound family 12 member ${linSiqi.memberId} 林思琪 from user ${linSiqi.boundUserId}`)
    }
  } catch (error) {
    report.skipped.push(`family 12 cleanup: ${error.message}`)
  }

  try {
    const credentials = readJson(CONFIG.credentialsFile)
    if (credentials?.password) {
      const founder = await loginAccount('founder', credentials.password)
      const familyId = readJson(CONFIG.stateFile)?.mainFamily?.familyId || 20
      const repairActions = await repairMainFamilyTree(founder.accessToken, familyId)
      report.actions.push(...repairActions.map((item) => `family ${familyId}: ${item}`))
    }
  } catch (error) {
    report.skipped.push(`family 20 repair: ${error.message}`)
  }

  try {
    const admin = await adminLogin()
    const adminToken = admin.accessToken
    const pendingMessages = await apiRequest('/admin/visitor-messages?status=PENDING&page=1&pageSize=100', {
      token: adminToken
    })
    for (const item of pendingMessages.items || []) {
      if (item.visitorName === '武汉访客' || (item.messageContent || '').includes('武汉')) {
        await apiRequest(`/admin/visitor-messages/${item.messageId}/reject`, {
          method: 'POST',
          token: adminToken,
          body: { reviewComment: 'H5 E2E cleanup' }
        })
        report.actions.push(`rejected visitor message ${item.messageId} ${item.visitorName}`)
      }
    }
    if (CONFIG.preservedFamilyIds.includes(8)) {
      report.skipped.push('family 8 is preserved; skipping take-down')
    } else {
      try {
        await apiRequest('/admin/families/8/take-down-public', {
          method: 'POST',
          token: adminToken,
          body: { reason: 'H5 E2E cleanup regression public family' }
        })
        report.actions.push('took down public display for family 8')
      } catch (error) {
        report.skipped.push(`family 8 take-down: ${error.message}`)
      }
    }
    const families = await apiRequest('/admin/families?page=1&pageSize=100', { token: adminToken })
    for (const family of families.items || []) {
      if (CONFIG.preservedFamilyIds.includes(Number(family.id))) continue
      const text = [family.familyName, family.description, family.regionText].filter(Boolean).join(' ')
      const matched = CONFIG.autoDataMarkers.some((marker) => text.includes(marker))
      if (matched) {
        report.skipped.push(`family ${family.id} ${family.familyName}: marked for manual admin dissolution`)
      }
    }
  } catch (error) {
    report.skipped.push(`admin cleanup: ${error.message}`)
  }

  writeJson('/tmp/tree-h5-real-e2e-cleanup-staging.json', report)
  log('cleanup-staging', 'done', report)
}

async function cleanup() {
  const state = readJson(CONFIG.stateFile)
  const credentials = readJson(CONFIG.credentialsFile)
  if (!state || !credentials?.password) return
  const founder = await loginAccount('founder', credentials.password)
  const outsider = await loginAccount('outsider', credentials.password)
  const actions = []
  if (state.pending?.joinRequestId) {
    try {
      await cancelJoinRequest(
        outsider.accessToken,
        state.mainFamily.familyId,
        state.pending.joinRequestId,
        'H5 E2E cleanup'
      )
      actions.push(`cancelled join ${state.pending.joinRequestId}`)
    } catch (error) {
      actions.push(`skip join: ${error.message}`)
    }
  }
  if (state.pending?.invitationId) {
    try {
      await cancelInvitation(founder.accessToken, state.pending.invitationId, 'H5 E2E cleanup')
      actions.push(`cancelled invite ${state.pending.invitationId}`)
    } catch (error) {
      actions.push(`skip invite: ${error.message}`)
    }
  }
  writeJson('/tmp/tree-h5-real-e2e-cleanup.json', { cleanedAt: new Date().toISOString(), actions })
  log('cleanup', 'done', actions)
}

async function main() {
  const command = process.argv[2] || 'status'
  if (command === 'audit') await auditStaging()
  else if (command === 'setup') await setup()
  else if (command === 'status') status()
  else if (command === 'cleanup') await cleanup()
  else if (command === 'repair-tree') await repairTree()
  else if (command === 'cleanup-staging') await cleanupStaging()
  else if (command === 'setup-admin') await setupAdmin()
  else {
    console.error(`Unknown command: ${command}`)
    process.exitCode = 1
  }
}

main().catch((error) => {
  console.error('[tree-h5-e2e:fatal]', error.message)
  process.exitCode = 1
})
