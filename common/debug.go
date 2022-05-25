package common

import (
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func Debug(msg interface{}) {
	glog.Debug(gctx.New(), msg)
}
