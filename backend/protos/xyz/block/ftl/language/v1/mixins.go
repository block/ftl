package languagepb

import (
	"fmt"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

// ErrorsFromProto converts a protobuf ErrorList to a []builderrors.Error.
func ErrorsFromProto(e *ErrorList) []builderrors.Error {
	if e == nil {
		return []builderrors.Error{}
	}
	return slices.Map(e.Errors, ErrorFromProto)
}

func ErrorsToProto(errs []builderrors.Error) *ErrorList {
	return &ErrorList{Errors: slices.Map(errs, ErrorToProto)}
}

// ErrorString formats a languagepb.Error with its position (if available) and message
func ErrorString(err *Error) string {
	if err.Pos != nil {
		return fmt.Sprintf("%s: %s", positionString(err.Pos), err.Msg)
	}
	return err.Msg
}

// ErrorListString formats all errors in a languagepb.ErrorList using builderrors.Error formatting
func ErrorListString(errList *ErrorList) string {
	if errList == nil || len(errList.Errors) == 0 {
		return ""
	}
	errs := ErrorsFromProto(errList)
	result := make([]string, len(errs))
	for i, err := range errs {
		result[i] = err.Error()
	}
	return strings.Join(result, ": ")
}

// positionString formats a languagepb.Position similar to builderrors.Position
func positionString(pos *Position) string {
	if pos == nil {
		return ""
	}
	columnStr := fmt.Sprintf("%d", pos.StartColumn)
	if pos.StartColumn != pos.EndColumn {
		columnStr += fmt.Sprintf("-%d", pos.EndColumn)
	}
	if pos.Filename == "" {
		return fmt.Sprintf("%d:%s", pos.Line, columnStr)
	}
	return fmt.Sprintf("%s:%d:%s", pos.Filename, pos.Line, columnStr)
}

func levelFromProto(level Error_ErrorLevel) builderrors.ErrorLevel {
	switch level {
	case Error_ERROR_LEVEL_INFO:
		return builderrors.INFO
	case Error_ERROR_LEVEL_WARN:
		return builderrors.WARN
	case Error_ERROR_LEVEL_ERROR:
		return builderrors.ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func levelToProto(level builderrors.ErrorLevel) Error_ErrorLevel {
	switch level {
	case builderrors.INFO:
		return Error_ERROR_LEVEL_INFO
	case builderrors.WARN:
		return Error_ERROR_LEVEL_WARN
	case builderrors.ERROR:
		return Error_ERROR_LEVEL_ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func ErrorFromProto(e *Error) builderrors.Error {
	return builderrors.Error{
		Pos:   PosFromProto(e.Pos),
		Msg:   e.Msg,
		Level: levelFromProto(e.Level),
		Type:  builderrors.ErrorType(e.Type),
	}
}

func typeToProto(errType builderrors.ErrorType) Error_ErrorType {
	switch errType {
	case builderrors.UNSPECIFIED:
		return Error_ERROR_TYPE_UNSPECIFIED
	case builderrors.FTL:
		return Error_ERROR_TYPE_FTL
	case builderrors.COMPILER:
		return Error_ERROR_TYPE_COMPILER
	default:
		return Error_ERROR_TYPE_UNSPECIFIED
	}
}

func ErrorToProto(e builderrors.Error) *Error {
	var pos *Position
	if bpos, ok := e.Pos.Get(); ok {
		pos = &Position{
			Filename:    bpos.Filename,
			StartColumn: int64(bpos.StartColumn),
			EndColumn:   int64(bpos.EndColumn),
			Line:        int64(bpos.Line),
		}
	}
	return &Error{
		Msg:   e.Msg,
		Pos:   pos,
		Level: levelToProto(e.Level),
		Type:  typeToProto(e.Type),
	}
}

func PosFromProto(pos *Position) optional.Option[builderrors.Position] {
	if pos == nil {
		return optional.None[builderrors.Position]()
	}
	return optional.Some(builderrors.Position{
		Line:        int(pos.Line),
		StartColumn: int(pos.StartColumn),
		EndColumn:   int(pos.EndColumn),
		Filename:    pos.Filename,
	})
}

// ModuleConfigToProto converts a moduleconfig.AbsModuleConfig to a protobuf ModuleConfig.
//
// Absolute configs are used because relative paths may change resolve differently between parties.
func ModuleConfigToProto(config moduleconfig.AbsModuleConfig) (*ModuleConfig, error) {
	proto := &ModuleConfig{
		Name:       config.Module,
		Dir:        config.Dir,
		DeployDir:  config.DeployDir,
		BuildLock:  config.BuildLock,
		Watch:      config.Watch,
		Language:   config.Language,
		SqlRootDir: config.SQLRootDir,
	}
	if config.Build != "" {
		proto.Build = &config.Build
	}
	if config.DevModeBuild != "" {
		proto.DevModeBuild = &config.DevModeBuild
	}

	langConfigProto, err := structpb.NewStruct(config.LanguageConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal language config")
	}
	proto.LanguageConfig = langConfigProto
	return proto, nil

}

// ModuleConfigFromProto converts a protobuf ModuleConfig to a moduleconfig.AbsModuleConfig.
func ModuleConfigFromProto(proto *ModuleConfig) moduleconfig.AbsModuleConfig {
	config := moduleconfig.AbsModuleConfig{
		Module:       proto.Name,
		Dir:          proto.Dir,
		DeployDir:    proto.DeployDir,
		Watch:        proto.Watch,
		Language:     proto.Language,
		Build:        proto.GetBuild(),
		DevModeBuild: proto.GetDevModeBuild(),
		BuildLock:    proto.BuildLock,
		SQLRootDir:   proto.SqlRootDir,
	}
	if proto.LanguageConfig != nil {
		config.LanguageConfig = proto.LanguageConfig.AsMap()
	}
	return config
}

func ProjectConfigToProto(projConfig projectconfig.Config) *ProjectConfig {
	return &ProjectConfig{
		Dir:    projConfig.Path,
		Name:   projConfig.Name,
		NoGit:  projConfig.NoGit,
		Hermit: projConfig.Hermit,
	}
}

func ProjectConfigFromProto(proto *ProjectConfig) projectconfig.Config {
	return projectconfig.Config{
		Path:   proto.Dir,
		Name:   proto.Name,
		NoGit:  proto.NoGit,
		Hermit: proto.Hermit,
	}
}

// CompilerErrorsToProto converts a slice of builderrors.Error to a protobuf ErrorList with all errors marked as compiler errors.
func CompilerErrorsToProto(errs []builderrors.Error) *ErrorList {
	// Set all errors to compiler type
	for i := range errs {
		errs[i].Type = builderrors.COMPILER
	}
	return ErrorsToProto(errs)
}
