const templateIDs = {
  audit: 'J01V0pQPCGmPh3Usc2QhERgEssGGn3OzlR9tMdAxstM',
  payment: '38Z7OnugGrWMHpkJJzzafizpP-wLKmm5QcSoAeZEenI',
  refund: 'Z3iuuKsJ-uVjvoEOO40rxE2AsWD5K4NwQ_ZgC54QELI',
  member: 'ascQYwPQerydgmHZSGoYoJSGYHbWu-ToWnCZ4XD5w98',
  memberChange: 'kYC9F28vTqXCiH0mahyw5kXoU2mFrrr13rcxzoQ2_uc',
} as const

export type SubscribeMessageKind = keyof typeof templateIDs

export type SubscribeMessageResult =
  | { status: 'accepted' }
  | { status: 'rejected' }
  | { status: 'unsupported' }

export type SubscribeMessageResults = Partial<Record<SubscribeMessageKind, SubscribeMessageResult>>

export function requestSubscribeMessages(kinds: SubscribeMessageKind[]): Promise<SubscribeMessageResults> {
  const uniqueKinds = [...new Set(kinds)]
  if (uniqueKinds.length === 0) return Promise.resolve({})
  if (typeof uni === 'undefined') return Promise.resolve(resultsFor(uniqueKinds, 'unsupported'))
  const uniRuntime = uni as unknown as {
    getSystemInfoSync?: () => { uniPlatform?: string }
  }
  if (uniRuntime.getSystemInfoSync?.().uniPlatform !== 'mp-weixin') return Promise.resolve(resultsFor(uniqueKinds, 'unsupported'))
  const uniSubscribe = uniRuntime as typeof uniRuntime & {
    requestSubscribeMessage?: (options: {
      tmplIds: string[]
      success?: (result: Record<string, unknown>) => void
      fail?: () => void
    }) => void
  }
  const request = uniSubscribe.requestSubscribeMessage ?? (globalThis as typeof globalThis & {
    wx?: {
      requestSubscribeMessage?: (options: {
        tmplIds: string[]
        success?: (result: Record<string, unknown>) => void
        fail?: () => void
      }) => void
    }
  }).wx?.requestSubscribeMessage
  if (!request) return Promise.resolve(resultsFor(uniqueKinds, 'unsupported'))

  const templateIDsToRequest = uniqueKinds.map((kind) => templateIDs[kind])
  return new Promise((resolve) => {
    request({
      tmplIds: templateIDsToRequest,
      success: (result) => {
        const statuses: SubscribeMessageResults = {}
        uniqueKinds.forEach((kind) => {
          statuses[kind] = { status: result[templateIDs[kind]] === 'accept' ? 'accepted' : 'rejected' }
        })
        resolve(statuses)
      },
      fail: () => resolve(resultsFor(uniqueKinds, 'rejected')),
    })
  })
}

export async function requestSubscribeMessage(kind: SubscribeMessageKind): Promise<SubscribeMessageResult> {
  return (await requestSubscribeMessages([kind]))[kind] ?? { status: 'rejected' }
}

function resultsFor(kinds: SubscribeMessageKind[], status: SubscribeMessageResult['status']): SubscribeMessageResults {
  return Object.fromEntries(kinds.map((kind) => [kind, { status }])) as SubscribeMessageResults
}
