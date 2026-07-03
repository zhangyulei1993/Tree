import { readFileSync } from 'node:fs'
import path from 'node:path'

export const VALID_MEMBER_TYPES = new Set(['LINEAGE_MEMBER', 'SPOUSE', 'EXTERNAL_MEMBER'])

export function normalizeMemberType(value) {
  return String(value || '').trim().toUpperCase()
}

export function parseAuditRows(output) {
  if (!output) return []
  return output.split('\n').map((line) => {
    const [familyId, memberId, displayName, memberType, spouseEdges, parentChildEdges] = line.split('\t')
    return {
      familyId: Number(familyId),
      memberId: Number(memberId),
      displayName,
      memberType: memberType || '',
      spouseEdges: Number(spouseEdges),
      parentChildEdges: Number(parentChildEdges)
    }
  })
}

export function buildAuditSql(familyIdFilter = null) {
  const familyClause = familyIdFilter && !Number.isNaN(familyIdFilter)
    ? ` AND fm.family_id = ${familyIdFilter}`
    : ''

  return `
SELECT
  fm.family_id,
  fm.id,
  fm.display_name,
  fm.member_type,
  COALESCE(sp.cnt, 0) AS spouse_edges,
  COALESCE(pc.cnt, 0) AS parent_child_edges
FROM family_members fm
LEFT JOIN (
  SELECT family_id, member_id, COUNT(*) AS cnt
  FROM (
    SELECT family_id, from_member_id AS member_id
    FROM family_relationships
    WHERE status = 'ACTIVE' AND relationship_type = 'SPOUSE' AND deleted_at IS NULL
    UNION ALL
    SELECT family_id, to_member_id AS member_id
    FROM family_relationships
    WHERE status = 'ACTIVE' AND relationship_type = 'SPOUSE' AND deleted_at IS NULL
  ) s
  GROUP BY family_id, member_id
) sp ON sp.family_id = fm.family_id AND sp.member_id = fm.id
LEFT JOIN (
  SELECT family_id, member_id, COUNT(*) AS cnt
  FROM (
    SELECT family_id, from_member_id AS member_id
    FROM family_relationships
    WHERE status = 'ACTIVE' AND relationship_type = 'PARENT_CHILD' AND deleted_at IS NULL
    UNION ALL
    SELECT family_id, to_member_id AS member_id
    FROM family_relationships
    WHERE status = 'ACTIVE' AND relationship_type = 'PARENT_CHILD' AND deleted_at IS NULL
  ) p
  GROUP BY family_id, member_id
) pc ON pc.family_id = fm.family_id AND pc.member_id = fm.id
WHERE fm.status = 'ACTIVE'
  AND fm.deleted_at IS NULL
  ${familyClause}
ORDER BY fm.family_id, fm.id;
`.trim()
}

export function classifyAuditIssues(row) {
  const normalized = normalizeMemberType(row.memberType)
  const issues = []

  if (!normalized) {
    issues.push('EMPTY_MEMBER_TYPE')
  } else if (!VALID_MEMBER_TYPES.has(normalized)) {
    issues.push(`INVALID_MEMBER_TYPE:${row.memberType}`)
  }

  if (normalized === 'LINEAGE_MEMBER' && row.spouseEdges > 0 && row.parentChildEdges === 0) {
    issues.push('SUSPICIOUS_LINEAGE_WITH_ONLY_SPOUSE_EDGES')
  }

  if (normalized === 'SPOUSE' && row.spouseEdges === 0 && row.parentChildEdges > 0) {
    issues.push('SUSPICIOUS_SPOUSE_WITH_PARENT_CHILD_EDGES')
  }

  if (normalized === 'SPOUSE' && row.spouseEdges === 0 && row.parentChildEdges === 0) {
    issues.push('SUSPICIOUS_SPOUSE_WITHOUT_RELATIONSHIP')
  }

  if (normalized === 'EXTERNAL_MEMBER' && row.spouseEdges > 0) {
    issues.push('SUSPICIOUS_EXTERNAL_WITH_SPOUSE_EDGES')
  }

  return issues
}

export function buildAuditReport(rows, options = {}) {
  const issues = []

  for (const row of rows) {
    const rowIssues = classifyAuditIssues(row)
    if (rowIssues.length === 0) continue
    issues.push({
      familyId: row.familyId,
      memberId: row.memberId,
      displayName: row.displayName,
      memberType: row.memberType,
      spouseEdges: row.spouseEdges,
      parentChildEdges: row.parentChildEdges,
      issues: rowIssues
    })
  }

  return {
    generatedAt: new Date().toISOString(),
    mode: options.mode || 'audit',
    phase: options.phase || 'snapshot',
    database: options.database || null,
    familyIdFilter: options.familyIdFilter ?? null,
    fixFile: options.fixFile || null,
    applied: options.applied || null,
    graphVersionBumps: options.graphVersionBumps || null,
    totals: {
      scanned: rows.length,
      withIssues: issues.length
    },
    issues
  }
}

export function memberKey(familyId, memberId) {
  return `${familyId}:${memberId}`
}

export function validateFixEntry(entry, index = 0) {
  const label = `fix[${index}]`
  if (!entry || typeof entry !== 'object') {
    throw new Error(`${label} must be an object`)
  }

  const familyId = Number(entry.familyId)
  const memberId = Number(entry.memberId)
  if (!Number.isInteger(familyId) || familyId <= 0) {
    throw new Error(`${label}.familyId must be a positive integer`)
  }
  if (!Number.isInteger(memberId) || memberId <= 0) {
    throw new Error(`${label}.memberId must be a positive integer`)
  }

  const expectedType = normalizeMemberType(entry.expectedType)
  const newType = normalizeMemberType(entry.newType)
  if (!VALID_MEMBER_TYPES.has(expectedType)) {
    throw new Error(`${label}.expectedType must be LINEAGE_MEMBER, SPOUSE, or EXTERNAL_MEMBER`)
  }
  if (!VALID_MEMBER_TYPES.has(newType)) {
    throw new Error(`${label}.newType must be LINEAGE_MEMBER, SPOUSE, or EXTERNAL_MEMBER`)
  }
  if (expectedType === newType) {
    throw new Error(`${label}: expectedType must differ from newType`)
  }

  const reason = String(entry.reason || '').trim()
  if (!reason) {
    throw new Error(`${label}.reason is required`)
  }

  return {
    familyId,
    memberId,
    expectedType,
    newType,
    reason
  }
}

export function parseFamilyIdFilter(rawValue) {
  if (rawValue === undefined || rawValue === null || String(rawValue).trim() === '') {
    throw new Error('--family-id must be a positive integer')
  }
  const value = Number(rawValue)
  if (!Number.isInteger(value) || value <= 0) {
    throw new Error('--family-id must be a positive integer')
  }
  return value
}

export function validateFixFileList(fixes, familyIdFilter = null) {
  if (!Array.isArray(fixes) || fixes.length === 0) {
    throw new Error('fix file must contain at least one entry')
  }

  const normalized = fixes.map((entry, index) => validateFixEntry(entry, index))
  const seen = new Set()

  for (const fix of normalized) {
    const key = memberKey(fix.familyId, fix.memberId)
    if (seen.has(key)) {
      throw new Error(`duplicate fix entry: familyId=${fix.familyId} memberId=${fix.memberId}`)
    }
    seen.add(key)

    if (familyIdFilter !== null && fix.familyId !== familyIdFilter) {
      throw new Error(
        `fix entry familyId=${fix.familyId} does not match --family-id=${familyIdFilter}`
      )
    }
  }

  return normalized
}

export function loadFixFile(filePath, familyIdFilter = null) {
  if (!path.isAbsolute(filePath)) {
    throw new Error('--fix-file must be an absolute path')
  }
  const raw = readFileSync(filePath, 'utf8')
  const parsed = JSON.parse(raw)
  if (!Array.isArray(parsed)) {
    throw new Error('fix file must contain a JSON array')
  }
  return validateFixFileList(parsed, familyIdFilter)
}

export function buildMemberTypeFixPlan(fixes, currentMembers) {
  const updates = []

  for (const fix of fixes) {
    const key = memberKey(fix.familyId, fix.memberId)
    const current = currentMembers.get(key)
    if (!current) {
      throw new Error(`member not found: familyId=${fix.familyId} memberId=${fix.memberId}`)
    }

    const currentType = normalizeMemberType(current.memberType)
    if (currentType !== fix.expectedType) {
      throw new Error(
        `expectedType mismatch for familyId=${fix.familyId} memberId=${fix.memberId}: `
        + `expected ${fix.expectedType}, current ${currentType || '(empty)'}`
      )
    }

    updates.push({
      familyId: fix.familyId,
      memberId: fix.memberId,
      expectedType: fix.expectedType,
      newType: fix.newType,
      reason: fix.reason
    })
  }

  const graphVersionBumps = [...new Set(updates.map((item) => item.familyId))].sort((a, b) => a - b)

  return {
    updates,
    graphVersionBumps
  }
}

export const FIX_ASSERT_TABLE = 'member_type_fix_assert'
export const FIX_ASSERT_BOOTSTRAP_KEY = 1
export const FIX_ASSERT_SUCCESS_KEY_BASE = 1000

function appendRowCountAssertion(statements, successKey) {
  statements.push('SET @member_type_fix_rc := ROW_COUNT()')
  statements.push(`
INSERT INTO ${FIX_ASSERT_TABLE} (assert_key, actual_count)
VALUES (IF(@member_type_fix_rc = 1, ${successKey}, ${FIX_ASSERT_BOOTSTRAP_KEY}), @member_type_fix_rc)
`.trim())
}

export function collectAssertionSuccessKeys(sql) {
  return [...sql.matchAll(/IF\(@member_type_fix_rc = 1, (\d+), 1\)/g)].map((match) => Number(match[1]))
}

export function buildFixTransactionSql(plan) {
  if (!plan.updates.length) {
    throw new Error('fix plan must contain at least one update')
  }

  const statements = ['START TRANSACTION']
  statements.push(`
CREATE TEMPORARY TABLE ${FIX_ASSERT_TABLE} (
  assert_key INT PRIMARY KEY,
  actual_count INT NOT NULL
)
`.trim())
  statements.push(
    `INSERT INTO ${FIX_ASSERT_TABLE} (assert_key, actual_count) VALUES (${FIX_ASSERT_BOOTSTRAP_KEY}, 1)`
  )

  let assertionStep = 0
  for (const update of plan.updates) {
    assertionStep += 1
    statements.push(`
UPDATE family_members
SET member_type = '${update.newType}', updated_at = CURRENT_TIMESTAMP
WHERE family_id = ${update.familyId}
  AND id = ${update.memberId}
  AND member_type = '${update.expectedType}'
  AND status = 'ACTIVE'
  AND deleted_at IS NULL
`.trim())
    appendRowCountAssertion(statements, FIX_ASSERT_SUCCESS_KEY_BASE + assertionStep)
  }

  for (const familyId of plan.graphVersionBumps) {
    assertionStep += 1
    statements.push(`
UPDATE families
SET graph_version = graph_version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ${familyId}
`.trim())
    appendRowCountAssertion(statements, FIX_ASSERT_SUCCESS_KEY_BASE + assertionStep)
  }

  statements.push('COMMIT')
  return statements.join(';\n')
}

export async function executeFixPlan(plan, executor) {
  if (plan.updates.length === 0) {
    return { applied: 0, graphVersionBumps: [] }
  }

  await executor.begin()
  try {
    for (const update of plan.updates) {
      const affected = await executor.updateMember(update)
      if (affected !== 1) {
        throw new Error(
          `member update failed: familyId=${update.familyId} memberId=${update.memberId}`
        )
      }
    }

    for (const familyId of plan.graphVersionBumps) {
      const affected = await executor.bumpGraphVersion(familyId)
      if (affected !== 1) {
        throw new Error(`graph_version bump failed: familyId=${familyId}`)
      }
    }

    await executor.commit()
    return {
      applied: plan.updates.length,
      graphVersionBumps: plan.graphVersionBumps
    }
  } catch (error) {
    await executor.rollback()
    throw error
  }
}

export function createInMemoryExecutor(state) {
  let inTransaction = false
  let snapshot = null

  return {
    async begin() {
      if (inTransaction) throw new Error('transaction already open')
      snapshot = {
        members: new Map(state.members),
        graphVersions: new Map(state.graphVersions)
      }
      inTransaction = true
    },
    async updateMember(update) {
      if (!inTransaction) throw new Error('no open transaction')
      const key = memberKey(update.familyId, update.memberId)
      const current = state.members.get(key)
      if (!current) return 0
      if (normalizeMemberType(current.memberType) !== update.expectedType) return 0
      state.members.set(key, { ...current, memberType: update.newType })
      return 1
    },
    async bumpGraphVersion(familyId) {
      if (!inTransaction) throw new Error('no open transaction')
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
        state.graphVersions = snapshot.graphVersions
      }
      snapshot = null
      inTransaction = false
    }
  }
}
