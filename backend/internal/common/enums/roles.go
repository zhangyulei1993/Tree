package enums

type AdminRole string

const (
	AdminRoleRootAdmin     AdminRole = "ROOT_ADMIN"
	AdminRoleSuperAdmin    AdminRole = "SUPER_ADMIN"
	AdminRolePlatformAdmin AdminRole = "PLATFORM_ADMIN"
)

type FamilyRole string

const (
	FamilyRoleFounder     FamilyRole = "FOUNDER"
	FamilyRoleFamilyAdmin FamilyRole = "FAMILY_ADMIN"
	FamilyRoleMember      FamilyRole = "MEMBER"
)

type RelationshipType string

const (
	RelationshipTypeParentChild RelationshipType = "PARENT_CHILD"
	RelationshipTypeSpouse      RelationshipType = "SPOUSE"
)

type ParentLinkType string

const (
	ParentLinkTypeFather  ParentLinkType = "FATHER"
	ParentLinkTypeMother  ParentLinkType = "MOTHER"
	ParentLinkTypePrimary ParentLinkType = "PRIMARY"
	ParentLinkTypeOther   ParentLinkType = "OTHER"
)
