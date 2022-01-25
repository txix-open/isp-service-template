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

func main() {
	boot := bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors())
	app := boot.App
	logger := app.Logger()

	assembly, err := assembly.New(boot)
	if err != nil {
		logger.Fatal(app.Context(), err)
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
		app.Shutdown()
		logger.Fatal(app.Context(), err)
	}
}
