#!/usr/bin/env node
/**
 * Seed an isolated patrilineal test family without touching familyId=22.
 *
 * Usage:
 *   TREE_API_BASE=https://tapi.bigbigboy.cn/api \
 *   TREE_SEED_PHONE=16600010001 \
 *   TREE_SEED_PASSWORD='***' \
 *   node scripts/seed-isolated-patrilineal-family.mjs
 *
 *   node scripts/seed-isolated-patrilineal-family.mjs --apply
 */

import fs from 'node:fs'
import path from 'node:path'

import {
  SEED_UNLOCATED_NAME,
  buildSeedPlan,
  summarizeSeedPlan
} from './lib/patrilineal-seed-plan.mjs'

const APPLY = process.argv.includes('--apply')
const CONFIG = {
  apiBase: process.env.TREE_API_BASE || 'https://tapi.bigbigboy.cn/api',
  phone: process.env.TREE_SEED_PHONE || process.env.TREE_FAMILY_22_PHONE || '',
  password: process.env.TREE_SEED_PASSWORD || process.env.TREE_FAMILY_22_PASSWORD || '',
  familyName: process.env.TREE_SEED_FAMILY_NAME || 'Tree父系隔离测试家',
  familySurname: process.env.TREE_SEED_FAMILY_SURNAME || '张',
  snapshotDir: process.env.TREE_SEED_SNAPSHOT_DIR || path.join(process.cwd(), 'scripts', '.snapshots')
}

const runState = {
  familyId: null,
  createdFamily: false,
  createdMemberIds: [],
  createdRelationshipIds: [],
  snapshotPath: null,
  writesAttempted: 0
}

async function apiRequest(pathname, { method = 'GET', token, body } = {}) {
  const isAuthLogin = method === 'POST' && pathname === '/auth/login-phone'
  if (!APPLY && method !== 'GET' && !isAuthLogin) {
    throw new Error(`dry-run blocked write attempt: ${method} ${pathname}`)
  }
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
    const message = payload.message || `HTTP ${response.status}`
    const error = new Error(message)
    error.payload = payload
    throw error
  }
  runState.writesAttempted += method === 'GET' || isAuthLogin ? 0 : 1
  return payload.data
}

async function loginPhone(phone, password) {
  return apiRequest('/auth/login-phone', {
    method: 'POST',
    token: null,
    body: { phone, password, clientType: 'WECHAT_MINI_PROGRAM' }
  })
}

async function listMyFamilies(token) {
  return apiRequest('/families', { token })
}

async function createFamily(token) {
  return apiRequest('/families', {
    method: 'POST',
    token,
    body: {
      surname: CONFIG.familySurname,
      founderGender: 'MALE',
      familyName: CONFIG.familyName
    }
  })
}

async function getPrivateTree(token, familyId) {
  return apiRequest(`/families/${familyId}/tree`, { token })
}

async function getFamilyDetail(token, familyId) {
  return apiRequest(`/families/${familyId}`, { token })
}

async function listMembers(token, familyId) {
  return apiRequest(`/families/${familyId}/members`, { token })
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

async function createUnlocatedMember(token, familyId, name, gender = 'MALE') {
  return apiRequest(`/families/${familyId}/members`, {
    method: 'POST',
    token,
    body: { name, gender, userBindingPolicy: 'OPTIONAL', isAlive: true }
  })
}

async function deleteRelationship(token, familyId, relationshipId, reason) {
  return apiRequest(`/families/${familyId}/relationships/${relationshipId}`, {
    method: 'DELETE',
    token,
    body: { reason }
  })
}

async function deleteMember(token, familyId, memberId, reason) {
  return apiRequest(`/families/${familyId}/members/${memberId}`, {
    method: 'DELETE',
    token,
    body: { reason }
  })
}

function familyIdOf(family) {
  return family?.id ?? family?.familyId
}

function ensureSnapshotDir() {
  fs.mkdirSync(CONFIG.snapshotDir, { recursive: true })
}

function writeSnapshot(label, data) {
  ensureSnapshotDir()
  const filename = `${label}-${Date.now()}.json`
  const fullPath = path.join(CONFIG.snapshotDir, filename)
  fs.writeFileSync(fullPath, JSON.stringify(data, null, 2), 'utf8')
  return fullPath
}

function findExistingFamily(families) {
  return (families || []).find((family) => family.familyName === CONFIG.familyName)
}

function founderDisplayName(detail, members, founderMemberId) {
  const fromMembers = (members || []).find((member) => Number(member.memberId) === Number(founderMemberId))
  if (fromMembers?.name) return fromMembers.name
  return detail?.founderDisplayName || '创始人'
}

function countCoupleUnits(tree) {
  const spousePairs = new Set()
  for (const edge of tree.edges || []) {
    if (edge.relationshipType !== 'SPOUSE') continue
    spousePairs.add([edge.fromMemberId, edge.toMemberId].sort((a, b) => a - b).join('-'))
  }
  return spousePairs.size
}

function unlocatedLineageMembers(tree) {
  const related = new Set()
  for (const edge of tree.edges || []) {
    related.add(edge.fromMemberId)
    related.add(edge.toMemberId)
  }
  return (tree.nodes || []).filter((node) => {
    if (related.has(node.memberId)) return false
    return String(node.memberType).toUpperCase() === 'LINEAGE_MEMBER'
  })
}

function validateTree(tree, familyId, plan) {
  if (Number(tree.familyId) !== Number(familyId)) {
    throw new Error(`family id mismatch: expected ${familyId}, got ${tree.familyId}`)
  }
  const memberCount = (tree.nodes || []).length
  const relationshipCount = (tree.edges || []).length
  const coupleUnits = countCoupleUnits(tree)
  const unlocated = unlocatedLineageMembers(tree)
  const founderNode = (tree.nodes || []).find(
    (node) => Number(node.memberId) === Number(plan.founderMemberId)
  )
  const duplicateFounderNames = (tree.nodes || []).filter((node) => node.displayName === plan.founderDisplayName)
  if (duplicateFounderNames.length !== 1) {
    throw new Error(`expected exactly one founder node named ${plan.founderDisplayName}, got ${duplicateFounderNames.length}`)
  }
  if (!founderNode) {
    throw new Error(`founder member ${plan.founderMemberId} missing from tree nodes`)
  }
  if (memberCount !== plan.expected.memberCount) {
    throw new Error(`member count mismatch: expected ${plan.expected.memberCount}, got ${memberCount}`)
  }
  if (relationshipCount !== plan.expected.relationshipCount) {
    throw new Error(`relationship count mismatch: expected ${plan.expected.relationshipCount}, got ${relationshipCount}`)
  }
  if (coupleUnits !== plan.expected.coupleUnits) {
    throw new Error(`couple unit mismatch: expected ${plan.expected.coupleUnits}, got ${coupleUnits}`)
  }
  if (unlocated.length !== plan.expected.unlocatedCount) {
    throw new Error(`unlocated mismatch: expected ${plan.expected.unlocatedCount}, got ${unlocated.length}`)
  }
  if (unlocated[0]?.displayName !== SEED_UNLOCATED_NAME) {
    throw new Error(`expected unlocated member ${SEED_UNLOCATED_NAME}`)
  }
  return {
    familyId: Number(familyId),
    memberCount,
    relationshipCount,
    coupleUnits,
    unlocatedCount: unlocated.length,
    founderMemberId: plan.founderMemberId,
    founderDisplayName: plan.founderDisplayName
  }
}

async function resolveFamily(token) {
  const mine = await listMyFamilies(token)
  const existing = findExistingFamily(mine)
  if (existing) {
    runState.familyId = familyIdOf(existing)
    return existing
  }
  if (!APPLY) {
    return { id: 'NEW', familyName: CONFIG.familyName, currentFounderMemberId: null }
  }
  const created = await createFamily(token)
  runState.familyId = familyIdOf(created)
  runState.createdFamily = true
  return created
}

async function executeSeedPlan(token, familyId, plan) {
  const ctx = { founderMemberId: plan.founderMemberId }
  for (const step of plan.steps) {
    console.log(`apply -> ${step.key}`)
    if (step.action === 'CREATE_MEMBER') {
      const result = await createUnlocatedMember(token, familyId, step.newMember.name, step.newMember.gender)
      runState.createdMemberIds.push(result.memberId)
      continue
    }
    const baseMemberId =
      step.baseMemberId === '$father'
        ? ctx.fatherId
        : step.baseMemberId === '$grandfather'
          ? ctx.grandfatherId
          : step.baseMemberId
    const payload = {
      ...step.newMember,
      isAlive: true,
      userBindingPolicy: step.newMember.memberType === 'SPOUSE' ? 'NOT_REQUIRED' : 'OPTIONAL'
    }
    const result = await addRelative(token, familyId, baseMemberId, step.addType, payload)
    runState.createdMemberIds.push(result.createdMember.memberId)
    runState.createdRelationshipIds.push(...result.relationships.map((rel) => rel.relationshipId))
    if (step.key === 'father') ctx.fatherId = result.createdMember.memberId
    if (step.key === 'grandfather') ctx.grandfatherId = result.createdMember.memberId
  }
}

async function cleanupPartial(token, familyId) {
  console.warn('[cleanup] rolling back relationships and members created in this run')
  for (const relationshipId of [...new Set(runState.createdRelationshipIds)].reverse()) {
    try {
      await deleteRelationship(token, familyId, relationshipId, 'seed-isolated-patrilineal-family rollback')
    } catch (error) {
      console.warn(`[cleanup] failed to delete relationship ${relationshipId}: ${error.message}`)
    }
  }
  for (const memberId of [...new Set(runState.createdMemberIds)].reverse()) {
    try {
      await deleteMember(token, familyId, memberId, 'seed-isolated-patrilineal-family rollback')
    } catch (error) {
      console.warn(`[cleanup] failed to delete member ${memberId}: ${error.message}`)
    }
  }
  if (runState.createdFamily) {
    console.warn('[cleanup] created family remains and must be dissolved manually if rollback is incomplete')
  }
}

async function main() {
  if (!CONFIG.phone || !CONFIG.password) {
    console.error('Set TREE_SEED_PHONE and TREE_SEED_PASSWORD before running.')
    process.exit(1)
  }

  const login = await loginPhone(CONFIG.phone, CONFIG.password)
  const token = login.accessToken
  const family = await resolveFamily(token)
  const familyId = familyIdOf(family)

  if (Number(familyId) === 22) {
    throw new Error('refusing to operate on familyId=22')
  }

  let detail = family
  let members = []
  let tree = { nodes: [], edges: [], familyId }
  let founderMemberId = detail.currentFounderMemberId

  if (familyId !== 'NEW') {
    detail = await getFamilyDetail(token, familyId)
    if (detail.familyName !== CONFIG.familyName) {
      throw new Error(`family name mismatch: expected ${CONFIG.familyName}, got ${detail.familyName}`)
    }
    founderMemberId = detail.currentFounderMemberId
    if (!founderMemberId) {
      throw new Error('currentFounderMemberId missing from family detail')
    }
    ;[members, tree] = await Promise.all([
      listMembers(token, familyId),
      getPrivateTree(token, familyId)
    ])
  }

  const founderName = founderMemberId
    ? founderDisplayName(detail, members, founderMemberId)
    : '(new founder)'
  const plan = buildSeedPlan({
    founderMemberId: founderMemberId || '(pending founder id)',
    founderDisplayName: founderName
  })
  const summary = summarizeSeedPlan(plan)

  if (!APPLY) {
    console.log('[dry-run] isolated patrilineal seed plan')
    console.log(JSON.stringify({
      mode: 'dry-run',
      apiBase: CONFIG.apiBase,
      familyId,
      familyName: detail.familyName || CONFIG.familyName,
      currentFounderMemberId: founderMemberId,
      founderDisplayName: founderName,
      currentMemberCount: tree.nodes?.length || 0,
      currentRelationshipCount: tree.edges?.length || 0,
      writesAttempted: runState.writesAttempted,
      plan: summary,
      steps: plan.steps.map((step) => ({
        key: step.key,
        action: step.action,
        addType: step.addType || null,
        base: step.baseLabel || null,
        member: step.newMember.name,
        memberType: step.newMember.memberType || 'LINEAGE_MEMBER',
        relationships: step.relationships
      })),
      expectedAfterApply: plan.expected
    }, null, 2))
    if (runState.writesAttempted !== 0) {
      throw new Error(`dry-run must not write, but writesAttempted=${runState.writesAttempted}`)
    }
    return
  }

  if ((tree.edges || []).length > 0) {
    const checks = validateTree(tree, familyId, plan)
    console.log('existing isolated family already satisfies verification:', checks)
    return
  }

  runState.snapshotPath = writeSnapshot(`isolated-family-${familyId}-before`, {
    family: detail,
    tree,
    capturedAt: new Date().toISOString()
  })
  console.log('snapshot saved:', runState.snapshotPath)

  try {
    await executeSeedPlan(token, familyId, plan)
    const treeAfter = await getPrivateTree(token, familyId)
    const checks = validateTree(treeAfter, familyId, plan)
    writeSnapshot(`isolated-family-${familyId}-after`, {
      family: detail,
      tree: treeAfter,
      verification: checks,
      capturedAt: new Date().toISOString()
    })
    console.log('seed complete:', checks)
  } catch (error) {
    await cleanupPartial(token, familyId)
    throw error
  }
}

main().catch((error) => {
  console.error('[seed-isolated-patrilineal-family]', error.message)
  process.exit(1)
})
