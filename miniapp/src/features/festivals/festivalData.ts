export interface FestivalItem {
  name: string
  date: string
  lunarDate: string
  description: string
  introduction: string
  customs: string
  holidayRange?: string
  workdays?: string
  category: '传统节日' | '法定节假日'
}

export interface SolarTermItem {
  name: string
  date: string
  description: string
}

export const festivalItems2026: FestivalItem[] = [
  {
    name: '腊八节',
    date: '2026-01-26',
    lunarDate: '农历腊月初八',
    description: '喝腊八粥，迎接年节。',
    introduction: '腊八节在农历十二月初八，是传统年节的序幕，也寄托着迎祥纳福的愿望。',
    customs: '常见习俗有喝腊八粥、泡腊八蒜，也有人从这一天开始准备年货。',
    category: '传统节日'
  },
  {
    name: '除夕',
    date: '2026-02-16',
    lunarDate: '农历腊月廿九',
    description: '辞旧迎新，家人团聚。',
    introduction: '除夕是农历一年最后一天的晚上，是辞旧迎新、家人团聚的重要时刻。',
    customs: '常见习俗有年夜饭、贴春联、守岁和给晚辈压岁钱。',
    holidayRange: '2月15日—2月23日放假调休',
    workdays: '2月14日、2月28日上班',
    category: '法定节假日'
  },
  {
    name: '春节',
    date: '2026-02-17',
    lunarDate: '农历正月初一',
    description: '一年之始，也是家人团聚、走亲访友的重要节日。',
    introduction: '春节是农历新年的开始，也是中国人最重要的传统节日之一。',
    customs: '人们会拜年、走亲访友、贴春联、看灯会，用热闹的方式迎接新年。',
    holidayRange: '2月15日—2月23日放假调休',
    workdays: '2月14日、2月28日上班',
    category: '法定节假日'
  },
  {
    name: '元宵节',
    date: '2026-03-03',
    lunarDate: '农历正月十五',
    description: '赏灯、猜灯谜，吃元宵或汤圆。',
    introduction: '元宵节是春节年俗的重要收尾，象征团圆和对美好生活的期盼。',
    customs: '常见习俗有赏花灯、猜灯谜、吃元宵或汤圆。',
    category: '传统节日'
  },
  {
    name: '清明节',
    date: '2026-04-05',
    lunarDate: '二十四节气·清明',
    description: '踏青、祭祖，寄托对先人的思念。',
    introduction: '清明节兼具节气与节日意义，是慎终追远、亲近自然的传统日子。',
    customs: '常见活动有祭扫、追思先人、踏青和插柳，表达思念与珍惜。',
    holidayRange: '4月4日—4月6日放假',
    category: '法定节假日'
  },
  {
    name: '劳动节',
    date: '2026-05-01',
    lunarDate: '公历节日',
    description: '劳动者的节日。',
    introduction: '劳动节是向劳动者致意的节日，也提醒我们尊重每一份认真而踏实的付出。',
    customs: '人们常利用假期休息、出游或陪伴家人，也会向身边辛勤工作的人表达感谢。',
    holidayRange: '5月1日—5月5日放假调休',
    workdays: '5月9日上班',
    category: '法定节假日'
  },
  {
    name: '端午节',
    date: '2026-06-19',
    lunarDate: '农历五月初五',
    description: '包粽子、赛龙舟，纪念传统文化。',
    introduction: '端午节是中国传统节日，承载着祈求安康、珍惜家国情怀的文化记忆。',
    customs: '常见习俗有包粽子、赛龙舟、佩香囊等，各地也有不同的节日做法。',
    holidayRange: '6月19日—6月21日放假',
    category: '法定节假日'
  },
  {
    name: '七夕节',
    date: '2026-08-19',
    lunarDate: '农历七月初七',
    description: '流传千年的爱情与乞巧节日。',
    introduction: '七夕节源于古老的星宿传说，也保留着人们对美好感情和心灵手巧的向往。',
    customs: '传统活动有乞巧、观星和寄托心愿，现代也常被当作表达心意的日子。',
    category: '传统节日'
  },
  {
    name: '中元节',
    date: '2026-08-27',
    lunarDate: '农历七月十五',
    description: '慎终追远，寄托对先人的追思。',
    introduction: '中元节在民俗中与祭祖、追思和祈福有关，提醒人们不忘来处、珍惜亲情。',
    customs: '各地习俗不同，常见做法有祭祖、追思先人和祈愿平安。',
    category: '传统节日'
  },
  {
    name: '中秋节',
    date: '2026-09-25',
    lunarDate: '农历八月十五',
    description: '赏月、团圆，寄托对家人的思念。',
    introduction: '中秋节以月圆寄托团圆，是家人表达思念、分享收获的重要节日。',
    customs: '常见习俗有赏月、吃月饼、饮桂花酒，也有各地独特的团圆习俗。',
    holidayRange: '9月25日—9月27日放假',
    category: '法定节假日'
  },
  {
    name: '国庆节',
    date: '2026-10-01',
    lunarDate: '公历节日',
    description: '中华人民共和国成立纪念日。',
    introduction: '国庆节是庆祝中华人民共和国成立的节日，也是全国共同欢庆的日子。',
    customs: '人们会观看庆祝活动、出行游览或与家人相聚，共同表达祝福。',
    holidayRange: '10月1日—10月7日放假调休',
    workdays: '9月20日、10月10日上班',
    category: '法定节假日'
  },
  {
    name: '重阳节',
    date: '2026-10-18',
    lunarDate: '农历九月初九',
    description: '登高、赏菊，也承载敬老与思亲的传统。',
    introduction: '重阳节有登高、敬老和思亲的文化内涵，适合向长辈表达关心与祝福。',
    customs: '常见习俗有登高、赏菊、佩茱萸、饮菊花酒和陪伴长辈。',
    category: '传统节日'
  },
  {
    name: '冬至',
    date: '2026-12-22',
    lunarDate: '二十四节气·冬至',
    description: '冬至阳生，北方吃饺子，南方吃汤圆。',
    introduction: '冬至是二十四节气之一，也是许多地方重视的传统节令，寓意寒尽阳生。',
    customs: '北方常吃饺子，南方常吃汤圆，各地也有围炉、祭祖和进补等习俗。',
    category: '传统节日'
  }
]

export const solarTermItems2026: SolarTermItem[] = [
  { name: '小寒', date: '2026-01-05', description: '天气渐寒，提醒人们顺应时节，注意保暖。' },
  { name: '大寒', date: '2026-01-20', description: '一年中寒冷的时节，年味也在此时渐渐浓起来。' },
  { name: '立春', date: '2026-02-04', description: '春季开始，万物更新，迎来新一年的生机。' },
  { name: '雨水', date: '2026-02-18', description: '降水开始增多，草木渐有春意。' },
  { name: '惊蛰', date: '2026-03-05', description: '春雷始鸣，蛰虫惊醒，大地慢慢恢复活力。' },
  { name: '春分', date: '2026-03-20', description: '昼夜平分，春意正盛，适合亲近自然。' },
  { name: '清明', date: '2026-04-05', description: '天气清和明朗，适合踏青，也寄托慎终追远的情怀。' },
  { name: '谷雨', date: '2026-04-20', description: '雨生百谷，春季将尽，农事渐忙。' },
  { name: '立夏', date: '2026-05-05', description: '夏季开始，万物进入生长旺盛的阶段。' },
  { name: '小满', date: '2026-05-21', description: '麦粒渐满但未成熟，提醒人们知足惜福。' },
  { name: '芒种', date: '2026-06-05', description: '有芒作物适宜播种，农事进入繁忙时节。' },
  { name: '夏至', date: '2026-06-21', description: '一年中白昼最长的一天，盛夏由此展开。' },
  { name: '小暑', date: '2026-07-07', description: '天气开始炎热，但还未到最热的时候。' },
  { name: '大暑', date: '2026-07-23', description: '一年中最炎热的时期，注意防暑与作息。' },
  { name: '立秋', date: '2026-08-07', description: '秋季开始，暑气未退而秋意渐生。' },
  { name: '处暑', date: '2026-08-23', description: '暑气渐止，早晚开始有了秋天的凉意。' },
  { name: '白露', date: '2026-09-07', description: '天气转凉，清晨草木上常见晶莹露珠。' },
  { name: '秋分', date: '2026-09-23', description: '昼夜再次平分，秋色渐深，收获将至。' },
  { name: '寒露', date: '2026-10-08', description: '露水更冷，天气由凉转寒。' },
  { name: '霜降', date: '2026-10-23', description: '天气渐寒，开始出现霜冻，秋季接近尾声。' },
  { name: '立冬', date: '2026-11-07', description: '冬季开始，万物收藏，生活节奏渐趋安静。' },
  { name: '小雪', date: '2026-11-22', description: '气温继续下降，部分地区开始降雪。' },
  { name: '大雪', date: '2026-12-07', description: '降雪可能增多，寒意渐浓，宜围炉添衣。' },
  { name: '冬至', date: '2026-12-22', description: '白昼最短，阳气开始回升，也是家人团聚的节令。' }
]

export function getFestivalItems(year: number): FestivalItem[] {
  return year === 2026 ? festivalItems2026 : []
}

export function getSolarTermItems(year: number): SolarTermItem[] {
  return year === 2026 ? solarTermItems2026 : []
}

export function dateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function festivalDateLabel(date: string): string {
  return `${Number(date.slice(5, 7))}月${Number(date.slice(8, 10))}日`
}
