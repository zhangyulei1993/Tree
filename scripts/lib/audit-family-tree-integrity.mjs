import { classifyAuditIssues, normalizeMemberType } from './audit-fix-member-type.mjs'

export function spousePairKey(memberIdA, memberIdB) {
  const a = Number(memberIdA)
  const b = Number(memberIdB)
  return a < b ? `${a}:${b}` : `${b}:${a}`
}

export function indexSpousePairs(edges) {
  const pairs = new Set()
  for (const edge of edges) {
    if (edge.relationshipType !== 'SPOUSE' || edge.status !== 'ACTIVE') continue
    pairs.add(spousePairKey(edge.fromMemberId, edge.toMemberId))
  }
  return pairs
}

export function indexActiveEdgesByMember(edges) {
  const map = new Map()
  for (const edge of edges) {
    if (edge.status !== 'ACTIVE') continue
    for (const memberId of [edge.fromMemberId, edge.toMemberId]) {
      if (!map.has(memberId)) map.set(memberId, [])
      map.get(memberId).push(edge)
    }
  }
  return map
}

export function findUnlocatedMembers(members, edges) {
  const edgeIndex = indexActiveEdgesByMember(edges)
  const issues = []
  for (const member of members) {
    if (member.status && member.status !== 'ACTIVE') continue
    const related = edgeIndex.get(member.memberId) || []
    if (related.length === 0) {
      issues.push({
        type: 'UNLOCATED_MEMBER',
        familyId: member.familyId,
        memberId: member.memberId,
        displayName: member.displayName,
        memberType: member.memberType,
        gender: member.gender || null
      })
    }
  }
  return issues
}

export function findMissingSpouseEdges(members, edges) {
  const memberById = new Map(members.map((member) => [member.memberId, member]))
  const spousePairs = indexSpousePairs(edges)
  const parentChild = edges.filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.status === 'ACTIVE')
  const childParents = new Map()

  for (const edge of parentChild) {
    if (!childParents.has(edge.toMemberId)) childParents.set(edge.toMemberId, new Set())
    childParents.get(edge.toMemberId).add(edge.fromMemberId)
  }

  const issues = []
  const seen = new Set()

  for (const [childId, parentIds] of childParents.entries()) {
    const parents = [...parentIds]
    if (parents.length < 2) continue
    for (let i = 0; i < parents.length; i += 1) {
      for (let j = i + 1; j < parents.length; j += 1) {
        const fromMemberId = parents[i]
        const toMemberId = parents[j]
        const pair = spousePairKey(fromMemberId, toMemberId)
        if (spousePairs.has(pair) || seen.has(pair)) continue
        seen.add(pair)
        const fromMember = memberById.get(fromMemberId)
        const toMember = memberById.get(toMemberId)
        issues.push({
          type: 'MISSING_SPOUSE_BETWEEN_PARENTS',
          familyId: fromMember?.familyId || toMember?.familyId || null,
          childMemberId: childId,
          fromMemberId,
          toMemberId,
          fromDisplayName: fromMember?.displayName || String(fromMemberId),
          toDisplayName: toMember?.displayName || String(toMemberId)
        })
      }
    }
  }

  return issues
}

export function findMemberTypeIssues(members, edges) {
  const edgeStats = new Map()
  for (const member of members) {
    edgeStats.set(member.memberId, { spouseEdges: 0, parentChildEdges: 0 })
  }
  for (const edge of edges) {
    if (edge.status !== 'ACTIVE') continue
    if (edge.relationshipType === 'SPOUSE') {
      for (const memberId of [edge.fromMemberId, edge.toMemberId]) {
        const stats = edgeStats.get(memberId)
        if (stats) stats.spouseEdges += 1
      }
    }
    if (edge.relationshipType === 'PARENT_CHILD') {
      for (const memberId of [edge.fromMemberId, edge.toMemberId]) {
        const stats = edgeStats.get(memberId)
        if (stats) stats.parentChildEdges += 1
      }
    }
  }

  const issues = []
  for (const member of members) {
    const stats = edgeStats.get(member.memberId) || { spouseEdges: 0, parentChildEdges: 0 }
    const rowIssues = classifyAuditIssues({
      familyId: member.familyId,
      memberId: member.memberId,
      displayName: member.displayName,
      memberType: member.memberType,
      spouseEdges: stats.spouseEdges,
      parentChildEdges: stats.parentChildEdges
    })
    if (rowIssues.length === 0) continue
    issues.push({
      type: 'SUSPICIOUS_MEMBER_TYPE',
      familyId: member.familyId,
      memberId: member.memberId,
      displayName: member.displayName,
      memberType: member.memberType,
      spouseEdges: stats.spouseEdges,
      parentChildEdges: stats.parentChildEdges,
      issues: rowIssues
    })
  }
  return issues
}

export function auditFamilyGraph(familyId, members, edges) {
  const scopedMembers = members.filter((member) => member.familyId === familyId)
  const scopedEdges = edges.filter((edge) => edge.familyId === familyId)
  const unlocated = findUnlocatedMembers(scopedMembers, scopedEdges)
  const missingSpouse = findMissingSpouseEdges(scopedMembers, scopedEdges)
  const memberTypeIssues = findMemberTypeIssues(scopedMembers, scopedEdges)

  return {
    familyId,
    totals: {
      members: scopedMembers.length,
      edges: scopedEdges.length,
      unlocated: unlocated.length,
      missingSpouse: missingSpouse.length,
      memberTypeIssues: memberTypeIssues.length,
      withIssues: unlocated.length + missingSpouse.length + memberTypeIssues.length
    },
    unlocated,
    missingSpouse,
    memberTypeIssues
  }
}

export function buildSuggestedFixList(familyAudit) {
  const memberTypeUpdates = []
  const spouseEdges = []
  const seenMember = new Set()
  const seenSpouse = new Set()

  for (const issue of familyAudit.memberTypeIssues) {
    if (!issue.issues.includes('SUSPICIOUS_LINEAGE_WITH_ONLY_SPOUSE_EDGES')) continue
    const key = `${issue.familyId}:${issue.memberId}`
    if (seenMember.has(key)) continue
    seenMember.add(key)
    memberTypeUpdates.push({
      familyId: issue.familyId,
      memberId: issue.memberId,
      displayName: issue.displayName,
      expectedType: normalizeMemberType(issue.memberType),
      suggestedType: 'SPOUSE',
      reason: `人工确认：${issue.displayName} 仅有配偶关系边，建议 memberType 调整为 SPOUSE`
    })
  }

  for (const issue of familyAudit.missingSpouse) {
    const key = spousePairKey(issue.fromMemberId, issue.toMemberId)
    if (seenSpouse.has(key)) continue
    seenSpouse.add(key)
    spouseEdges.push({
      familyId: issue.familyId,
      fromMemberId: issue.fromMemberId,
      toMemberId: issue.toMemberId,
      fromDisplayName: issue.fromDisplayName,
      toDisplayName: issue.toDisplayName,
      reason: `人工确认：为共同父母 ${issue.fromDisplayName} 与 ${issue.toDisplayName} 补 SPOUSE 边`
    })
  }

  return {
    familyId: familyAudit.familyId,
    memberTypeUpdates,
    spouseEdges
  }
}

export function buildIntegrityAuditReport(familyAudits, options = {}) {
  const families = familyAudits.map((audit) => ({
    ...audit,
    suggestedFixes: buildSuggestedFixList(audit)
  }))
  const totals = families.reduce((acc, audit) => {
    acc.families += 1
    acc.withIssues += audit.totals.withIssues
    return acc
  }, { families: 0, withIssues: 0 })

  return {
    generatedAt: new Date().toISOString(),
    mode: options.mode || 'audit',
    source: options.source || 'mysql',
    database: options.database || null,
    familyIdFilter: options.familyIdFilter ?? null,
    fixFile: options.fixFile || null,
    totals,
    families
  }
}

export function parseMembersRows(output) {
  if (!output) return []
  return output.split('\n').map((line) => {
    const [familyId, memberId, displayName, memberType, gender, status] = line.split('\t')
    return {
      familyId: Number(familyId),
      memberId: Number(memberId),
      displayName,
      memberType: memberType || '',
      gender: gender || '',
      status: status || 'ACTIVE'
    }
  })
}

export function parseRelationshipRows(output) {
  if (!output) return []
  return output.split('\n').map((line) => {
    const [familyId, relationshipId, fromMemberId, toMemberId, relationshipType, status] = line.split('\t')
    return {
      familyId: Number(familyId),
      relationshipId: Number(relationshipId),
      fromMemberId: Number(fromMemberId),
      toMemberId: Number(toMemberId),
      relationshipType,
      status: status || 'ACTIVE'
    }
  })
}

export function buildMembersSql(familyIdFilter = null) {
  const clause = familyIdFilter ? ` AND fm.family_id = ${familyIdFilter}` : ''
  return `
SELECT fm.family_id, fm.id, fm.display_name, fm.member_type, fm.gender, fm.status
FROM family_members fm
JOIN families f ON f.id = fm.family_id AND f.deleted_at IS NULL AND f.status != 'DISSOLVED'
WHERE fm.deleted_at IS NULL${clause}
ORDER BY fm.family_id, fm.id;
`.trim()
}

export function buildRelationshipsSql(familyIdFilter = null) {
  const clause = familyIdFilter ? ` AND fr.family_id = ${familyIdFilter}` : ''
  return `
SELECT fr.family_id, fr.id, fr.from_member_id, fr.to_member_id, fr.relationship_type, fr.status
FROM family_relationships fr
JOIN families f ON f.id = fr.family_id AND f.deleted_at IS NULL AND f.status != 'DISSOLVED'
WHERE fr.deleted_at IS NULL${clause}
ORDER BY fr.family_id, fr.id;
`.trim()
}

export function validateIntegrityFixFile(parsed) {
  if (!parsed || typeof parsed !== 'object') {
    throw new Error('fix file must be a JSON object')
  }
  if (!Array.isArray(parsed.families) || parsed.families.length === 0) {
    throw new Error('fix file must contain families array')
  }
  return parsed.families.map((family, index) => validateFamilyFixEntry(family, index))
}

export function validateFamilyFixEntry(entry, index = 0) {
  const label = `families[${index}]`
  const familyId = Number(entry?.familyId)
  if (!Number.isInteger(familyId) || familyId <= 0) {
    throw new Error(`${label}.familyId must be a positive integer`)
  }

  const memberTypeUpdates = Array.isArray(entry.memberTypeUpdates) ? entry.memberTypeUpdates : []
  const spouseEdges = Array.isArray(entry.spouseEdges) ? entry.spouseEdges : []
  if (memberTypeUpdates.length === 0 && spouseEdges.length === 0) {
    throw new Error(`${label} must include memberTypeUpdates and/or spouseEdges`)
  }

  const normalizedMemberUpdates = memberTypeUpdates.map((item, itemIndex) => {
    const memberId = Number(item.memberId)
    const expectedType = normalizeMemberType(item.expectedType)
    const newType = normalizeMemberType(item.newType)
    const reason = String(item.reason || '').trim()
    if (!Number.isInteger(memberId) || memberId <= 0) {
      throw new Error(`${label}.memberTypeUpdates[${itemIndex}].memberId invalid`)
    }
    if (!expectedType || !newType || expectedType === newType) {
      throw new Error(`${label}.memberTypeUpdates[${itemIndex}] expectedType/newType invalid`)
    }
    if (!reason) throw new Error(`${label}.memberTypeUpdates[${itemIndex}].reason is required`)
    return { familyId, memberId, expectedType, newType, reason }
  })

  const normalizedSpouseEdges = spouseEdges.map((item, itemIndex) => {
    const fromMemberId = Number(item.fromMemberId)
    const toMemberId = Number(item.toMemberId)
    const reason = String(item.reason || '').trim()
    if (!Number.isInteger(fromMemberId) || !Number.isInteger(toMemberId) || fromMemberId <= 0 || toMemberId <= 0) {
      throw new Error(`${label}.spouseEdges[${itemIndex}] member ids invalid`)
    }
    if (fromMemberId === toMemberId) {
      throw new Error(`${label}.spouseEdges[${itemIndex}] cannot be self-spouse`)
    }
    if (!reason) throw new Error(`${label}.spouseEdges[${itemIndex}].reason is required`)
    return { familyId, fromMemberId, toMemberId, reason }
  })

  return {
    familyId,
    memberTypeUpdates: normalizedMemberUpdates,
    spouseEdges: normalizedSpouseEdges
  }
}

export function buildIntegrityFixTransactionSql(familyFix) {
  const statements = ['START TRANSACTION']
  let step = 0

  for (const update of familyFix.memberTypeUpdates) {
    step += 1
    statements.push(`
UPDATE family_members
SET member_type = '${update.newType}', updated_at = CURRENT_TIMESTAMP
WHERE family_id = ${update.familyId}
  AND id = ${update.memberId}
  AND member_type = '${update.expectedType}'
  AND status = 'ACTIVE'
  AND deleted_at IS NULL
`.trim())
    statements.push('SET @integrity_fix_rc := ROW_COUNT()')
    statements.push(`INSERT INTO integrity_fix_assert (assert_key, actual_count) VALUES (${1000 + step}, @integrity_fix_rc)`)
  }

  for (const edge of familyFix.spouseEdges) {
    step += 1
    const fromId = Math.min(edge.fromMemberId, edge.toMemberId)
    const toId = Math.max(edge.fromMemberId, edge.toMemberId)
    statements.push(`
INSERT INTO family_relationships (
  family_id, from_member_id, to_member_id, relationship_type, status, created_at, updated_at
)
SELECT ${edge.familyId}, ${fromId}, ${toId}, 'SPOUSE', 'ACTIVE', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM family_relationships
  WHERE family_id = ${edge.familyId}
    AND status = 'ACTIVE'
    AND deleted_at IS NULL
    AND relationship_type = 'SPOUSE'
    AND (
      (from_member_id = ${fromId} AND to_member_id = ${toId})
      OR (from_member_id = ${toId} AND to_member_id = ${fromId})
    )
)
`.trim())
    statements.push('SET @integrity_fix_rc := ROW_COUNT()')
    statements.push(`INSERT INTO integrity_fix_assert (assert_key, actual_count) VALUES (${1000 + step}, @integrity_fix_rc)`)
  }

  if (familyFix.memberTypeUpdates.length > 0 || familyFix.spouseEdges.length > 0) {
    step += 1
    statements.push(`
UPDATE families
SET graph_version = graph_version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ${familyFix.familyId}
`.trim())
    statements.push('SET @integrity_fix_rc := ROW_COUNT()')
    statements.push(`INSERT INTO integrity_fix_assert (assert_key, actual_count) VALUES (${1000 + step}, @integrity_fix_rc)`)
  }

  statements.unshift(`
CREATE TEMPORARY TABLE integrity_fix_assert (
  assert_key INT PRIMARY KEY,
  actual_count INT NOT NULL
)
`.trim())
  statements.splice(1, 0, 'INSERT INTO integrity_fix_assert (assert_key, actual_count) VALUES (1, 1)')
  statements.push('COMMIT')
  return statements.join(';\n')
}

export async function executeIntegrityFixPlan(familyFix, executor) {
  await executor.begin()
  try {
    for (const update of familyFix.memberTypeUpdates) {
      const affected = await executor.updateMemberType(update)
      if (affected !== 1) {
        throw new Error(`member type update failed: familyId=${update.familyId} memberId=${update.memberId}`)
      }
    }
    for (const edge of familyFix.spouseEdges) {
      const affected = await executor.addSpouseEdge(edge)
      if (affected !== 1) {
        throw new Error(`spouse edge insert failed: familyId=${edge.familyId} ${edge.fromMemberId}<->${edge.toMemberId}`)
      }
    }
    if (familyFix.memberTypeUpdates.length > 0 || familyFix.spouseEdges.length > 0) {
      const affected = await executor.bumpGraphVersion(familyFix.familyId)
      if (affected !== 1) {
        throw new Error(`graph_version bump failed: familyId=${familyFix.familyId}`)
      }
    }
    await executor.commit()
  } catch (error) {
    await executor.rollback()
    throw error
  }
}

export function createIntegrityInMemoryExecutor(state) {
  let inTransaction = false
  let snapshot = null

  return {
    async begin() {
      if (inTransaction) throw new Error('transaction already open')
      snapshot = {
        members: new Map(state.members),
        edges: [...state.edges],
        graphVersions: new Map(state.graphVersions)
      }
      inTransaction = true
    },
    async updateMemberType(update) {
      const key = `${update.familyId}:${update.memberId}`
      const current = state.members.get(key)
      if (!current || normalizeMemberType(current.memberType) !== update.expectedType) return 0
      state.members.set(key, { ...current, memberType: update.newType })
      return 1
    },
    async addSpouseEdge(edge) {
      const pair = spousePairKey(edge.fromMemberId, edge.toMemberId)
      const exists = state.edges.some((item) => item.familyId === edge.familyId
        && item.relationshipType === 'SPOUSE'
        && item.status === 'ACTIVE'
        && spousePairKey(item.fromMemberId, item.toMemberId) === pair)
      if (exists) return 0
      state.edges.push({
        familyId: edge.familyId,
        relationshipId: state.edges.length + 1,
        fromMemberId: Math.min(edge.fromMemberId, edge.toMemberId),
        toMemberId: Math.max(edge.fromMemberId, edge.toMemberId),
        relationshipType: 'SPOUSE',
        status: 'ACTIVE'
      })
      return 1
    },
    async bumpGraphVersion(familyId) {
      if (!state.graphVersions.has(familyId)) return 0
      state.graphVersions.set(familyId, state.graphVersions.get(familyId) + 1)
      return 1
    },
    async commit() {
      snapshot = null
      inTransaction = false
    },
    async rollback() {
      if (snapshot) {
        state.members = snapshot.members
        state.edges = snapshot.edges
        state.graphVersions = snapshot.graphVersions
      }
      snapshot = null
      inTransaction = false
    }
  }
}
