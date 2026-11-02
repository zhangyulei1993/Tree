let launchSplashShown = false

export function shouldShowLaunchSplash() {
  return !launchSplashShown
}

export function markLaunchSplashShown() {
  launchSplashShown = true
}
