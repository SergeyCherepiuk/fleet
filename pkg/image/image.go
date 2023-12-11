package image

import (
	"fmt"
	"path"
)

// NOTE(SergeyCherepiuk): Only pulled images are supported for now
type Image struct {
	ID      string
	Registy string
	Tag     string
	Version string
}

func (i Image) RawRef() string {
	path := path.Join(i.Registy, i.Tag)
	return fmt.Sprintf("%s:%s", path, i.Version)
}
