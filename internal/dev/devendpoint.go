package dev

type LocalEndpoint struct {
	Module            string
	Endpoint          string
	DebugPort         int
	Language          string
	HotReloadEndpoint string
	HotReloadVersion  int64
}
