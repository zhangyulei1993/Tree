#!/usr/bin/env node
/**
 * Fresh WeChat-only quota E2E session generator (local only).
 *
 * Usage:
 *   node scripts/tree-fresh-quota-e2e-session.mjs
 *   node scripts/tree-fresh-quota-e2e-session.mjs --run-id my-run-01
 */

import process from 'node:process'

import {
  createFreshQuotaSession,
  parseRunId,
  redactToken,
  resolveApiBase,
  writeSessionFile
} from './tree-quota-e2e-lib.mjs'

async function main() {
  const runId = parseRunId(process.argv.slice(2))
  const apiBase = resolveApiBase()

  console.log(`[tree-fresh-quota-e2e] runId=${runId}`)
  console.log(`[tree-fresh-quota-e2e] api=${apiBase}`)

  const session = await createFreshQuotaSession({ apiBase, runId })
  const outFile = writeSessionFile(session)

  console.log(`[tree-fresh-quota-e2e] session written: ${outFile}`)
  console.log(`[tree-fresh-quota-e2e] userId=${session.user.id} trust=${session.capabilities.trustTier} token=${redactToken(session.token)}`)
  console.log(`[tree-fresh-quota-e2e] capabilities owned=${session.capabilities.usage.ownedFamilies}/${session.capabilities.limits.maxOwnedFamilies} joined=${session.capabilities.usage.joinedFamilies}/${session.capabilities.limits.maxJoinedFamilies} members=${session.capabilities.limits.maxMembersPerOwnedFamily}`)
  console.log('[tree-fresh-quota-e2e] Playwright: page.addInitScript(() => { eval(session.playwrightInitScript) })')
}

main().catch((error) => {
  console.error('[tree-fresh-quota-e2e] failed:', error.message)
  if (error.code) console.error('[tree-fresh-quota-e2e] code:', error.code)
  process.exit(1)
})
