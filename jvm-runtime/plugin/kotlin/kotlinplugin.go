package kotlin

import (
	"github.com/block/ftl/jvm-runtime/kotlin"
	"github.com/block/ftl/jvm-runtime/plugin/common"
)

func NewCmdService() *common.CmdService {
	return common.NewCmdService(kotlin.Files())
}
