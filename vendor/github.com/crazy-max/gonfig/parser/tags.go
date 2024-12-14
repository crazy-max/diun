package parser

const (
	// TagLabel allows to apply a custom behavior.
	// - "allowEmpty": allows the creation of a type that is supposed to have children
	// (i.e: struct, pointer of struct, and map), without any children.
	// - "-": ignore the field.
	TagLabel = "label"

	// TagFile allows to apply a custom behavior.
	// - "allowEmpty": allows the creation of a type that is supposed to have children
	// (i.e: struct, pointer of struct, and map), without any children.
	// - "-": ignore the field.
	TagFile = "file"

	// TagLabelSliceAsStruct allows to use a slice of struct by creating one entry into the slice.
	// The value is the substitution name used in the label to access the slice.
	TagLabelSliceAsStruct = "label-slice-as-struct"

	// TagDescription is the documentation for the field.
	// - "-": ignore the field.
	TagDescription = "description"

	// TagLabelAllowEmpty is related to TagLabel.
	TagLabelAllowEmpty = "allowEmpty"
)
