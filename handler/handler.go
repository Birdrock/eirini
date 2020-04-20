package handler

import (
	"net/http"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/lager"
	"github.com/julienschmidt/httprouter"
)

func New(bifrost eirini.Bifrost,
	buildpackStager eirini.Stager,
	dockerStager eirini.Stager,
	lager lager.Logger) http.Handler {
	handler := httprouter.New()

	appHandler := NewAppHandler(bifrost, lager)
	stageHandler := NewStageHandler(bifrost, buildpackStager, dockerStager, lager)
	taskHandler := NewTaskHandler(lager, bifrost)

	registerAppsEndpoints(handler, appHandler)
	registerStageEndpoints(handler, stageHandler)
	registerTaskEndpoints(handler, taskHandler)

	return handler
}

func registerAppsEndpoints(handler *httprouter.Router, appHandler *App) {
	handler.GET("/apps", appHandler.List)
	handler.PUT("/apps/:process_guid", appHandler.Desire)
	handler.POST("/apps/:process_guid", appHandler.Update)
	handler.PUT("/apps/:process_guid/:version_guid/stop", appHandler.Stop)
	handler.PUT("/apps/:process_guid/:version_guid/stop/:instance", appHandler.StopInstance)
	handler.GET("/apps/:process_guid/:version_guid/instances", appHandler.GetInstances)
	handler.GET("/apps/:process_guid/:version_guid", appHandler.GetApp)
}

func registerStageEndpoints(handler *httprouter.Router, stageHandler *Stage) {
	handler.POST("/stage/:staging_guid", stageHandler.Stage)
	handler.PUT("/stage/:staging_guid/completed", stageHandler.StagingComplete)
}

func registerTaskEndpoints(handler *httprouter.Router, taskHandler *Task) {
	handler.POST("/tasks/:task_guid", taskHandler.Run)
}
