#!/usr/bin/env node
/**
 * Read-only family tree integrity audit (all families) and optional manual fix-file apply.
 *
 * Audit mode (default):
 *   MYSQL_HOST=... MYSQL_USER=... MYSQL_PASSWORD=... MYSQL_DATABASE=tree_platform \
 *   node scripts/audit-family-tree-integrity.mjs
 *
 *   node scripts/audit-family-tree-integrity.mjs --family-id=20
 *
 * Fix mode (manual absolute-path fix file only):
 *   node scripts/audit-family-tree-integrity.mjs --fix-file=/absolute/path/family-tree-fixes.json
 */

import { readFileSync } from 'node:fs'
import { mkdirSync, writeFileSync } from 'node:fs'
import { spawnSync } from 'node:child_process'
import path from 'node:path'

import {
  auditFamilyGraph,
  buildIntegrityAuditReport,
  buildIntegrityFixTransactionSql,
  buildMembersSql,
  buildRelationshipsSql,
  buildSuggestedFixList,
  executeIntegrityFixPlan,
  parseMembersRows,
  parseRelationshipRows,
  validateIntegrityFixFile
} from './lib/audit-family-tree-integrity.mjs'

const familyIdArg = process.argv.find((arg) => arg.startsWith('--family-id='))
const familyIdFilter = familyIdArg ? Number(familyIdArg.split('=')[1]) : null
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
  const args = ['-h', CONFIG.host, '-P', CONFIG.port, '-u', CONFIG.user, CONFIG.database, '-N', '-B', '-e', sql]
  const env = { ...process.env, MYSQL_PWD: CONFIG.password }
  const result = spawnSync('mysql', args, { encoding: 'utf8', env })
  if (result.error) throw new Error(`mysql client unavailable: ${result.error.message}`)
  if (result.status !== 0) throw new Error(result.stderr?.trim() || `mysql exited with code ${result.status}`)
  return result.stdout.trim()
}

function writeReport(report, suffix) {
  const snapshotDir = path.join(process.cwd(), 'scripts', '.snapshots')
  mkdirSync(snapshotDir, { recursive: true })
  const reportPath = path.join(snapshotDir, `family-tree-integrity-${suffix}-${Date.now()}.json`)
  writeFileSync(reportPath, `${JSON.stringify(report, null, 2)}\n`, 'utf8')
  return reportPath
}

function loadGraphRows() {
  const members = parseMembersRows(mysqlQuery(buildMembersSql(familyIdFilter)))
  const edges = parseRelationshipRows(mysqlQuery(buildRelationshipsSql(familyIdFilter)))
  const familyIds = [...new Set(members.map((member) => member.familyId))].sort((a, b) => a - b)
  return { members, edges, familyIds }
}

function loadFixFile(filePath) {
  if (!path.isAbsolute(filePath)) {
    throw new Error('--fix-file must be an absolute path')
  }
  const parsed = JSON.parse(readFileSync(filePath, 'utf8'))
  return validateIntegrityFixFile(parsed)
}

function applyFixFamilies(families) {
  for (const familyFix of families) {
    if (familyIdFilter && familyFix.familyId !== familyIdFilter) {
      throw new Error(`fix familyId=${familyFix.familyId} does not match --family-id=${familyIdFilter}`)
    }
    const sql = buildIntegrityFixTransactionSql(familyFix)
    mysqlQuery(sql)
    console.log(`applied familyId=${familyFix.familyId} memberTypeUpdates=${familyFix.memberTypeUpdates.length} spouseEdges=${familyFix.spouseEdges.length}`)
  }
}

function printFamilySummary(audit) {
  console.log(`familyId=${audit.familyId} members=${audit.totals.members} issues=${audit.totals.withIssues}`)
  if (audit.unlocated.length > 0) {
    console.log('  unlocated:', audit.unlocated.map((item) => `${item.memberId}:${item.displayName}`).join(', '))
  }
  if (audit.missingSpouse.length > 0) {
    console.log('  missingSpouse:', audit.missingSpouse.map((item) => `${item.fromDisplayName}<->${item.toDisplayName}`).join(', '))
  }
  if (audit.memberTypeIssues.length > 0) {
    console.log('  memberType:', audit.memberTypeIssues.map((item) => `${item.memberId}:${item.displayName}(${item.issues.join('|')})`).join(', '))
  }
}

async function main() {
  const { members, edges, familyIds } = loadGraphRows()
  const familyAudits = familyIds.map((familyId) => auditFamilyGraph(familyId, members, edges))
  const report = buildIntegrityAuditReport(familyAudits, {
    mode: fixFilePath ? 'fix' : 'audit',
    source: 'mysql',
    database: CONFIG.database,
    familyIdFilter,
    fixFile: fixFilePath
  })
  const reportPath = writeReport(report, fixFilePath ? 'before-fix' : 'audit')
  console.log(`family-tree integrity audit: families=${report.totals.families} issues=${report.totals.withIssues}`)
  console.log(`report=${reportPath}`)

  for (const audit of familyAudits) {
    if (audit.totals.withIssues > 0) printFamilySummary(audit)
    const suggested = buildSuggestedFixList(audit)
    if (suggested.memberTypeUpdates.length > 0 || suggested.spouseEdges.length > 0) {
      const suggestionPath = writeReport({ familyId: audit.familyId, suggested }, `family-${audit.familyId}-suggested-fixes`)
      console.log(`suggestedFixes familyId=${audit.familyId} => ${suggestionPath}`)
    }
  }

  if (!fixFilePath) {
    if (report.totals.withIssues > 0) {
      console.log('manual fixes require --fix-file=/absolute/path/family-tree-fixes.json')
    }
    return
  }

  const fixes = loadFixFile(fixFilePath)
  applyFixFamilies(fixes)

  const afterRows = loadGraphRows()
  const afterAudits = afterRows.familyIds.map((familyId) => auditFamilyGraph(familyId, afterRows.members, afterRows.edges))
  const afterReport = buildIntegrityAuditReport(afterAudits, {
    mode: 'fix',
    phase: 'after-fix',
    source: 'mysql',
    database: CONFIG.database,
    familyIdFilter,
    fixFile: fixFilePath
  })
  const afterReportPath = writeReport(afterReport, 'after-fix')
  console.log(`family-tree integrity audit(after): issues=${afterReport.totals.withIssues}`)
  console.log(`report=${afterReportPath}`)
}

main().catch((error) => {
  console.error(error.message || String(error))
  process.exitCode = 1
})
