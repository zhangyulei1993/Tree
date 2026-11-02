package enums

type RecordStatus string

const (
	StatusActive              RecordStatus = "ACTIVE"
	StatusInactive            RecordStatus = "INACTIVE"
	StatusDisabled            RecordStatus = "DISABLED"
	StatusDeleted             RecordStatus = "DELETED"
	StatusCancelled           RecordStatus = "CANCELLED"
	StatusMerged              RecordStatus = "MERGED"
	StatusPending             RecordStatus = "PENDING"
	StatusRejected            RecordStatus = "REJECTED"
	StatusApproved            RecordStatus = "APPROVED"
	StatusExpired             RecordStatus = "EXPIRED"
	StatusUsed                RecordStatus = "USED"
	StatusNormal              RecordStatus = "NORMAL"
	StatusDissolutionCooldown RecordStatus = "DISSOLUTION_COOLDOWN"
	StatusDissolved           RecordStatus = "DISSOLVED"
	StatusPendingClaim        RecordStatus = "PENDING_CLAIM"
	StatusPendingBind         RecordStatus = "PENDING_PHONE_BIND"
	StatusPrivate             RecordStatus = "PRIVATE"
	StatusTakenDown           RecordStatus = "TAKEN_DOWN"
	StatusSuccess             RecordStatus = "SUCCESS"
	StatusFailed              RecordStatus = "FAILED"
	StatusLineageMember       RecordStatus = "LINEAGE_MEMBER"
)

type OperatorType string

const (
	OperatorTypeAdmin  OperatorType = "ADMIN"
	OperatorTypeUser   OperatorType = "USER"
	OperatorTypeSystem OperatorType = "SYSTEM"
)

type TreeMode string

const (
	TreeModeListTree TreeMode = "LIST_TREE"
)

type Gender string

const (
	GenderMale    Gender = "MALE"
	GenderFemale  Gender = "FEMALE"
	GenderUnknown Gender = "UNKNOWN"
)

type UserBindingPolicy string

const (
	UserBindingOptional    UserBindingPolicy = "OPTIONAL"
	UserBindingRequired    UserBindingPolicy = "REQUIRED"
	UserBindingNotRequired UserBindingPolicy = "NOT_REQUIRED"
)
