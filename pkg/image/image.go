package image

// NOTE(SergeyCherepiuk): Only pulled images are supported for now
type Image struct {
	Id  string `yaml:"-"`
	Ref string // registry/tag:version
}
