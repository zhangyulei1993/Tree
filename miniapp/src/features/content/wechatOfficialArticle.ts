import { isMpWeixinPlatform } from '@/features/session/wechatLogin'

export interface WechatOfficialArticleRuntime {
  openOfficialAccountArticle?: (options: {
    url: string
    success?: () => void
    fail?: (error: unknown) => void
  }) => void
}

function readWechatRuntime(): WechatOfficialArticleRuntime | null {
  try {
    const runtime = (globalThis as typeof globalThis & { wx?: WechatOfficialArticleRuntime }).wx
    return runtime || null
  } catch {
    return null
  }
}

export function isWechatOfficialArticleUrl(value: string): boolean {
  return /^https:\/\/mp\.weixin\.qq\.com\/s(?:\/[^?\s#]+|\?[^#\s]+)$/i.test(value.trim())
}

export async function openWechatOfficialArticle(
  url: string,
  platform?: string,
  runtime = readWechatRuntime()
): Promise<void> {
  if (!isWechatOfficialArticleUrl(url)) {
    throw new Error('公众号文章链接无效')
  }
  if (!isMpWeixinPlatform(platform)) {
    throw new Error('请在微信小程序中阅读公众号文章')
  }
  if (!runtime?.openOfficialAccountArticle) {
    throw new Error('当前微信版本暂不支持打开公众号文章，请升级微信后重试')
  }
  await new Promise<void>((resolve, reject) => {
    runtime.openOfficialAccountArticle?.({
      url: url.trim(),
      success: resolve,
      fail: () => reject(new Error('公众号文章暂时无法打开'))
    })
  })
}
