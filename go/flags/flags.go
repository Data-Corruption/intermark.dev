package flags

import (
	"os"
	"slices"
)

func Present(arg string) bool {
	return slices.Contains(os.Args, arg)
}

func PresentAny(args ...string) bool {
	for _, arg := range os.Args {
		if slices.Contains(args, arg) {
			return true
		}
	}
	return false
}
