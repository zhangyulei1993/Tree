import assert from 'node:assert/strict'
import test from 'node:test'

const storage = new Map<string, unknown>()

globalThis.uni = {
  getStorageSync(key: string) {
    return storage.get(key)
  },
  setStorageSync(key: string, value: unknown) {
    storage.set(key, value)
  },
  removeStorageSync(key: string) {
    storage.delete(key)
  }
} as typeof uni

const guide = await import('./newUserGuide')

test('manual guide restart clears recorded progress and enables replay mode', () => {
  const restarted = guide.resetNewUserGuideProgress()

  assert.equal(guide.countNewUserGuideCompleted(restarted), 0)
  assert.equal(guide.isNewUserGuideReplay(restarted), true)

  const restored = guide.readNewUserGuideProgress()
  assert.equal(guide.countNewUserGuideCompleted(restored), 0)
  assert.equal(guide.isNewUserGuideReplay(restored), true)
})

test('the first replay action records only the newly completed guide step', () => {
  guide.resetNewUserGuideProgress()
  const progress = guide.markNewUserGuideMilestone('SHOWCASE_VIEWED')

  assert.equal(guide.countNewUserGuideCompleted(progress), 1)
  assert.equal(guide.hasNewUserGuideMilestone(progress, 'SHOWCASE_VIEWED'), true)
  assert.equal(guide.hasNewUserGuideMilestone(progress, 'LOGIN_DONE'), false)
  assert.equal(guide.isNewUserGuideReplay(progress), true)
})

test('dismissing the guide only hides it temporarily and keeps its progress', () => {
  guide.resetNewUserGuideProgress()
  guide.markNewUserGuideMilestone('LOGIN_DONE')
  const dismissed = guide.dismissNewUserGuide(3)

  assert.equal(guide.hasNewUserGuideMilestone(dismissed, 'LOGIN_DONE'), true)
  assert.equal(guide.isNewUserGuideDismissed(dismissed, Date.now()), true)
  assert.equal(guide.isNewUserGuideDismissed(dismissed, Date.now() + 4 * 24 * 60 * 60 * 1000), false)
})

test('a joined member can complete their own path without creating another family', () => {
  guide.resetNewUserGuideProgress()
  guide.markNewUserGuideMilestone('TREE_GUIDE_OPENED')
  const progress = guide.markNewUserGuideMilestone('INVITE_GUIDE_OPENED')

  assert.equal(
    guide.areNewUserGuideMilestonesComplete(progress, ['TREE_GUIDE_OPENED', 'INVITE_GUIDE_OPENED']),
    true
  )
  assert.equal(guide.isNewUserGuideComplete(progress), false)
})
