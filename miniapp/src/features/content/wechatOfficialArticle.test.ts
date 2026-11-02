import assert from 'node:assert/strict'
import test from 'node:test'

import {
  isWechatOfficialArticleUrl,
  openWechatOfficialArticle,
  type WechatOfficialArticleRuntime
} from './wechatOfficialArticle'

test('accepts only HTTPS WeChat official article URLs', () => {
  assert.equal(isWechatOfficialArticleUrl('https://mp.weixin.qq.com/s/example'), true)
  assert.equal(isWechatOfficialArticleUrl('https://mp.weixin.qq.com/s?__biz=example'), true)
  assert.equal(isWechatOfficialArticleUrl('http://mp.weixin.qq.com/s/example'), false)
  assert.equal(isWechatOfficialArticleUrl('https://mp.weixin.qq.com.evil.example/s/example'), false)
  assert.equal(isWechatOfficialArticleUrl('https://mp.weixin.qq.com:444/s/example'), false)
})

test('opens an official article in WeChat mini program', async () => {
  let openedURL = ''
  const runtime: WechatOfficialArticleRuntime = {
    openOfficialAccountArticle(options) {
      openedURL = options.url
      options.success?.()
    }
  }
  await openWechatOfficialArticle('https://mp.weixin.qq.com/s/example', 'mp-weixin', runtime)
  assert.equal(openedURL, 'https://mp.weixin.qq.com/s/example')
})

test('rejects H5 and unsupported WeChat runtimes clearly', async () => {
  await assert.rejects(
    openWechatOfficialArticle('https://mp.weixin.qq.com/s/example', 'h5', {}),
    /微信小程序/
  )
  await assert.rejects(
    openWechatOfficialArticle('https://mp.weixin.qq.com/s/example', 'mp-weixin', {}),
    /升级微信/
  )
})
