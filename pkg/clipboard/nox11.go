//go:build linux && !x11

package clipboard

import "fmt"

func Init() (supported bool, err error) {
	return false, nil
}

func WriteStr(value string) {
	fmt.Println(value)
}
