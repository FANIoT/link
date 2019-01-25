package grifts

import (
	"github.com/FANIoT/link/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
