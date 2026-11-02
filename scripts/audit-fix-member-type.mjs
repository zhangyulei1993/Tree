#!/usr/bin/env node
/**
 * Audit family_members.member_type and apply only explicit manual fix files.
 *
 * Audit mode never guesses member types from relationship edges.
 *
 * Usage:
 *   MYSQL_HOST=127.0.0.1 MYSQL_USER=tree_user MYSQL_PASSWORD=tree_pass \
 *   MYSQL_DATABASE=tree_platform node scripts/audit-fix-member-type.mjs
 *
 *   node scripts/audit-fix-member-type.mjs --family-id=22
 *   node scripts/audit-fix-member-type.mjs --fix-file=/absolute/path/member-type-fixes.json
 */

import { spawnSync } from 'node:child_process'
import { mkdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'

import {
  buildAuditReport,
  buildAuditSql,
  buildFixTransactionSql,
  buildMemberTypeFixPlan,
  loadFixFile,
  memberKey,
  parseAuditRows,
  parseFamilyIdFilter
} from './lib/audit-fix-member-type.mjs'

const familyIdArg = process.argv.find((arg) => arg.startsWith('--family-id='))
const familyIdFilter = familyIdArg ? parseFamilyIdFilter(familyIdArg.split('=')[1]) : null
const fixFileArg = process.argv.find((arg) => arg.startsWith('--fix-file='))
const fixFilePath = fixFileArg ? fixFileArg.slice('--fix-file='.length) : null

const CONFIG = {
  host: process.env.MYSQL_HOST || '127.0.0.1',
  port: process.env.MYSQL_PORT || '3306',
  user: process.env.MYSQL_USER || 'tree_user',
  password: process.env.MYSQL_PASSWORD || 'tree_pass',
  database: process.env.MYSQL_DATABASE || 'tree_platform'
}

function mysqlQuery(sql) {
  const args = [
    '-h', CONFIG.host,
    '-P', CONFIG.port,
    '-u', CONFIG.user,
    CONFIG.database,
    '-N',
    '-B',
    '-e',
    sql
  ]
  const env = { ...process.env, MYSQL_PWD: CONFIG.password }
  const result = spawnSync('mysql', args, { encoding: 'utf8', env })
  if (result.error) {
    throw new Error(`mysql client unavailable: ${result.error.message}`)
  }
  if (result.status !== 0) {
    throw new Error(result.stderr?.trim() || `mysql exited with code ${result.status}`)
  }
  return result.stdout.trim()
}

function writeAuditReport(report) {
  const snapshotDir = path.join(process.cwd(), 'scripts', '.snapshots')
  mkdirSync(snapshotDir, { recursive: true })
  const reportPath = path.join(snapshotDir, `member-type-audit-${Date.now()}.json`)
  writeFileSync(reportPath, `${JSON.stringify(report, null, 2)}\n`, 'utf8')
  return reportPath
}

function queryAuditRows() {
  return parseAuditRows(mysqlQuery(buildAuditSql(familyIdFilter)))
}

function queryCurrentMembers(fixes) {
  const tuples = fixes.map((fix) => `(${fix.familyId}, ${fix.memberId})`).join(', ')
  const sql = `
SELECT family_id, id, display_name, member_type
FROM family_members
WHERE status = 'ACTIVE'
  AND deleted_at IS NULL
  AND (family_id, id) IN (${tuples});
`.trim()
  const output = mysqlQuery(sql)
  const currentMembers = new Map()
  if (!output) return currentMembers

  for (const line of output.split('\n')) {
    const [familyId, memberId, displayName, memberType] = line.split('\t')
    currentMembers.set(memberKey(Number(familyId), Number(memberId)), {
      familyId: Number(familyId),
      memberId: Number(memberId),
      displayName,
      memberType: memberType || ''
    })
  }
  return currentMembers
}

function executeFixTransaction(plan) {
  const sql = buildFixTransactionSql(plan)
  mysqlQuery(sql)
}

async function main() {
  const beforeRows = queryAuditRows()
  const beforeReport = buildAuditReport(beforeRows, {
    mode: fixFilePath ? 'fix' : 'audit',
    phase: fixFilePath ? 'before-fix' : 'snapshot',
    database: CONFIG.database,
    familyIdFilter,
    fixFile: fixFilePath
  })
  const beforeReportPath = writeAuditReport(beforeReport)

  console.log(`member_type audit(before): scanned=${beforeReport.totals.scanned} issues=${beforeReport.totals.withIssues}`)
  console.log(`report=${beforeReportPath}`)

  if (!fixFilePath) {
    if (beforeReport.totals.withIssues > 0) {
      console.log('manual fixes require --fix-file=/absolute/path/member-type-fixes.json')
    }
    return
  }

  const fixes = loadFixFile(fixFilePath, familyIdFilter)
  const currentMembers = queryCurrentMembers(fixes)
  const plan = buildMemberTypeFixPlan(fixes, currentMembers)

  executeFixTransaction(plan)
  console.log(`applied=${plan.updates.length} graphVersionBumps=${plan.graphVersionBumps.join(',')}`)

  const afterRows = queryAuditRows()
  const afterReport = buildAuditReport(afterRows, {
    mode: 'fix',
    phase: 'after-fix',
    database: CONFIG.database,
    familyIdFilter,
    fixFile: fixFilePath,
    applied: plan.updates,
    graphVersionBumps: plan.graphVersionBumps
  })
  const afterReportPath = writeAuditReport(afterReport)
  console.log(`member_type audit(after): scanned=${afterReport.totals.scanned} issues=${afterReport.totals.withIssues}`)
  console.log(`report=${afterReportPath}`)
}

main().catch((error) => {
  console.error(error.message || String(error))
  process.exitCode = 1
})
