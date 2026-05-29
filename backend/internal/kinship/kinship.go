package kinship

import (
	"strconv"
	"strings"
	"time"
)

type Person struct {
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Birthday string `json:"birthday"`
	Age      int    `json:"age"`
}

type PathRow struct {
	Relation       string `json:"relation"`
	TargetName     string `json:"target_name"`
	TargetGender   string `json:"target_gender"`
	TargetBirthday string `json:"target_birthday"`
	TargetAge      int    `json:"target_age"`
}

type ResolveRequest struct {
	Self Person    `json:"self"`
	Rows []PathRow `json:"rows"`
}

type DebugStep struct {
	From         string `json:"from"`
	To           string `json:"to"`
	BaseRelation string `json:"base_relation"`
	DetailStep   string `json:"detail_step"`
	DetailName   string `json:"detail_name"`
}

type ResolveResult struct {
	Name       string      `json:"name"`
	Matched    bool        `json:"matched"`
	BasePath   []string    `json:"base_path"`
	DetailPath []string    `json:"detail_path"`
	DebugSteps []DebugStep `json:"debug_steps"`
}

func Resolve(req ResolveRequest) ResolveResult {
	self := normalizePerson(req.Self, "我")
	persons := []Person{self}
	baseSteps := make([]string, 0, len(req.Rows))
	detailSteps := make([]string, 0, len(req.Rows))
	debugSteps := make([]DebugStep, 0, len(req.Rows))

	for _, row := range req.Rows {
		from := persons[len(persons)-1]
		base := normalizeBaseRelation(row.Relation)
		to := Person{
			Name:     row.TargetName,
			Gender:   row.TargetGender,
			Birthday: row.TargetBirthday,
			Age:      row.TargetAge,
		}
		to = normalizePerson(to, BaseRelationText(base))

		detail := baseToDetailStep(base, from, to)

		persons = append(persons, to)
		baseSteps = append(baseSteps, BaseRelationText(base))
		detailSteps = append(detailSteps, detail)
		debugSteps = append(debugSteps, DebugStep{
			From:         from.Name,
			To:           to.Name,
			BaseRelation: BaseRelationText(base),
			DetailStep:   detail,
			DetailName:   DetailStepText(detail),
		})
	}

	name, matched := resolveName(self, persons, detailSteps)

	return ResolveResult{
		Name:       name,
		Matched:    matched,
		BasePath:   baseSteps,
		DetailPath: detailSteps,
		DebugSteps: debugSteps,
	}
}

func RankTitle(rank int, suffix string) string {
	if rank <= 0 {
		return suffix
	}

	if rank == 1 {
		return "大" + suffix
	}

	return chineseNumber(rank) + suffix
}

func baseToDetailStep(base string, from Person, to Person) string {
	switch base {
	case "parent":
		switch normalizeGender(to.Gender) {
		case "male":
			return "father"
		case "female":
			return "mother"
		default:
			return "parent"
		}
	case "child":
		switch normalizeGender(to.Gender) {
		case "male":
			return "son"
		case "female":
			return "daughter"
		default:
			return "child"
		}
	case "sibling":
		return siblingDetailStep(from, to)
	case "spouse":
		return "spouse"
	default:
		return base
	}
}

func siblingDetailStep(from Person, to Person) string {
	gender := normalizeGender(to.Gender)
	compared, known := compareOlder(to, from)

	switch gender {
	case "male":
		if known && compared > 0 {
			return "elder_brother"
		}
		if known && compared < 0 {
			return "younger_brother"
		}
		return "brother"
	case "female":
		if known && compared > 0 {
			return "elder_sister"
		}
		if known && compared < 0 {
			return "younger_sister"
		}
		return "sister"
	default:
		return "sibling"
	}
}

func resolveName(self Person, persons []Person, originalSteps []string) (string, bool) {
	if len(originalSteps) == 0 {
		return "本人", true
	}

	if name, matched := resolveCousin(self, persons, originalSteps); name != "" {
		return name, matched
	}

	steps := reduceSteps(originalSteps)
	key := strings.Join(steps, ">")
	rules := kinshipRules(normalizeGender(self.Gender))

	if name, ok := rules[key]; ok {
		return name, true
	}

	if len(steps) == 1 {
		if name := singleStepName(normalizeGender(self.Gender), steps[0]); name != "" {
			return name, true
		}
	}

	names := make([]string, 0, len(steps))
	for _, step := range steps {
		names = append(names, DetailStepText(step))
	}
	return strings.Join(names, "的"), false
}

func resolveCousin(self Person, persons []Person, steps []string) (string, bool) {
	base, ok := parseCousinBase(self, persons, steps)
	if !ok {
		return "", false
	}

	switch len(base.rest) {
	case 0:
		return cousinTitle(base), base.known
	case 1:
		if base.rest[0] == "spouse" {
			return cousinSpouseTitle(base)
		}
		if isChildStep(base.rest[0]) {
			return cousinChildTitle(base, persons[len(persons)-1])
		}
	case 2:
		if base.rest[0] == "spouse" && isChildStep(base.rest[1]) {
			return cousinChildTitle(base, persons[len(persons)-1])
		}
	}

	return "", false
}

type cousinBase struct {
	prefix string
	gender string
	older  int
	known  bool
	rest   []string
}

func parseCousinBase(self Person, persons []Person, steps []string) (cousinBase, bool) {
	if len(steps) < 3 {
		return cousinBase{}, false
	}

	first, second, third := steps[0], steps[1], steps[2]
	if !isParentStep(first) || !isSiblingStep(second) {
		return cousinBase{}, false
	}

	cousinIndex := 3
	consumed := 3
	if isChildStep(third) {
		cousinIndex = 3
	} else if third == "spouse" && len(steps) >= 4 && isChildStep(steps[3]) {
		cousinIndex = 4
		consumed = 4
	} else {
		return cousinBase{}, false
	}

	if len(persons) <= cousinIndex {
		return cousinBase{}, false
	}

	parentSiblingMale := second == "elder_brother" || second == "younger_brother" || second == "brother"
	prefix := "表"
	if first == "father" && parentSiblingMale {
		prefix = "堂"
	}

	cousin := persons[cousinIndex]
	older, known := compareOlder(cousin, self)
	return cousinBase{
		prefix: prefix,
		gender: normalizeGender(cousin.Gender),
		older:  older,
		known:  known,
		rest:   steps[consumed:],
	}, true
}

func cousinTitle(base cousinBase) string {
	switch base.gender {
	case "male":
		if base.known && base.older > 0 {
			return base.prefix + "哥"
		}
		if base.known && base.older < 0 {
			return base.prefix + "弟"
		}
		return base.prefix + "兄弟"
	case "female":
		if base.known && base.older > 0 {
			return base.prefix + "姐"
		}
		if base.known && base.older < 0 {
			return base.prefix + "妹"
		}
		return base.prefix + "姐妹"
	default:
		return base.prefix + "亲"
	}
}

func cousinSpouseTitle(base cousinBase) (string, bool) {
	switch base.gender {
	case "male":
		if base.known && base.older > 0 {
			return base.prefix + "嫂", true
		}
		if base.known && base.older < 0 {
			return base.prefix + "弟媳", true
		}
		return base.prefix + "兄弟的配偶", false
	case "female":
		if base.known && base.older > 0 {
			return base.prefix + "姐夫", true
		}
		if base.known && base.older < 0 {
			return base.prefix + "妹夫", true
		}
		return base.prefix + "姐妹的配偶", false
	default:
		return base.prefix + "亲的配偶", false
	}
}

func cousinChildTitle(base cousinBase, child Person) (string, bool) {
	childGender := normalizeGender(child.Gender)
	switch base.gender {
	case "male":
		switch childGender {
		case "male":
			return base.prefix + "侄", true
		case "female":
			return base.prefix + "侄女", true
		default:
			return base.prefix + "侄辈", false
		}
	case "female":
		switch childGender {
		case "male":
			return base.prefix + "外甥", true
		case "female":
			return base.prefix + "外甥女", true
		default:
			return base.prefix + "外甥辈", false
		}
	default:
		return base.prefix + "亲的子女", false
	}
}

func reduceSteps(steps []string) []string {
	result := append([]string{}, steps...)
	changed := true

	for changed {
		changed = false

		for i := 0; i <= len(result)-3; i++ {
			if isParentStep(result[i]) && isSiblingStep(result[i+1]) && isParentStep(result[i+2]) {
				result = append(result[:i+1], result[i+2:]...)
				changed = true
				break
			}
		}
		if changed {
			continue
		}

		for i := 0; i <= len(result)-2; i++ {
			if isSiblingStep(result[i]) && isParentStep(result[i+1]) {
				result = append(result[:i], result[i+1:]...)
				changed = true
				break
			}
		}
	}

	return result
}

func kinshipRules(selfGender string) map[string]string {
	rules := map[string]string{
		"father>father": "爷爷",
		"father>mother": "奶奶",
		"mother>father": "外公",
		"mother>mother": "外婆",

		"father>elder_brother":   "伯父",
		"father>younger_brother": "叔叔",
		"father>brother":         "叔伯",
		"father>elder_sister":    "姑妈",
		"father>younger_sister":  "姑姑",
		"father>sister":          "姑姑",

		"mother>elder_brother":   "舅舅",
		"mother>younger_brother": "舅舅",
		"mother>brother":         "舅舅",
		"mother>elder_sister":    "姨妈",
		"mother>younger_sister":  "姨妈",
		"mother>sister":          "姨妈",

		"father>elder_brother>spouse":   "伯母",
		"father>younger_brother>spouse": "婶婶",
		"father>brother>spouse":         "婶婶",
		"father>sister>spouse":          "姑父",
		"mother>brother>spouse":         "舅妈",
		"mother>sister>spouse":          "姨父",

		"elder_brother>spouse":   "嫂子",
		"younger_brother>spouse": "弟媳",
		"elder_sister>spouse":    "姐夫",
		"younger_sister>spouse":  "妹夫",

		"son>spouse":        "儿媳",
		"daughter>spouse":   "女婿",
		"son>son":           "孙子",
		"son>daughter":      "孙女",
		"daughter>son":      "外孙",
		"daughter>daughter": "外孙女",

		"brother>son":              "侄子",
		"elder_brother>son":        "侄子",
		"younger_brother>son":      "侄子",
		"brother>daughter":         "侄女",
		"elder_brother>daughter":   "侄女",
		"younger_brother>daughter": "侄女",
		"sister>son":               "外甥",
		"elder_sister>son":         "外甥",
		"younger_sister>son":       "外甥",
		"sister>daughter":          "外甥女",
		"elder_sister>daughter":    "外甥女",
		"younger_sister>daughter":  "外甥女",
	}

	switch selfGender {
	case "male":
		rules["spouse"] = "妻子"
		rules["spouse>father"] = "岳父"
		rules["spouse>mother"] = "岳母"
		rules["spouse>elder_brother"] = "大舅子"
		rules["spouse>younger_brother"] = "小舅子"
		rules["spouse>brother"] = "舅子"
		rules["spouse>elder_sister"] = "大姨子"
		rules["spouse>younger_sister"] = "小姨子"
		rules["spouse>sister"] = "姨子"
	case "female":
		rules["spouse"] = "丈夫"
		rules["spouse>father"] = "公公"
		rules["spouse>mother"] = "婆婆"
		rules["spouse>elder_brother"] = "大伯子"
		rules["spouse>younger_brother"] = "小叔子"
		rules["spouse>brother"] = "叔伯"
		rules["spouse>elder_sister"] = "大姑子"
		rules["spouse>younger_sister"] = "小姑子"
		rules["spouse>sister"] = "姑子"
	default:
		rules["spouse"] = "配偶"
	}

	return rules
}

func singleStepName(selfGender string, step string) string {
	if step == "spouse" {
		switch selfGender {
		case "male":
			return "妻子"
		case "female":
			return "丈夫"
		default:
			return "配偶"
		}
	}

	names := map[string]string{
		"father":          "父亲",
		"mother":          "母亲",
		"parent":          "父母",
		"son":             "儿子",
		"daughter":        "女儿",
		"child":           "子女",
		"elder_brother":   "哥哥",
		"younger_brother": "弟弟",
		"brother":         "兄弟",
		"elder_sister":    "姐姐",
		"younger_sister":  "妹妹",
		"sister":          "姐妹",
		"sibling":         "兄弟姐妹",
	}

	return names[step]
}

func DetailStepText(step string) string {
	if name := singleStepName("unknown", step); name != "" {
		return name
	}
	return step
}

func BaseRelationText(relation string) string {
	switch normalizeBaseRelation(relation) {
	case "parent":
		return "父母"
	case "child":
		return "子女"
	case "sibling":
		return "兄弟姐妹"
	case "spouse":
		return "配偶"
	default:
		return relation
	}
}

func normalizeBaseRelation(value string) string {
	text := strings.ToLower(strings.TrimSpace(value))
	switch text {
	case "parent", "parents", "father", "mother", "父母", "父亲", "母亲", "爸爸", "妈妈":
		return "parent"
	case "child", "children", "son", "daughter", "子女", "儿子", "女儿":
		return "child"
	case "sibling", "siblings", "brother", "sister", "兄弟姐妹", "兄弟", "姐妹", "哥哥", "弟弟", "姐姐", "妹妹":
		return "sibling"
	case "spouse", "wife", "husband", "配偶", "妻子", "丈夫", "老婆", "老公":
		return "spouse"
	default:
		return text
	}
}

func normalizePerson(person Person, fallbackName string) Person {
	person.Name = strings.TrimSpace(person.Name)
	if person.Name == "" {
		person.Name = fallbackName
	}
	person.Gender = normalizeGender(person.Gender)
	return person
}

func normalizeGender(value string) string {
	text := strings.ToLower(strings.TrimSpace(value))
	switch text {
	case "male", "m", "man", "1", "男", "男性":
		return "male"
	case "female", "f", "woman", "2", "女", "女性":
		return "female"
	default:
		return "unknown"
	}
}

func compareOlder(a Person, b Person) (int, bool) {
	aKey := birthKey(a)
	bKey := birthKey(b)
	if aKey == 0 || bKey == 0 {
		return 0, false
	}
	if aKey < bKey {
		return 1, true
	}
	if aKey > bKey {
		return -1, true
	}
	return 0, true
}

func birthKey(person Person) int {
	birthday := strings.TrimSpace(person.Birthday)
	if birthday != "" {
		value := strings.NewReplacer("年", "-", "月", "-", "日", "", "/", "-", ".", "-").Replace(birthday)
		parts := strings.Split(value, "-")
		year := atoiDefault(part(parts, 0), 0)
		month := atoiDefault(part(parts, 1), 1)
		day := atoiDefault(part(parts, 2), 1)
		if year > 0 {
			return year*10000 + month*100 + day
		}
	}

	if person.Age > 0 {
		year := time.Now().Year() - person.Age
		return year*10000 + 101
	}

	return 0
}

func isParentStep(step string) bool {
	return step == "father" || step == "mother" || step == "parent"
}

func isChildStep(step string) bool {
	return step == "son" || step == "daughter" || step == "child"
}

func isSiblingStep(step string) bool {
	switch step {
	case "elder_brother", "younger_brother", "brother", "elder_sister", "younger_sister", "sister", "sibling":
		return true
	default:
		return false
	}
}

func chineseNumber(n int) string {
	digits := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	if n <= 10 {
		if n == 10 {
			return "十"
		}
		return digits[n]
	}
	if n < 20 {
		return "十" + digits[n%10]
	}
	if n < 100 {
		tens := n / 10
		ones := n % 10
		if ones == 0 {
			return digits[tens] + "十"
		}
		return digits[tens] + "十" + digits[ones]
	}
	return strconv.Itoa(n)
}

func atoiDefault(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func part(parts []string, index int) string {
	if index < 0 || index >= len(parts) {
		return ""
	}
	return parts[index]
}
