package dev

const RunnerVersionComment = "FTL Hot Reload Runner Version: "

type LocalEndpoint struct {
	Module            string
	Endpoint          string
	DebugPort         int
	Language          string
	HotReloadEndpoint string
}
