export type NewUserGuideMilestone =
  | 'SHOWCASE_VIEWED'
  | 'LOGIN_DONE'
  | 'PROFILE_DONE'
  | 'FAMILY_CREATED'
  | 'MEMBER_GUIDE_OPENED'
  | 'RELATION_GUIDE_OPENED'
  | 'INVITE_GUIDE_OPENED'
  | 'TREE_GUIDE_OPENED'

export interface NewUserGuideProgress {
  version: 1
  completed: Partial<Record<NewUserGuideMilestone, string>>
  lastMilestone?: NewUserGuideMilestone
  updatedAt?: string
  /** A manual restart replays steps instead of deriving them from existing account data. */
  replaying?: boolean
  /** Hide the card temporarily without losing the route or completed steps. */
  dismissedUntil?: string
}

const NEW_USER_GUIDE_STORAGE_KEY = 'tree_miniapp_new_user_guide_v1'
export const NEW_USER_GUIDE_REQUIRED_MILESTONES: NewUserGuideMilestone[] = [
  'SHOWCASE_VIEWED',
  'LOGIN_DONE',
  'FAMILY_CREATED',
  'MEMBER_GUIDE_OPENED',
  'RELATION_GUIDE_OPENED',
  'INVITE_GUIDE_OPENED'
]

function emptyProgress(): NewUserGuideProgress {
  return {
    version: 1,
    completed: {}
  }
}

function normalizeProgress(value: unknown): NewUserGuideProgress {
  if (!value || typeof value !== 'object') return emptyProgress()
  const source = value as Partial<NewUserGuideProgress>
  if (source.version !== 1 || !source.completed || typeof source.completed !== 'object') {
    return emptyProgress()
  }
  return {
    version: 1,
    completed: { ...source.completed },
    lastMilestone: source.lastMilestone,
    updatedAt: source.updatedAt,
    replaying: source.replaying === true,
    dismissedUntil: source.dismissedUntil
  }
}

export function readNewUserGuideProgress(): NewUserGuideProgress {
  try {
    const stored = uni.getStorageSync(NEW_USER_GUIDE_STORAGE_KEY)
    const parsed = typeof stored === 'string' ? JSON.parse(stored) : stored
    return normalizeProgress(parsed)
  } catch {
    uni.removeStorageSync(NEW_USER_GUIDE_STORAGE_KEY)
    return emptyProgress()
  }
}

export function hasNewUserGuideMilestone(
  progress: NewUserGuideProgress,
  milestone: NewUserGuideMilestone
) {
  return Boolean(progress.completed[milestone])
}

export function isNewUserGuideComplete(progress: NewUserGuideProgress) {
  return areNewUserGuideMilestonesComplete(progress, NEW_USER_GUIDE_REQUIRED_MILESTONES)
}

/**
 * A person who joined a family does not need to create another one. Keep the
 * completion check scoped to the path currently shown to that person.
 */
export function areNewUserGuideMilestonesComplete(
  progress: NewUserGuideProgress,
  milestones: NewUserGuideMilestone[]
) {
  return milestones.every((milestone) => hasNewUserGuideMilestone(progress, milestone))
}

export function isNewUserGuideReplay(progress: NewUserGuideProgress) {
  return progress.replaying === true
}

export function isNewUserGuideDismissed(progress: NewUserGuideProgress, now = Date.now()) {
  if (!progress.dismissedUntil) return false
  const until = new Date(progress.dismissedUntil).getTime()
  return Number.isFinite(until) && until > now
}

export function countNewUserGuideCompleted(
  progress: NewUserGuideProgress,
  milestones = NEW_USER_GUIDE_REQUIRED_MILESTONES
) {
  return milestones.filter((milestone) =>
    hasNewUserGuideMilestone(progress, milestone)
  ).length
}

export function markNewUserGuideMilestone(milestone: NewUserGuideMilestone) {
  const progress = readNewUserGuideProgress()
  const now = new Date().toISOString()
  progress.completed[milestone] = progress.completed[milestone] || now
  progress.lastMilestone = milestone
  progress.updatedAt = now
  progress.dismissedUntil = undefined
  uni.setStorageSync(NEW_USER_GUIDE_STORAGE_KEY, JSON.stringify(progress))
  return progress
}

export function dismissNewUserGuide(days = 3) {
  const progress = readNewUserGuideProgress()
  const now = new Date()
  progress.dismissedUntil = new Date(now.getTime() + Math.max(1, days) * 24 * 60 * 60 * 1000).toISOString()
  progress.updatedAt = now.toISOString()
  uni.setStorageSync(NEW_USER_GUIDE_STORAGE_KEY, JSON.stringify(progress))
  return progress
}

export function resumeNewUserGuide() {
  const progress = readNewUserGuideProgress()
  progress.dismissedUntil = undefined
  progress.updatedAt = new Date().toISOString()
  uni.setStorageSync(NEW_USER_GUIDE_STORAGE_KEY, JSON.stringify(progress))
  return progress
}

export function markNewUserGuideMilestones(milestones: NewUserGuideMilestone[]) {
  let progress = readNewUserGuideProgress()
  for (const milestone of milestones) {
    progress = markNewUserGuideMilestone(milestone)
  }
  return progress
}

export function resetNewUserGuideProgress() {
  const progress: NewUserGuideProgress = {
    ...emptyProgress(),
    replaying: true,
    updatedAt: new Date().toISOString()
  }
  uni.setStorageSync(NEW_USER_GUIDE_STORAGE_KEY, JSON.stringify(progress))
  return progress
}
