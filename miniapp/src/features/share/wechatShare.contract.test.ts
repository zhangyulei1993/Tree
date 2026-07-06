import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import {
  buildArticleSharePayload,
  buildContentDetailSharePath,
  buildHomeSharePayload,
  buildInviteSharePath,
  buildInviteSharePayload,
  HOME_SHARE_PATH,
  MINI_PROGRAM_SHARE_TITLE
} from './wechatShare'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

function readPage(relativePath: string): string {
  return readFileSync(join(miniappSrc, relativePath), 'utf8')
}

test('article share path encodes article id and excludes invite token', () => {
  const path = buildContentDetailSharePath('article/12?x=1')
  assert.match(path, /^\/pages\/content\/detail\?id=/)
  assert.match(path, /article%2F12%3Fx%3D1/)
  assert.equal(path.includes('inviteToken'), false)
})

test('home share path is fixed and excludes invite token', () => {
  const payload = buildHomeSharePayload()
  assert.equal(payload.path, HOME_SHARE_PATH)
  assert.equal(payload.title, MINI_PROGRAM_SHARE_TITLE)
  assert.equal(payload.path.includes('inviteToken'), false)
})

test('invite share path keeps inviteToken and invite semantics', () => {
  const path = buildInviteSharePath('invite token/with+special')
  assert.match(path, /^\/pages\/invite\/detail\?inviteToken=/)
  assert.match(path, /invite%20token%2Fwith%2Bspecial/)
  const payload = buildInviteSharePayload({
    inviteToken: 'abc123',
    familyName: '张氏家族',
    targetMemberName: '张三'
  })
  assert.match(payload.title, /邀请你确认/)
  assert.match(payload.path, /inviteToken=abc123/)
  assert.equal(payload.path.includes('/pages/home/index'), false)
})

test('article share uses cover when https and falls back otherwise', () => {
  const withCover = buildArticleSharePayload({
    id: 9,
    title: '家谱入门',
    coverUrl: 'https://cdn.example.com/cover.jpg'
  })
  assert.equal(withCover.imageUrl, 'https://cdn.example.com/cover.jpg')
  assert.match(withCover.path, /id=9/)

  const withoutCover = buildArticleSharePayload({
    id: 9,
    title: '家谱入门',
    coverUrl: 'http://insecure.example.com/cover.jpg'
  })
  assert.match(withoutCover.imageUrl, /^\/static\/share\//)
})

test('content detail page wires wechat share button and handler', () => {
  const source = readPage('pages/content/detail.vue')
  assert.match(source, /onShareAppMessage/)
  assert.match(source, /buildArticleSharePayload/)
  assert.match(source, /#ifdef MP-WEIXIN/)
  assert.match(source, /open-type="share"/)
  assert.match(source, /分享给微信好友/)
})

test('home page wires mini program share entry and handler', () => {
  const source = readPage('pages/home/index.vue')
  assert.match(source, /onShareAppMessage/)
  assert.match(source, /buildHomeSharePayload/)
  assert.match(source, /分享小程序/)
  assert.match(source, /#ifdef MP-WEIXIN/)
  assert.match(source, /open-type="share"/)
})

test('invite sent page keeps inviteToken share semantics', () => {
  const source = readPage('pages/invite/sent.vue')
  assert.match(source, /buildInviteSharePayload/)
  assert.match(source, /shareResult\.value\.inviteToken/)
})
