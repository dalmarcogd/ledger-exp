package gomockeq

import "github.com/google/go-cmp/cmp"

func IgnoreFields(fs ...string) cmp.Option {
	return cmp.FilterPath(
		func(p cmp.Path) bool {
			for _, f := range fs {
				if p.String() == f {
					return true
				}
			}
			return false
		},
		cmp.Ignore(),
	)
}
