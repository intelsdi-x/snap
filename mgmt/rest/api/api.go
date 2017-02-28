package api

import (
	"github.com/julienschmidt/httprouter"
)

type API interface {
	GetRoutes() []Route
	BindMetricManager(Metrics)
	BindTaskManager(Tasks)
	BindTribeManager(Tribe)
	BindConfigManager(Config)
}

type Route struct {
	Method, Path string
	Handle       httprouter.Handle
}
