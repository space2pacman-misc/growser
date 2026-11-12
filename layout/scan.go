package layout

import "fmt"

func sscan(s string, f *float64) (int, error) {
	return fmt.Sscan(s, f)
}