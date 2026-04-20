//go:build !linux || x11

package clipboard

import "golang.design/x/clipboard"

func Init() (supported bool, err error) {
	return true, clipboard.Init()
}

func WriteStr(value string) {
	clipboard.Write(clipboard.FmtText, []byte(value))
}
