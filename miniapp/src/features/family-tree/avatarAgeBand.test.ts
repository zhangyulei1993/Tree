import assert from 'node:assert/strict'
import test from 'node:test'

import { resolveAvatarAgeBand } from './avatarAgeBand'

const referenceDate = '2026-10-01'

test('avatar age bands use exact, non-overlapping anniversary boundaries', () => {
  assert.equal(resolveAvatarAgeBand('2020-10-01', referenceDate), 'toddler')
  assert.equal(resolveAvatarAgeBand('2019-10-01', referenceDate), 'child')
  assert.equal(resolveAvatarAgeBand('2014-10-01', referenceDate), 'child')
  assert.equal(resolveAvatarAgeBand('2013-10-01', referenceDate), 'teen')
  assert.equal(resolveAvatarAgeBand('2009-10-01', referenceDate), 'teen')
  assert.equal(resolveAvatarAgeBand('2008-10-01', referenceDate), 'youth')
  assert.equal(resolveAvatarAgeBand('1997-10-01', referenceDate), 'youth')
  assert.equal(resolveAvatarAgeBand('1996-10-01', referenceDate), 'adult')
  assert.equal(resolveAvatarAgeBand('1982-10-01', referenceDate), 'adult')
  assert.equal(resolveAvatarAgeBand('1981-10-01', referenceDate), 'mature')
  assert.equal(resolveAvatarAgeBand('1967-10-01', referenceDate), 'mature')
  assert.equal(resolveAvatarAgeBand('1966-10-01', referenceDate), 'senior')
  assert.equal(resolveAvatarAgeBand('1952-10-01', referenceDate), 'senior')
  assert.equal(resolveAvatarAgeBand('1951-10-01', referenceDate), 'elder')
  assert.equal(resolveAvatarAgeBand('1927-10-01', referenceDate), 'elder')
  assert.equal(resolveAvatarAgeBand('1926-10-01', referenceDate), 'centenarian')
})

test('avatar age does not advance before the birthday and uses a safe default', () => {
  assert.equal(resolveAvatarAgeBand('1997-10-02', referenceDate), 'youth')
  assert.equal(resolveAvatarAgeBand(undefined, referenceDate), 'adult')
  assert.equal(resolveAvatarAgeBand('2027-10-01', referenceDate), 'adult')
})
