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
