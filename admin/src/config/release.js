export const APP_VERSION = __APP_VERSION__
export const LATEST_RELEASE_URL = 'https://api.github.com/repos/Rodert/agi-platform/releases/latest'

export const compareVersions = (left, right) => {
  const leftParts = left.replace(/^v/, '').split('.').map(Number)
  const rightParts = right.replace(/^v/, '').split('.').map(Number)
  const length = Math.max(leftParts.length, rightParts.length)

  for (let index = 0; index < length; index += 1) {
    const difference = (leftParts[index] || 0) - (rightParts[index] || 0)
    if (difference !== 0) return difference
  }
  return 0
}
