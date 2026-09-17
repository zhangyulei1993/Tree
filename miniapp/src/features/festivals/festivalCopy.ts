import type { FestivalItem } from './festivalData'

export type FestivalCopyScene = 'personal' | 'group' | 'moments'

export const festivalCopySceneOptions: Array<{ value: FestivalCopyScene; label: string }> = [
  { value: 'personal', label: '个人祝福' },
  { value: 'group', label: '群聊祝福' },
  { value: 'moments', label: '朋友圈文案' }
]

const festivalDetails: Record<string, string> = {
  春节: '新春伊始，愿旧岁所有遗憾都被安放，新的一年有新的收获。',
  元宵节: '灯火可亲，愿家人团圆相伴，日子越过越亮堂。',
  清明节: '清明慎终追远，也愿我们珍惜眼前人，记得常回家看看。',
  端午节: '端午安康，愿家人平安健康，生活有滋有味。',
  七夕节: '愿有情人相知相伴，也愿每一份真心都被温柔以待。',
  中秋节: '月圆人团圆，愿无论相隔多远，家人的牵挂始终都在。',
  重阳节: '愿长辈身体康健，愿一家人常伴左右，把日子过得从容安稳。',
  冬至: '冬至阳生，愿家人平安过冬，常聚一桌热饭，常说几句家常。'
}

const sceneTemplates: Record<FestivalCopyScene, Array<(festival: FestivalItem, detail: string) => string>> = {
  personal: [
    (festival, detail) => `${festival.name}快乐。${detail}愿你平安健康，生活顺心如意。`,
    (festival, detail) => `${festival.name}到了，愿你所念皆有回应，所行都能顺利。${detail}`,
    (festival, detail) => `借${festival.name}的好日子，送上一份祝福：愿你平安喜乐，万事顺意。`,
    (festival, detail) => `${festival.name}安康，愿忙碌有收获，闲暇有欢喜，日子越过越踏实。`,
    (festival, detail) => `又逢${festival.name}，愿你心中有暖，身边有人，脚下有路。${detail}`,
    (festival, detail) => `${festival.name}到了，愿生活有惊喜，家人常牵挂，往后的每一天都顺遂。`,
    (festival, detail) => `祝${festival.name}快乐，愿你平安常在、好事常来，生活从容自在。`,
    (festival, detail) => `${festival.name}，愿一切美好如期而至，愿每份努力都值得。${detail}`,
    (festival, detail) => `在这个${festival.name}，愿你有好心情，也有值得期待的明天。`,
    (festival, detail) => `${festival.name}快乐，愿日子清简有味，家人朋友平安，心中常有欢喜。`
  ],
  group: [
    (festival, detail) => `${festival.name}快乐。${detail}愿大家平安顺遂，常联系、常团聚。`,
    (festival, detail) => `${festival.name}到了，祝大家身体健康、工作顺利、生活如意。${detail}`,
    (festival) => `值此${festival.name}，愿群里的每一位都平安喜乐，事事顺心。`,
    (festival, detail) => `${festival.name}安康，愿大家有空常联系，有机会常相聚。${detail}`,
    (festival) => `又是一年${festival.name}，愿我们各自安好，也一直保持联系。`,
    (festival, detail) => `${festival.name}到了，愿大家心中有暖、手里有事、家中有笑声。${detail}`,
    (festival) => `祝大家${festival.name}快乐，愿新的一程平安相伴，好事连连。`,
    (festival, detail) => `借${festival.name}送上问候，愿大家平安健康，所愿皆成。${detail}`,
    (festival) => `${festival.name}，愿大家忙有所获、闲有所乐，家人朋友都安好。`,
    (festival, detail) => `${festival.name}快乐，愿这份节日的喜气传给每一个人。${detail}`
  ],
  moments: [
    (festival, detail) => `${festival.name}到了。${detail}愿日子有暖、有盼，家人朋友都平安顺遂。`,
    (festival) => `${festival.name}，把思念放进祝福里，把平安留给身边的人。`,
    (festival) => `又逢${festival.name}，愿所遇皆温柔，所行皆顺意。`,
    (festival, detail) => `${festival.name}安康。愿生活不急不慢，家人朋友常伴左右。${detail}`,
    (festival) => `日子有节气，生活有欢喜。祝${festival.name}快乐。`,
    (festival, detail) => `${festival.name}到了，愿每一个普通的日子都值得珍惜。${detail}`,
    (festival) => `愿这个${festival.name}，有人惦记，有所期待，有好消息发生。`,
    (festival, detail) => `${festival.name}，愿烟火日常里有小确幸，也有稳稳的平安。${detail}`,
    (festival) => `把祝福写在今天：愿${festival.name}平安喜乐，万事顺意。`,
    (festival, detail) => `${festival.name}快乐，愿远近亲友都安好，愿团聚总有时。${detail}`
  ]
}

export const FESTIVAL_COPY_VARIANT_COUNT = 10

export function generateFestivalCopy(
  festival: FestivalItem,
  scene: FestivalCopyScene,
  variantIndex = 0
): string {
  const detail = festivalDetails[festival.name] || `愿这个${festival.name}为大家带来平安与好心情。`
  const templates = sceneTemplates[scene]
  const normalizedIndex = ((variantIndex % templates.length) + templates.length) % templates.length
  return templates[normalizedIndex](festival, detail)
}
