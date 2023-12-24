package image

// NOTE(SergeyCherepiuk): Only pulled images are supported for now
type Image struct {
	ID  string
	Ref string // registry/tag:version
}
