-- 平台独立展示内容初始化：只操作 showcase_* 表。
-- 人物、父子、配偶关系按公开史料整理，仅用于历史示例展示，不代表完整世系考据。
SET NAMES utf8mb4;
START TRANSACTION;

DELETE FROM showcase_family_relationships;
DELETE FROM showcase_family_members;
DELETE FROM showcase_family_configs;
UPDATE showcase_display_settings SET pinned_showcase_family_id = NULL, pinned_until = NULL;

-- JSON 条目格式：[姓名, 父/夫节点序号（null 为根）, 节点类型, 性别]。
-- 同一展示册中先按数组顺序建节点，再按父/夫序号创建 PARENT_CHILD 或 SPOUSE 关系。

INSERT INTO showcase_family_configs (display_name, region_text, summary, status, sort_order, published_at)
VALUES ('汉高祖刘邦五代家系', '沛县 · 西汉', '历史示例 · 根据公开史料整理。以刘邦为起点，展示五代内的主要支系与配偶关系。', 'PUBLISHED', 10, NOW());
SET @han = LAST_INSERT_ID();
SET @han_seed = '[
["刘邦",null,"LINEAGE_MEMBER","MALE"],["曹氏",1,"SPOUSE","FEMALE"],["吕雉",1,"SPOUSE","FEMALE"],["戚夫人",1,"SPOUSE","FEMALE"],["薄姬",1,"SPOUSE","FEMALE"],["赵姬",1,"SPOUSE","FEMALE"],
["刘肥",1,"LINEAGE_MEMBER","MALE"],["刘盈",1,"LINEAGE_MEMBER","MALE"],["刘如意",1,"LINEAGE_MEMBER","MALE"],["刘恒",1,"LINEAGE_MEMBER","MALE"],["刘恢",1,"LINEAGE_MEMBER","MALE"],["刘友",1,"LINEAGE_MEMBER","MALE"],["刘长",1,"LINEAGE_MEMBER","MALE"],["刘建",1,"LINEAGE_MEMBER","MALE"],["鲁元公主",1,"LINEAGE_MEMBER","FEMALE"],
["刘强",8,"LINEAGE_MEMBER","MALE"],["刘不疑",8,"LINEAGE_MEMBER","MALE"],["刘弘",8,"LINEAGE_MEMBER","MALE"],["刘朝",8,"LINEAGE_MEMBER","MALE"],["刘武（惠帝子）",8,"LINEAGE_MEMBER","MALE"],["刘太",8,"LINEAGE_MEMBER","MALE"],
["窦皇后",10,"SPOUSE","FEMALE"],["刘启",10,"LINEAGE_MEMBER","MALE"],["刘武（文帝子）",10,"LINEAGE_MEMBER","MALE"],["刘参",10,"LINEAGE_MEMBER","MALE"],["刘揖",10,"LINEAGE_MEMBER","MALE"],["刘嫖",10,"LINEAGE_MEMBER","FEMALE"],
["刘荣",23,"LINEAGE_MEMBER","MALE"],["刘德",23,"LINEAGE_MEMBER","MALE"],["刘阏",23,"LINEAGE_MEMBER","MALE"],["刘余",23,"LINEAGE_MEMBER","MALE"],["刘非",23,"LINEAGE_MEMBER","MALE"],["刘发",23,"LINEAGE_MEMBER","MALE"],["刘端",23,"LINEAGE_MEMBER","MALE"],["刘彭祖",23,"LINEAGE_MEMBER","MALE"],["刘胜",23,"LINEAGE_MEMBER","MALE"],["刘彻",23,"LINEAGE_MEMBER","MALE"],["刘越",23,"LINEAGE_MEMBER","MALE"],["刘寄",23,"LINEAGE_MEMBER","MALE"],["刘乘",23,"LINEAGE_MEMBER","MALE"],["刘舜",23,"LINEAGE_MEMBER","MALE"],["阳信公主",23,"LINEAGE_MEMBER","FEMALE"],["南宫公主",23,"LINEAGE_MEMBER","FEMALE"],["隆虑公主",23,"LINEAGE_MEMBER","FEMALE"],
["刘据",37,"LINEAGE_MEMBER","MALE"],["刘闳",37,"LINEAGE_MEMBER","MALE"],["刘旦",37,"LINEAGE_MEMBER","MALE"],["刘胥",37,"LINEAGE_MEMBER","MALE"],["刘髆",37,"LINEAGE_MEMBER","MALE"],["刘弗陵",37,"LINEAGE_MEMBER","MALE"]]';

INSERT INTO showcase_family_configs (display_name, region_text, summary, status, sort_order, published_at)
VALUES ('唐高祖李渊五代家系', '长安 · 唐朝', '历史示例 · 根据公开史料整理。以李渊为起点，展示五代内的主要支系与配偶关系。', 'PUBLISHED', 20, NOW());
SET @tang = LAST_INSERT_ID();
SET @tang_seed = '[
["李渊",null,"LINEAGE_MEMBER","MALE"],["窦皇后",1,"SPOUSE","FEMALE"],["万贵妃",1,"SPOUSE","FEMALE"],["尹德妃",1,"SPOUSE","FEMALE"],
["李建成",1,"LINEAGE_MEMBER","MALE"],["李世民",1,"LINEAGE_MEMBER","MALE"],["李玄霸",1,"LINEAGE_MEMBER","MALE"],["李元吉",1,"LINEAGE_MEMBER","MALE"],["李智云",1,"LINEAGE_MEMBER","MALE"],["李元景",1,"LINEAGE_MEMBER","MALE"],["李元昌",1,"LINEAGE_MEMBER","MALE"],["李元亨",1,"LINEAGE_MEMBER","MALE"],["李元方",1,"LINEAGE_MEMBER","MALE"],["李元礼",1,"LINEAGE_MEMBER","MALE"],["李元嘉",1,"LINEAGE_MEMBER","MALE"],["李元则",1,"LINEAGE_MEMBER","MALE"],["李元懿",1,"LINEAGE_MEMBER","MALE"],["李元轨",1,"LINEAGE_MEMBER","MALE"],["李凤",1,"LINEAGE_MEMBER","MALE"],["李元庆",1,"LINEAGE_MEMBER","MALE"],["李元裕",1,"LINEAGE_MEMBER","MALE"],
["长孙皇后",6,"SPOUSE","FEMALE"],["韦贵妃",6,"SPOUSE","FEMALE"],["杨妃",6,"SPOUSE","FEMALE"],["阴妃",6,"SPOUSE","FEMALE"],["燕德妃",6,"SPOUSE","FEMALE"],
["李承乾",6,"LINEAGE_MEMBER","MALE"],["李宽",6,"LINEAGE_MEMBER","MALE"],["李恪",6,"LINEAGE_MEMBER","MALE"],["李泰",6,"LINEAGE_MEMBER","MALE"],["李祐",6,"LINEAGE_MEMBER","MALE"],["李愔",6,"LINEAGE_MEMBER","MALE"],["李恽",6,"LINEAGE_MEMBER","MALE"],["李贞",6,"LINEAGE_MEMBER","MALE"],["李治",6,"LINEAGE_MEMBER","MALE"],["李慎",6,"LINEAGE_MEMBER","MALE"],["李嚣",6,"LINEAGE_MEMBER","MALE"],["李简",6,"LINEAGE_MEMBER","MALE"],["李福",6,"LINEAGE_MEMBER","MALE"],["李明",6,"LINEAGE_MEMBER","MALE"],["李灵夔",6,"LINEAGE_MEMBER","MALE"],
["李忠",35,"LINEAGE_MEMBER","MALE"],["李孝",35,"LINEAGE_MEMBER","MALE"],["李上金",35,"LINEAGE_MEMBER","MALE"],["李素节",35,"LINEAGE_MEMBER","MALE"],["李弘",35,"LINEAGE_MEMBER","MALE"],["李贤",35,"LINEAGE_MEMBER","MALE"],["李显",35,"LINEAGE_MEMBER","MALE"],["李旦",35,"LINEAGE_MEMBER","MALE"],["太平公主",35,"LINEAGE_MEMBER","FEMALE"]]';

INSERT INTO showcase_family_configs (display_name, region_text, summary, status, sort_order, published_at)
VALUES ('宋太祖赵匡胤五代家系', '开封 · 北宋', '历史示例 · 根据公开史料整理。以赵匡胤为起点，展示五代内的主要支系与配偶关系。', 'PUBLISHED', 30, NOW());
SET @song = LAST_INSERT_ID();
SET @song_seed = '[
["赵匡胤",null,"LINEAGE_MEMBER","MALE"],["贺氏",1,"SPOUSE","FEMALE"],["王氏",1,"SPOUSE","FEMALE"],["宋氏",1,"SPOUSE","FEMALE"],
["赵德秀",1,"LINEAGE_MEMBER","MALE"],["赵德昭",1,"LINEAGE_MEMBER","MALE"],["赵德林",1,"LINEAGE_MEMBER","MALE"],["赵德芳",1,"LINEAGE_MEMBER","MALE"],["魏国长公主",1,"LINEAGE_MEMBER","FEMALE"],["鲁国长公主",1,"LINEAGE_MEMBER","FEMALE"],["陈国长公主",1,"LINEAGE_MEMBER","FEMALE"],["申国长公主",1,"LINEAGE_MEMBER","FEMALE"],["成国长公主",1,"LINEAGE_MEMBER","FEMALE"],["永国长公主",1,"LINEAGE_MEMBER","FEMALE"],
["赵惟正",6,"LINEAGE_MEMBER","MALE"],["赵惟吉",6,"LINEAGE_MEMBER","MALE"],["赵惟固",6,"LINEAGE_MEMBER","MALE"],["赵惟忠",6,"LINEAGE_MEMBER","MALE"],["赵惟和",6,"LINEAGE_MEMBER","MALE"],
["焦氏",8,"SPOUSE","FEMALE"],["符氏",8,"SPOUSE","FEMALE"],["王氏（赵德芳配偶）",8,"SPOUSE","FEMALE"],["李氏",8,"SPOUSE","FEMALE"],
["赵惟叙",8,"LINEAGE_MEMBER","MALE"],["赵惟宪",8,"LINEAGE_MEMBER","MALE"],["赵惟能",8,"LINEAGE_MEMBER","MALE"],["永寿县主",8,"LINEAGE_MEMBER","FEMALE"],
["赵从郁",25,"LINEAGE_MEMBER","MALE"],["赵从演",25,"LINEAGE_MEMBER","MALE"],["赵从戒",25,"LINEAGE_MEMBER","MALE"],["赵从式",25,"LINEAGE_MEMBER","MALE"],["赵从戎",25,"LINEAGE_MEMBER","MALE"],["赵从湜",25,"LINEAGE_MEMBER","MALE"],["赵从贲",25,"LINEAGE_MEMBER","MALE"],
["赵世将",28,"LINEAGE_MEMBER","MALE"],["赵世经",29,"LINEAGE_MEMBER","MALE"],["赵世居",30,"LINEAGE_MEMBER","MALE"],["赵世尧",24,"LINEAGE_MEMBER","MALE"],
["赵令譮",35,"LINEAGE_MEMBER","MALE"],["赵令望",35,"LINEAGE_MEMBER","MALE"],["赵令旼",35,"LINEAGE_MEMBER","MALE"],["赵令宪",36,"LINEAGE_MEMBER","MALE"],["赵令友",36,"LINEAGE_MEMBER","MALE"],["赵令誉",37,"LINEAGE_MEMBER","MALE"],["赵令周",37,"LINEAGE_MEMBER","MALE"],["赵令振",38,"LINEAGE_MEMBER","MALE"],["赵令度",38,"LINEAGE_MEMBER","MALE"],["赵令钦",31,"LINEAGE_MEMBER","MALE"],["赵令舒",32,"LINEAGE_MEMBER","MALE"],["赵令原",33,"LINEAGE_MEMBER","MALE"],["赵令达",34,"LINEAGE_MEMBER","MALE"],["赵令衡",26,"LINEAGE_MEMBER","MALE"],["赵令荣",27,"LINEAGE_MEMBER","MALE"]]';

CREATE TEMPORARY TABLE seed_showcase_members (
  showcase_family_id BIGINT NOT NULL,
  node_no INT NOT NULL,
  parent_no INT NULL,
  member_type VARCHAR(40) NOT NULL,
  PRIMARY KEY (showcase_family_id, node_no)
);

INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, sort_order)
SELECT @han, row_data.node_type, row_data.display_name, row_data.gender, row_data.node_no * 10
FROM JSON_TABLE(@han_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, display_name VARCHAR(80) PATH '$[0]', node_type VARCHAR(40) PATH '$[2]', gender VARCHAR(20) PATH '$[3]')) row_data;
INSERT INTO seed_showcase_members
SELECT @han, row_data.node_no, row_data.parent_no, row_data.node_type
FROM JSON_TABLE(@han_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, parent_no INT PATH '$[1]' NULL ON EMPTY, node_type VARCHAR(40) PATH '$[2]')) row_data;

INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, sort_order)
SELECT @tang, row_data.node_type, row_data.display_name, row_data.gender, row_data.node_no * 10
FROM JSON_TABLE(@tang_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, display_name VARCHAR(80) PATH '$[0]', node_type VARCHAR(40) PATH '$[2]', gender VARCHAR(20) PATH '$[3]')) row_data;
INSERT INTO seed_showcase_members
SELECT @tang, row_data.node_no, row_data.parent_no, row_data.node_type
FROM JSON_TABLE(@tang_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, parent_no INT PATH '$[1]' NULL ON EMPTY, node_type VARCHAR(40) PATH '$[2]')) row_data;

INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, sort_order)
SELECT @song, row_data.node_type, row_data.display_name, row_data.gender, row_data.node_no * 10
FROM JSON_TABLE(@song_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, display_name VARCHAR(80) PATH '$[0]', node_type VARCHAR(40) PATH '$[2]', gender VARCHAR(20) PATH '$[3]')) row_data;
INSERT INTO seed_showcase_members
SELECT @song, row_data.node_no, row_data.parent_no, row_data.node_type
FROM JSON_TABLE(@song_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, parent_no INT PATH '$[1]' NULL ON EMPTY, node_type VARCHAR(40) PATH '$[2]')) row_data;

INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT seed.showcase_family_id, parent.id, child.id,
       CASE WHEN seed.member_type = 'SPOUSE' THEN 'SPOUSE' ELSE 'PARENT_CHILD' END
FROM seed_showcase_members seed
JOIN showcase_family_members child ON child.showcase_family_id = seed.showcase_family_id AND child.sort_order = seed.node_no * 10
JOIN showcase_family_members parent ON parent.showcase_family_id = seed.showcase_family_id AND parent.sort_order = seed.parent_no * 10
WHERE seed.parent_no IS NOT NULL;

DROP TEMPORARY TABLE seed_showcase_members;

-- 展示册只保留主线和较高辨识度人物，避免把大量仅有姓名的宗室节点作为内容展示。
INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, sort_order)
SELECT id, 'SPOUSE', '武则天', 'FEMALE', 9950 FROM showcase_family_configs WHERE display_name = '唐高祖李渊五代家系';
INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT family.id, li_zhi.id, wu_zetian.id, 'SPOUSE'
FROM showcase_family_configs family
JOIN showcase_family_members li_zhi ON li_zhi.showcase_family_id = family.id AND li_zhi.display_name = '李治'
JOIN showcase_family_members wu_zetian ON wu_zetian.showcase_family_id = family.id AND wu_zetian.display_name = '武则天'
WHERE family.display_name = '唐高祖李渊五代家系';

DELETE relation FROM showcase_family_relationships relation
JOIN showcase_family_members member ON member.id = relation.from_member_id OR member.id = relation.to_member_id
JOIN showcase_family_configs family ON family.id = member.showcase_family_id
WHERE (family.display_name = '汉高祖刘邦五代家系' AND member.display_name NOT IN ('刘邦','吕雉','刘肥','刘盈','刘如意','刘恒','刘长','窦皇后','刘启','刘武（文帝子）','刘嫖','刘彻','刘据','刘弗陵'))
   OR (family.display_name = '唐高祖李渊五代家系' AND member.display_name NOT IN ('李渊','窦皇后','李建成','李世民','李元吉','李智云','长孙皇后','杨妃','李承乾','李恪','李泰','李治','武则天','李弘','李贤','李显','李旦','太平公主'))
   OR (family.display_name = '宋太祖赵匡胤五代家系' AND member.display_name NOT IN ('赵匡胤','贺氏','王氏','赵德昭','赵德芳','魏国长公主','鲁国长公主','赵惟正','赵惟吉','焦氏','符氏','赵惟宪','赵惟能','赵从郁','赵从式','赵世将','赵令譮'));
DELETE member FROM showcase_family_members member
JOIN showcase_family_configs family ON family.id = member.showcase_family_id
WHERE (family.display_name = '汉高祖刘邦五代家系' AND member.display_name NOT IN ('刘邦','吕雉','刘肥','刘盈','刘如意','刘恒','刘长','窦皇后','刘启','刘武（文帝子）','刘嫖','刘彻','刘据','刘弗陵'))
   OR (family.display_name = '唐高祖李渊五代家系' AND member.display_name NOT IN ('李渊','窦皇后','李建成','李世民','李元吉','李智云','长孙皇后','杨妃','李承乾','李恪','李泰','李治','武则天','李弘','李贤','李显','李旦','太平公主'))
   OR (family.display_name = '宋太祖赵匡胤五代家系' AND member.display_name NOT IN ('赵匡胤','贺氏','王氏','赵德昭','赵德芳','魏国长公主','鲁国长公主','赵惟正','赵惟吉','焦氏','符氏','赵惟宪','赵惟能','赵从郁','赵从式','赵世将','赵令譮'));

-- 补充三册中辨识度高、且能直接接入主线的核心人物。
INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, identity_text, sort_order)
SELECT id, 'SPOUSE', '卫子夫', 'FEMALE', '汉武帝皇后', 9940 FROM showcase_family_configs WHERE display_name = '汉高祖刘邦五代家系';
INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT family.id, liu_che.id, wei_zifu.id, 'SPOUSE' FROM showcase_family_configs family
JOIN showcase_family_members liu_che ON liu_che.showcase_family_id = family.id AND liu_che.display_name = '刘彻'
JOIN showcase_family_members wei_zifu ON wei_zifu.showcase_family_id = family.id AND wei_zifu.display_name = '卫子夫'
WHERE family.display_name = '汉高祖刘邦五代家系';
INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, identity_text, sort_order)
SELECT id, 'LINEAGE_MEMBER', '李重俊', 'MALE', '节愍太子', 9940 FROM showcase_family_configs WHERE display_name = '唐高祖李渊五代家系';
INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT family.id, li_xian.id, li_chongjun.id, 'PARENT_CHILD' FROM showcase_family_configs family
JOIN showcase_family_members li_xian ON li_xian.showcase_family_id = family.id AND li_xian.display_name = '李显'
JOIN showcase_family_members li_chongjun ON li_chongjun.showcase_family_id = family.id AND li_chongjun.display_name = '李重俊'
WHERE family.display_name = '唐高祖李渊五代家系';

-- 汉朝：补入皇室中辨识度较高的妃后、公主、诸侯王与太子支系，展示册合计 30 人。
INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, identity_text, sort_order) VALUES
(@han,'SPOUSE','戚夫人','FEMALE','汉高祖宠妃',9930),(@han,'SPOUSE','薄姬','FEMALE','汉文帝生母',9920),(@han,'LINEAGE_MEMBER','鲁元公主','FEMALE','汉高祖长女',9910),(@han,'LINEAGE_MEMBER','张偃','MALE','鲁元公主之子',9900),(@han,'LINEAGE_MEMBER','刘安','MALE','淮南王',9890),(@han,'SPOUSE','栗姬','FEMALE','汉景帝宠妃',9880),(@han,'SPOUSE','王娡','FEMALE','汉武帝生母',9870),(@han,'LINEAGE_MEMBER','刘荣','MALE','临江闵王',9860),(@han,'LINEAGE_MEMBER','刘胜','MALE','中山靖王',9850),(@han,'SPOUSE','陈阿娇','FEMALE','汉武帝首任皇后',9840),(@han,'SPOUSE','李夫人','FEMALE','汉武帝宠妃',9830),(@han,'SPOUSE','钩弋夫人','FEMALE','汉昭帝生母',9820),(@han,'LINEAGE_MEMBER','刘旦','MALE','燕剌王',9810),(@han,'LINEAGE_MEMBER','刘胥','MALE','广陵厉王',9800),(@han,'LINEAGE_MEMBER','刘髆','MALE','昌邑哀王',9790);
INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT @han, parent.id, child.id, relation.kind FROM (
  SELECT '刘邦' parent_name,'戚夫人' child_name,'SPOUSE' kind UNION ALL SELECT '刘邦','薄姬','SPOUSE' UNION ALL SELECT '刘邦','鲁元公主','PARENT_CHILD' UNION ALL SELECT '鲁元公主','张偃','PARENT_CHILD' UNION ALL SELECT '刘长','刘安','PARENT_CHILD' UNION ALL SELECT '刘启','栗姬','SPOUSE' UNION ALL SELECT '刘启','王娡','SPOUSE' UNION ALL SELECT '刘启','刘荣','PARENT_CHILD' UNION ALL SELECT '刘启','刘胜','PARENT_CHILD' UNION ALL SELECT '刘彻','陈阿娇','SPOUSE' UNION ALL SELECT '刘彻','李夫人','SPOUSE' UNION ALL SELECT '刘彻','钩弋夫人','SPOUSE' UNION ALL SELECT '刘彻','刘旦','PARENT_CHILD' UNION ALL SELECT '刘彻','刘胥','PARENT_CHILD' UNION ALL SELECT '刘彻','刘髆','PARENT_CHILD'
) relation JOIN showcase_family_members parent ON parent.showcase_family_id=@han AND parent.display_name=relation.parent_name JOIN showcase_family_members child ON child.showcase_family_id=@han AND child.display_name=relation.child_name;

-- 唐朝：补入唐初、武周及开元前后的核心皇室人物，展示册合计 32 人。
INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, identity_text, sort_order) VALUES
(@tang,'LINEAGE_MEMBER','李元景','MALE','荆王',9930),(@tang,'LINEAGE_MEMBER','李元嘉','MALE','汉王',9920),(@tang,'SPOUSE','阴妃','FEMALE','唐太宗妃嫔',9910),(@tang,'LINEAGE_MEMBER','李贞','MALE','越王',9900),(@tang,'LINEAGE_MEMBER','李丽质','FEMALE','长乐公主',9890),(@tang,'SPOUSE','王皇后','FEMALE','唐高宗皇后',9880),(@tang,'SPOUSE','萧淑妃','FEMALE','唐高宗妃嫔',9870),(@tang,'SPOUSE','韦皇后','FEMALE','唐中宗皇后',9860),(@tang,'LINEAGE_MEMBER','李重茂','MALE','唐殇帝',9850),(@tang,'LINEAGE_MEMBER','李隆基','MALE','唐玄宗',9840),(@tang,'LINEAGE_MEMBER','李成器','MALE','宁王',9830),(@tang,'SPOUSE','王皇后（唐玄宗）','FEMALE','唐玄宗皇后',9820),(@tang,'SPOUSE','杨贵妃','FEMALE','唐玄宗贵妃',9810);
INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT @tang, parent.id, child.id, relation.kind FROM (
  SELECT '李渊' parent_name,'李元景' child_name,'PARENT_CHILD' kind UNION ALL SELECT '李渊','李元嘉','PARENT_CHILD' UNION ALL SELECT '李世民','阴妃','SPOUSE' UNION ALL SELECT '李世民','李贞','PARENT_CHILD' UNION ALL SELECT '李世民','李丽质','PARENT_CHILD' UNION ALL SELECT '李治','王皇后','SPOUSE' UNION ALL SELECT '李治','萧淑妃','SPOUSE' UNION ALL SELECT '李显','韦皇后','SPOUSE' UNION ALL SELECT '李显','李重茂','PARENT_CHILD' UNION ALL SELECT '李旦','李隆基','PARENT_CHILD' UNION ALL SELECT '李旦','李成器','PARENT_CHILD' UNION ALL SELECT '李隆基','王皇后（唐玄宗）','SPOUSE' UNION ALL SELECT '李隆基','杨贵妃','SPOUSE'
) relation JOIN showcase_family_members parent ON parent.showcase_family_id=@tang AND parent.display_name=relation.parent_name JOIN showcase_family_members child ON child.showcase_family_id=@tang AND child.display_name=relation.child_name;

UPDATE showcase_family_members SET identity_text = CASE display_name
  WHEN '刘邦' THEN '汉高祖' WHEN '吕雉' THEN '汉高后' WHEN '刘肥' THEN '齐悼惠王' WHEN '刘盈' THEN '汉惠帝' WHEN '刘如意' THEN '赵隐王' WHEN '刘恒' THEN '汉文帝' WHEN '刘长' THEN '淮南厉王' WHEN '窦皇后' THEN '汉文帝皇后' WHEN '刘启' THEN '汉景帝' WHEN '刘武（文帝子）' THEN '梁孝王' WHEN '刘嫖' THEN '馆陶长公主' WHEN '刘彻' THEN '汉武帝' WHEN '刘据' THEN '戾太子' WHEN '刘弗陵' THEN '汉昭帝'
  WHEN '李渊' THEN '唐高祖' WHEN '李建成' THEN '隐太子' WHEN '李世民' THEN '唐太宗' WHEN '李元吉' THEN '巢刺王' WHEN '李智云' THEN '楚王' WHEN '长孙皇后' THEN '文德皇后' WHEN '杨妃' THEN '隋炀帝之女' WHEN '李承乾' THEN '恒山愍王' WHEN '李恪' THEN '吴王' WHEN '李泰' THEN '魏王' WHEN '李治' THEN '唐高宗' WHEN '武则天' THEN '武周皇帝' WHEN '李弘' THEN '孝敬皇帝' WHEN '李贤' THEN '章怀太子' WHEN '李显' THEN '唐中宗' WHEN '李旦' THEN '唐睿宗' WHEN '太平公主' THEN '唐高宗之女'
  WHEN '赵匡胤' THEN '宋太祖' WHEN '贺氏' THEN '孝惠皇后' WHEN '王氏' THEN '孝明皇后' WHEN '赵德昭' THEN '燕懿王' WHEN '赵德芳' THEN '秦康惠王' WHEN '魏国长公主' THEN '宋太祖长女' WHEN '鲁国长公主' THEN '宋太祖女' WHEN '赵惟正' THEN '同安郡王' WHEN '赵惟吉' THEN '冀王' WHEN '焦氏' THEN '赵德芳配偶' WHEN '符氏' THEN '赵德芳配偶' WHEN '赵惟宪' THEN '英国公' WHEN '赵惟能' THEN '南康郡公' WHEN '赵从郁' THEN '新兴侯' WHEN '赵从式' THEN '安定郡王' WHEN '赵世将' THEN '华阴侯' WHEN '赵令譮' THEN '庆国公'
  ELSE identity_text END;
UPDATE showcase_family_members member JOIN showcase_family_configs family ON family.id = member.showcase_family_id
SET member.identity_text = '唐高祖皇后'
WHERE family.display_name = '唐高祖李渊五代家系' AND member.display_name = '窦皇后';

-- 宋朝帝王世系：为准确表示宋太祖与宋太宗两支，将共同父亲赵弘殷设为根节点。
DELETE FROM showcase_family_relationships WHERE showcase_family_id = @song;
DELETE FROM showcase_family_members WHERE showcase_family_id = @song;
UPDATE showcase_family_configs
SET summary = '历史示例 · 根据公开史料整理。以宋宣祖赵弘殷为共同根节点，呈现宋太祖、宋太宗两支及主要帝王世系。'
WHERE id = @song;
SET @song_imperial_seed = '[
["赵弘殷",null,"LINEAGE_MEMBER","MALE","宋宣祖"],["杜太后",1,"SPOUSE","FEMALE","昭宪太后"],
["赵匡胤",1,"LINEAGE_MEMBER","MALE","宋太祖"],["赵光义",1,"LINEAGE_MEMBER","MALE","宋太宗"],["赵廷美",1,"LINEAGE_MEMBER","MALE","魏王"],
["赵德芳",3,"LINEAGE_MEMBER","MALE","秦康惠王"],["赵惟宪",6,"LINEAGE_MEMBER","MALE","英国公"],["赵从郁",7,"LINEAGE_MEMBER","MALE","新兴侯"],["赵世将",8,"LINEAGE_MEMBER","MALE","华阴侯"],["赵令譮",9,"LINEAGE_MEMBER","MALE","庆国公"],["赵子偁",10,"LINEAGE_MEMBER","MALE","秀王"],["赵昚",11,"LINEAGE_MEMBER","MALE","宋孝宗"],["赵惇",12,"LINEAGE_MEMBER","MALE","宋光宗"],["赵扩",13,"LINEAGE_MEMBER","MALE","宋宁宗"],
["赵恒",4,"LINEAGE_MEMBER","MALE","宋真宗"],["赵祯",15,"LINEAGE_MEMBER","MALE","宋仁宗"],["赵允让",4,"LINEAGE_MEMBER","MALE","濮安懿王"],["赵曙",17,"LINEAGE_MEMBER","MALE","宋英宗"],["赵顼",18,"LINEAGE_MEMBER","MALE","宋神宗"],["赵佶",19,"LINEAGE_MEMBER","MALE","宋徽宗"],["赵桓",20,"LINEAGE_MEMBER","MALE","宋钦宗"],["赵构",20,"LINEAGE_MEMBER","MALE","宋高宗"],
["贺皇后",3,"SPOUSE","FEMALE","宋太祖皇后"],["王皇后",3,"SPOUSE","FEMALE","宋太祖皇后"],["李皇后",4,"SPOUSE","FEMALE","宋太宗皇后"],["刘娥",15,"SPOUSE","FEMALE","章献明肃皇后"],["曹皇后",16,"SPOUSE","FEMALE","慈圣光献皇后"],["高滔滔",18,"SPOUSE","FEMALE","宣仁圣烈皇后"],["向皇后",19,"SPOUSE","FEMALE","钦圣宪肃皇后"],["郑皇后",20,"SPOUSE","FEMALE","显肃皇后"],["吴皇后",22,"SPOUSE","FEMALE","宪圣慈烈皇后"],["夏皇后",12,"SPOUSE","FEMALE","成穆皇后"],["李凤娘",13,"SPOUSE","FEMALE","慈懿皇后"],["杨皇后",14,"SPOUSE","FEMALE","恭圣仁烈皇后"]]';
CREATE TEMPORARY TABLE seed_song_imperial (
  node_no INT NOT NULL PRIMARY KEY,
  parent_no INT NULL,
  member_type VARCHAR(40) NOT NULL
);
INSERT INTO showcase_family_members (showcase_family_id, member_type, display_name, gender, identity_text, sort_order)
SELECT @song, item.member_type, item.display_name, item.gender, item.identity_text, item.node_no * 10
FROM JSON_TABLE(@song_imperial_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, display_name VARCHAR(80) PATH '$[0]', member_type VARCHAR(40) PATH '$[2]', gender VARCHAR(20) PATH '$[3]', identity_text VARCHAR(120) PATH '$[4]')) item;
INSERT INTO seed_song_imperial
SELECT item.node_no, item.parent_no, item.member_type
FROM JSON_TABLE(@song_imperial_seed, '$[*]' COLUMNS (node_no FOR ORDINALITY, parent_no INT PATH '$[1]' NULL ON EMPTY, member_type VARCHAR(40) PATH '$[2]')) item;
INSERT INTO showcase_family_relationships (showcase_family_id, from_member_id, to_member_id, relationship_type)
SELECT @song, parent.id, child.id, CASE WHEN seed.member_type = 'SPOUSE' THEN 'SPOUSE' ELSE 'PARENT_CHILD' END
FROM seed_song_imperial seed
JOIN showcase_family_members child ON child.showcase_family_id = @song AND child.sort_order = seed.node_no * 10
JOIN showcase_family_members parent ON parent.showcase_family_id = @song AND parent.sort_order = seed.parent_no * 10
WHERE seed.parent_no IS NOT NULL;
DROP TEMPORARY TABLE seed_song_imperial;
COMMIT;
