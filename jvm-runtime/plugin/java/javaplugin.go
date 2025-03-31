package java

import (
	"github.com/block/ftl/jvm-runtime/java"
	"github.com/block/ftl/jvm-runtime/plugin/common"
)

func NewCmdService() *common.CmdService {
	return common.NewCmdService(java.Files())
}
