package main

import (
	"context"
	"os"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/jvm-runtime/plugin/common"
	"github.com/block/ftl/jvm-runtime/plugin/kotlin"
)

func main() {
	plugin.Start(context.Background(),
		os.Getenv("FTL_NAME"),
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler)
}

func createService(ctx context.Context, config any) (context.Context, *common.Service, error) {
	svc := kotlin.New()
	return ctx, svc, nil
}
