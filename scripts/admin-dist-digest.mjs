import { createHash } from 'node:crypto'
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative, sep } from 'node:path'

export const ADMIN_BUILD_META_FILE = 'admin-build-meta.json'

function toPosixPath(path) {
  return path.split(sep).join('/')
}

export function listDistFiles(distDir) {
  if (!existsSync(distDir)) {
    return []
  }

  const files = []

  function walk(currentDir) {
    for (const name of readdirSync(currentDir).sort()) {
      const fullPath = join(currentDir, name)
      const stat = statSync(fullPath)
      if (stat.isDirectory()) {
        walk(fullPath)
        continue
      }
      files.push(toPosixPath(relative(distDir, fullPath)))
    }
  }

  walk(distDir)
  return files.sort()
}

export function computeDistDigest(distDir, excludeFiles = [ADMIN_BUILD_META_FILE]) {
  const exclude = new Set(excludeFiles)
  const hash = createHash('sha256')

  for (const file of listDistFiles(distDir)) {
    if (exclude.has(file)) {
      continue
    }
    const content = readFileSync(join(distDir, file))
    hash.update(file)
    hash.update('\0')
    hash.update(content)
    hash.update('\0')
  }

  return hash.digest('hex')
}
