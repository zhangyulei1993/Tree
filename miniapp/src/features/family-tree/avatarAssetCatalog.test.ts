import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'
import { portraitAssetBaseURL, resolveRegionalPortraitAsset, resolveRegionalPortraitGlassesAsset } from './portraitCatalog'

const ageBands = ['toddler', 'child', 'teen', 'youth', 'adult', 'mature', 'senior', 'elder', 'centenarian'] as const
const genders = ['male', 'female'] as const
const moduleDir = dirname(fileURLToPath(import.meta.url))
const avatarDir = join(moduleDir, '../../static/tree/avatars')

function readPngDimensions(file: Buffer) {
  assert.deepEqual([...file.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10])
  return { width: file.readUInt32BE(16), height: file.readUInt32BE(20) }
}

test('age-based avatar catalog resolves each independent variation through the static avatar service', () => {
  for (const gender of genders) for (const ageBand of ageBands) {
    for (let variant = 1; variant <= 12; variant += 1) {
      assert.match(
        resolveRegionalPortraitAsset({ theme: 'north', gender, ageBand, variant }),
        new RegExp(`^${portraitAssetBaseURL.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/age-only/base/${gender}-${ageBand}-${variant}\\.png\\?v=`)
      )
      assert.match(
        resolveRegionalPortraitGlassesAsset({ theme: 'south', gender, ageBand, variant }),
        new RegExp(`^${portraitAssetBaseURL.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/age-only/glasses/${gender}-${ageBand}-${variant}\\.png\\?v=`)
      )
    }
  }
})

test('the neutral fallback remains bundled for an unavailable external portrait', () => {
  const image = readFileSync(join(avatarDir, 'avatar-neutral.png'))
  assert.deepEqual(readPngDimensions(image), { width: 192, height: 192 })
})

test('the optional portrait atlas stays outside the mini program package', () => {
  assert.match(portraitAssetBaseURL, /^https:\/\/tapi\.bigbigboy\.cn\/api\/static\/avatars\//)
})
