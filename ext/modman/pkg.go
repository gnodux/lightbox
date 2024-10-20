package modman

import "fmt"

type Require struct {
	Name    string `json:"name,omitempty" yaml:"name"`
	Version string `json:"version,omitempty" yaml:"version"`
}

func (r *Require) String() string {
	if r.Version == "" {
		return r.Name
	}
	return fmt.Sprintf(pkgNameFmt, r.Name, r.Version)
}
func (r *Require) ZipName() string {
	if r.Version == "" {
		return r.Name + ".zip"
	}
	return fmt.Sprintf(pkgNameFmt, r.Name, r.Version) + ".zip"
}
