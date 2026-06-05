package enum

type RelationshipType string

const (
	RelationshipTypeParentChild RelationshipType = "PARENT_CHILD"
	RelationshipTypeSpouse      RelationshipType = "SPOUSE"
)

// No SIBLING enum is intentionally defined. Sibling operations must be modeled
// by shared PARENT_CHILD relationships in the service layer.
