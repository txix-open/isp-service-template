package main

import (
	"isp-service-template/assembly"
	"isp-service-template/conf"
	"isp-service-template/routes"

	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/shutdown"
)

var (
	version = "1.0.0"
)

// @title isp-service-template
// @version 1.0.0
// @description Шаблон сервиса

// @license.name GNU GPL v3.0

// @host localhost:9000
// @BasePath /api/isp-service-template

//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json docs/docs.go
func main() {
	boot := bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors(), cluster.GrpcTransport)
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
