package enum

type RelationshipType string

const (
	RelationshipTypeParentChild RelationshipType = "PARENT_CHILD"
	RelationshipTypeSpouse      RelationshipType = "SPOUSE"
)

type AddType string

const (
	AddTypeFather  AddType = "ADD_FATHER"
	AddTypeMother  AddType = "ADD_MOTHER"
	AddTypeChild   AddType = "ADD_CHILD"
	AddTypeSpouse  AddType = "ADD_SPOUSE"
	AddTypeSibling AddType = "ADD_SIBLING"
)

type ParentLinkType string

const (
	ParentLinkTypePrimary    ParentLinkType = "PRIMARY"
	ParentLinkTypeStep       ParentLinkType = "STEP"
	ParentLinkTypeAdoptive   ParentLinkType = "ADOPTIVE"
	ParentLinkTypeSuccession ParentLinkType = "SUCCESSION"
	ParentLinkTypeNoteOnly   ParentLinkType = "NOTE_ONLY"
	ParentLinkTypeOther      ParentLinkType = "OTHER"
)

// No SIBLING enum is intentionally defined. Sibling operations must be modeled
// by shared PARENT_CHILD relationships in the service layer.
