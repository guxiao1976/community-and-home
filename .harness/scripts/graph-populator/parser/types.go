package parser

// RelDef defines a relationship between two graph nodes.
// It is shared by all parsers and the populator.
type RelDef struct {
	FromID string
	Type   string
	ToID   string
}
