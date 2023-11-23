package main

import (
	"github.com/integration-system/isp-kit/bootstrap"
	"github.com/integration-system/isp-kit/shutdown"
	"msp-service-template/assembly"
	"msp-service-template/conf"
	"msp-service-template/routes"
)

var (
	version = "1.0.0"
)

// @title msp-service-template
// @version 1.0.0
// @description Шаблон сервиса

// @license.name GNU GPL v3.0

// @host localhost:9000
// @BasePath /api/msp-service-template

//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json docs/docs.go
func main() {
	boot := bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors())
	app := boot.App
	logger := app.Logger()

	assembly, err := assembly.New(boot)
	if err != nil {
		boot.Fatal(err)
	}
	app.AddRunners(assembly.Runners()...)
	app.AddClosers(assembly.Closers()...)

	shutdown.On(func() {
		logger.Info(app.Context(), "starting shutdown")
		app.Shutdown()
		logger.Info(app.Context(), "shutdown completed")
	})

	err = app.Run()
	if err != nil {
		boot.Fatal(err)
	}
}
