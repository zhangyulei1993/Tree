package contentsafety

type Scene int

const (
	SceneProfile Scene = 1
	SceneSocial  Scene = 2
)

const (
	SuggestPass   = "pass"
	SuggestReview = "review"
	SuggestRisky  = "risky"
)

type Field struct {
	Label string
	Value string
}

type CheckInput struct {
	UserID    uint64
	Scene     Scene
	Fields    []Field
	LogAction string
	FamilyID  *uint64
	IP        string
	UserAgent string
}

type Result struct {
	Suggest string
	Label   int
	TraceID string
}
