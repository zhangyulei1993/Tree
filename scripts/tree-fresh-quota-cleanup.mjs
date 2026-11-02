#!/usr/bin/env node
/**
 * Remove fresh quota E2E artifacts from local /tmp after wechat reauth cancel.
 */

import fs from 'node:fs'
import process from 'node:process'

import { resolveApiBase, sessionFilePath } from './tree-quota-e2e-lib.mjs'
import { teardownSessionArtifact } from './tree-quota-e2e-harness.mjs'

function parseArgs(argv) {
  const args = { sessionFile: '', runId: '', allTmp: false }
  for (let i = 0; i < argv.length; i += 1) {
    if (argv[i] === '--session-file') args.sessionFile = argv[i + 1] || ''
    if (argv[i] === '--run-id') args.runId = argv[i + 1] || ''
    if (argv[i] === '--all-tmp') args.allTmp = true
  }
  return args
}

async function cleanupSessionFile(file) {
  if (!file || !fs.existsSync(file)) {
    console.warn(`[tree-fresh-quota-cleanup] session file not found: ${file}`)
    return
  }
  const session = JSON.parse(fs.readFileSync(file, 'utf8'))
  session.apiBase = session.apiBase || resolveApiBase()
  await teardownSessionArtifact(session, file)
  console.log(`[tree-fresh-quota-cleanup] cancelled and removed ${file}`)
}

async function main() {
  const args = parseArgs(process.argv.slice(2))
  if (args.allTmp) {
    const files = fs.readdirSync('/tmp').filter((name) => name.startsWith('tree-quota-e2e-session-') && name.endsWith('.json'))
    for (const name of files) {
      await cleanupSessionFile(`/tmp/${name}`)
    }
    return
  }

  const file = args.sessionFile || (args.runId ? sessionFilePath(args.runId) : '')
  if (!file) {
    console.error('usage: --session-file <path> | --run-id <id> | --all-tmp')
    process.exit(1)
  }
  await cleanupSessionFile(file)
}

main().catch((error) => {
  console.error('[tree-fresh-quota-cleanup] failed:', error.message)
  process.exit(1)
})
