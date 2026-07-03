#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'

import {
  LARGE_FAMILY_NAME,
  LARGE_FAMILY_SURNAME,
  LARGE_UNLOCATED_NAME,
  buildLargePatrilinealSeedPlan
} from './lib/large-patrilineal-seed-plan.mjs'

const APPLY = process.argv.includes('--apply')
const CONFIG = {
  apiBase: process.env.TREE_API_BASE || 'https://tapi.bigbigboy.cn/api',
  founderPhone: process.env.TREE_LARGE_FOUNDER_PHONE || '16600020101',
  adminPhone: process.env.TREE_LARGE_ADMIN_PHONE || '16600020102',
  memberPhone: process.env.TREE_LARGE_MEMBER_PHONE || '16600020103',
  password: process.env.TREE_LARGE_PASSWORD || '',
  familyName: process.env.TREE_LARGE_FAMILY_NAME || LARGE_FAMILY_NAME,
  snapshotDir: process.env.TREE_SEED_SNAPSHOT_DIR || path.join(process.cwd(), 'scripts', '.snapshots'),
  statePath: process.env.TREE_LARGE_STATE_PATH || '/tmp/tree-large-patrilineal-family-state.json'
}

const runState = {
  createdFamily: false,
  familyId: null,
  createdMemberIds: [],
  createdRelationshipIds: [],
  writesAttempted: 0
}

async function apiRequest(pathname, { method = 'GET', token, body } = {}) {
  const login = method === 'POST' && pathname === '/auth/login-phone'
  if (!APPLY && method !== 'GET' && !login) {
    throw new Error(`dry-run blocked write: ${method} ${pathname}`)
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
    const error = new Error(payload.message || `HTTP ${response.status}`)
    error.payload = payload
    throw error
  }
  if (method !== 'GET' && !login) runState.writesAttempted += 1
  return payload.data
}

async function login(phone) {
  return apiRequest('/auth/login-phone', {
    method: 'POST',
    body: { phone, password: CONFIG.password, clientType: 'WECHAT_MINI_PROGRAM' }
  })
}

async function familyTree(token, familyId) {
  return apiRequest(`/families/${familyId}/tree`, { token })
}

async function familyMembers(token, familyId) {
  return apiRequest(`/families/${familyId}/members`, { token })
}

async function updateMember(token, familyId, memberId, body) {
  return apiRequest(`/families/${familyId}/members/${memberId}`, {
    method: 'PUT',
    token,
    body
  })
}

async function createUnlocatedMember(token, familyId, item) {
  return apiRequest(`/families/${familyId}/members`, {
    method: 'POST',
    token,
    body: {
      name: item.name,
      gender: item.gender,
      birthYear: item.birthYear,
      isAlive: true,
      userBindingPolicy: 'OPTIONAL'
    }
  })
}

async function createRelative(token, familyId, baseMemberId, addType, item, memberType) {
  return apiRequest(`/families/${familyId}/relationships`, {
    method: 'POST',
    token,
    body: {
      baseMemberId,
      addType,
      newMember: {
        name: item.name,
        gender: item.gender,
        memberType,
        birthYear: item.birthYear,
        isAlive: true,
        userBindingPolicy: memberType === 'SPOUSE' ? 'NOT_REQUIRED' : 'OPTIONAL'
      },
      relationship: { parentLinkType: 'PRIMARY' }
    }
  })
}

async function bindUser(token, familyId, memberId, userId) {
  return apiRequest(`/families/${familyId}/members/${memberId}/bind-user`, {
    method: 'POST',
    token,
    body: { userId }
  })
}

async function setAdmin(token, familyId, memberId) {
  return apiRequest(`/families/${familyId}/members/${memberId}/set-admin`, {
    method: 'POST',
    token,
    body: {}
  })
}

async function deleteRelationship(token, familyId, relationshipId) {
  return apiRequest(`/families/${familyId}/relationships/${relationshipId}`, {
    method: 'DELETE',
    token,
    body: { reason: 'large patrilineal seed rollback' }
  })
}

async function deleteMember(token, familyId, memberId) {
  return apiRequest(`/families/${familyId}/members/${memberId}`, {
    method: 'DELETE',
    token,
    body: { reason: 'large patrilineal seed rollback' }
  })
}

function familyIdOf(family) {
  return family?.id ?? family?.familyId
}

function rememberMutation(result) {
  if (result?.createdMember?.memberId) {
    runState.createdMemberIds.push(result.createdMember.memberId)
  }
  for (const relationship of result?.relationships || []) {
    runState.createdRelationshipIds.push(relationship.relationshipId)
  }
}

function countCouples(tree) {
  return (tree.edges || []).filter((edge) => edge.relationshipType === 'SPOUSE').length
}

function unlocatedLineage(tree) {
  const related = new Set()
  for (const edge of tree.edges || []) {
    related.add(edge.fromMemberId)
    related.add(edge.toMemberId)
  }
  return (tree.nodes || []).filter(
    (node) => node.memberType === 'LINEAGE_MEMBER' && !related.has(node.memberId)
  )
}

function validateTree(tree, plan) {
  const unlocated = unlocatedLineage(tree)
  const result = {
    memberCount: tree.nodes?.length || 0,
    relationshipCount: tree.edges?.length || 0,
    coupleUnits: countCouples(tree),
    unlocatedCount: unlocated.length
  }
  if (result.memberCount !== plan.expected.memberCount) {
    throw new Error(`member count ${result.memberCount}, expected ${plan.expected.memberCount}`)
  }
  if (result.relationshipCount !== plan.expected.relationshipCount) {
    throw new Error(`relationship count ${result.relationshipCount}, expected ${plan.expected.relationshipCount}`)
  }
  if (result.coupleUnits !== plan.expected.coupleUnits) {
    throw new Error(`couple count ${result.coupleUnits}, expected ${plan.expected.coupleUnits}`)
  }
  if (result.unlocatedCount !== plan.expected.unlocatedCount || unlocated[0]?.displayName !== LARGE_UNLOCATED_NAME) {
    throw new Error(`unlocated members invalid: ${unlocated.map((item) => item.displayName).join(',')}`)
  }
  return result
}

function saveSnapshot(label, data) {
  fs.mkdirSync(CONFIG.snapshotDir, { recursive: true })
  const filename = path.join(CONFIG.snapshotDir, `${label}-${Date.now()}.json`)
  fs.writeFileSync(filename, JSON.stringify(data, null, 2), 'utf8')
  return filename
}

async function resolveFamily(token) {
  const families = await apiRequest('/families', { token })
  const existing = families.find((family) => family.familyName === CONFIG.familyName)
  if (existing) return existing
  if (!APPLY) return { id: 'NEW', familyName: CONFIG.familyName }
  const created = await apiRequest('/families', {
    method: 'POST',
    token,
    body: {
      familyName: CONFIG.familyName,
      surname: LARGE_FAMILY_SURNAME,
      founderGender: 'MALE'
    }
  })
  runState.createdFamily = true
  return created
}

async function executePlan(token, familyId, plan) {
  const refs = new Map([['$founder', Number(plan.founderMemberId)]])

  await updateMember(token, familyId, plan.founderMemberId, plan.founderPatch)

  for (const family of plan.families) {
    const baseId = refs.get(family.baseRef)
    if (!baseId) throw new Error(`missing base ref ${family.baseRef}`)
    const firstChild = family.children[0]
    console.log(`create family unit ${family.key}: ${firstChild.name}`)
    const firstResult = await createRelative(
      token,
      familyId,
      baseId,
      'ADD_CHILD',
      firstChild,
      'LINEAGE_MEMBER'
    )
    rememberMutation(firstResult)
    const firstChildId = firstResult.createdMember.memberId
    refs.set(firstChild.key, firstChildId)

    const parentType = family.baseGender === 'MALE' ? 'ADD_MOTHER' : 'ADD_FATHER'
    const spouseResult = await createRelative(
      token,
      familyId,
      firstChildId,
      parentType,
      family.spouse,
      'SPOUSE'
    )
    rememberMutation(spouseResult)

    for (const sibling of family.children.slice(1)) {
      const siblingResult = await createRelative(
        token,
        familyId,
        firstChildId,
        'ADD_SIBLING',
        sibling,
        'LINEAGE_MEMBER'
      )
      rememberMutation(siblingResult)
      refs.set(sibling.key, siblingResult.createdMember.memberId)
    }
  }

  const unlocated = await createUnlocatedMember(token, familyId, plan.unlocated)
  runState.createdMemberIds.push(unlocated.memberId)
  refs.set(plan.unlocated.key, unlocated.memberId)
  return refs
}

async function bindAccounts(token, familyId, sessions, plan) {
  const members = await familyMembers(token, familyId)
  const byName = new Map(members.map((member) => [member.name, member]))
  const adminMember = byName.get(plan.accountTargets.familyAdmin)
  const plainMember = byName.get(plan.accountTargets.member)
  if (!adminMember || !plainMember) throw new Error('account target members missing')

  if (!adminMember.boundUserId) {
    await bindUser(token, familyId, adminMember.memberId, sessions.admin.user.id)
  }
  const refreshedAdmin = (await familyMembers(token, familyId))
    .find((member) => member.memberId === adminMember.memberId)
  if (refreshedAdmin?.boundFamilyRole !== 'FAMILY_ADMIN') {
    await setAdmin(token, familyId, adminMember.memberId)
  }
  if (!plainMember.boundUserId) {
    await bindUser(token, familyId, plainMember.memberId, sessions.member.user.id)
  }

  return {
    founder: { userId: sessions.founder.user.id, memberId: plan.founderMemberId, role: 'FOUNDER' },
    familyAdmin: { userId: sessions.admin.user.id, memberId: adminMember.memberId, role: 'FAMILY_ADMIN' },
    member: { userId: sessions.member.user.id, memberId: plainMember.memberId, role: 'MEMBER' }
  }
}

async function rollback(token, familyId) {
  console.warn('rolling back large seed mutations')
  for (const id of [...new Set(runState.createdRelationshipIds)].reverse()) {
    try {
      await deleteRelationship(token, familyId, id)
    } catch (error) {
      console.warn(`relationship rollback failed ${id}: ${error.message}`)
    }
  }
  for (const id of [...new Set(runState.createdMemberIds)].reverse()) {
    try {
      await deleteMember(token, familyId, id)
    } catch (error) {
      console.warn(`member rollback failed ${id}: ${error.message}`)
    }
  }
}

async function main() {
  if (!CONFIG.password) throw new Error('Set TREE_LARGE_PASSWORD before running')

  const sessions = {
    founder: await login(CONFIG.founderPhone),
    admin: await login(CONFIG.adminPhone),
    member: await login(CONFIG.memberPhone)
  }
  const token = sessions.founder.accessToken
  const family = await resolveFamily(token)
  const familyId = familyIdOf(family)
  if (Number(familyId) === 22) throw new Error('refusing to operate on familyId=22')
  runState.familyId = familyId

  if (!APPLY) {
    const plan = buildLargePatrilinealSeedPlan(family.currentFounderMemberId || 'PENDING')
    console.log(JSON.stringify({
      mode: 'dry-run',
      apiBase: CONFIG.apiBase,
      familyId,
      familyName: CONFIG.familyName,
      writesAttempted: runState.writesAttempted,
      expected: plan.expected,
      accountTargets: plan.accountTargets
    }, null, 2))
    if (runState.writesAttempted !== 0) throw new Error('dry-run attempted writes')
    return
  }

  const detail = await apiRequest(`/families/${familyId}`, { token })
  const founderMemberId = detail.currentFounderMemberId
  if (!founderMemberId) throw new Error('currentFounderMemberId missing')
  const plan = buildLargePatrilinealSeedPlan(founderMemberId)
  const before = await familyTree(token, familyId)

  if ((before.nodes || []).length === plan.expected.memberCount) {
    const verification = validateTree(before, plan)
    const accounts = await bindAccounts(token, familyId, sessions, plan)
    fs.writeFileSync(CONFIG.statePath, JSON.stringify({ familyId, familyName: CONFIG.familyName, verification, accounts }, null, 2))
    console.log('existing large family verified', { familyId, verification, accounts })
    return
  }
  if ((before.edges || []).length > 0 || (before.nodes || []).length !== 1) {
    throw new Error(`existing family is partial: members=${before.nodes?.length || 0}, edges=${before.edges?.length || 0}`)
  }

  const snapshot = saveSnapshot(`large-family-${familyId}-before`, { family: detail, tree: before })
  console.log(`snapshot saved ${snapshot}`)
  try {
    await executePlan(token, familyId, plan)
    const after = await familyTree(token, familyId)
    const verification = validateTree(after, plan)
    const accounts = await bindAccounts(token, familyId, sessions, plan)
    const output = {
      familyId: Number(familyId),
      familyName: CONFIG.familyName,
      graphVersion: after.graphVersion,
      verification,
      accounts,
      generatedAt: new Date().toISOString()
    }
    fs.writeFileSync(CONFIG.statePath, JSON.stringify(output, null, 2))
    saveSnapshot(`large-family-${familyId}-after`, { ...output, tree: after })
    console.log('large family seed complete', output)
  } catch (error) {
    await rollback(token, familyId)
    throw error
  }
}

main().catch((error) => {
  console.error('[seed-large-patrilineal-family]', error.message)
  if (error.payload) console.error(JSON.stringify(error.payload))
  process.exit(1)
})
