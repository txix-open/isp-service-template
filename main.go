package main

import (
	"context"
	"os"

	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/bootstrap"
	"github.com/integration-system/isp-lib/v2/config/schema"
	"github.com/integration-system/isp-lib/v2/metric"
	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"msp-service-template/conf"
	_ "msp-service-template/docs"
	"msp-service-template/helper"
	"msp-service-template/model"
)

var (
	version = "0.1.0-dev"
)

// TODO
// @title {название сервиса}
// @version 1.0.0
// @description {описание сервиса}

// @license.name GNU GPL v3.0

// @host localhost:9000
// TODO
// @BasePath /api/msp-service

//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json
func main() {
	bootstrap.
		ServiceBootstrap(&conf.Configuration{}, &conf.RemoteConfig{}).
		DefaultRemoteConfigPath(schema.ResolveDefaultConfigPath("default_remote_config.json")).
		OnLocalConfigLoad(onLocalConfigLoad).
		SocketConfiguration(socketConfiguration).
		OnSocketErrorReceive(onRemoteErrorReceive).
		OnConfigErrorReceive(onRemoteConfigErrorReceive).
		DeclareMe(makeDeclaration).
		/**- create "requires" **/
		OnRemoteConfigReceive(onRemoteConfigReceive).
		OnShutdown(onShutdown).
		Run()
}

func socketConfiguration(cfg interface{}) structure.SocketConfiguration {
	appConfig := cfg.(*conf.Configuration)
	return structure.SocketConfiguration{
		Host:   appConfig.ConfigServiceAddress.IP,
		Port:   appConfig.ConfigServiceAddress.Port,
		Secure: false,
		UrlParams: map[string]string{
			"module_name":   appConfig.ModuleName,
			"instance_uuid": appConfig.InstanceUuid,
		},
	}
}

func onShutdown(_ context.Context, _ os.Signal) {
	backend.StopGrpcServer()
	_ = model.DbClient.Close()
}

func onRemoteConfigReceive(remoteConfig, _ *conf.RemoteConfig) {
	model.DbClient.ReceiveConfiguration(remoteConfig.Database)
	metric.InitHttpServer(remoteConfig.Metrics)
}

func onRemoteErrorReceive(errorMessage map[string]interface{}) {
	log.WithMetadata(errorMessage).Error(stdcodes.ReceiveErrorFromConfig, "error from config service")
}

func onRemoteConfigErrorReceive(errorMessage string) {
	log.WithMetadata(map[string]interface{}{
		"message": errorMessage,
	}).Error(stdcodes.ReceiveErrorOnGettingConfigFromConfig, "error on getting remote configuration")
}

func onLocalConfigLoad(cfg *conf.Configuration) {
	endpoints := helper.GetAllEndpoints(cfg.ModuleName)
	grpcService := backend.NewDefaultService(endpoints)
	backend.StartBackendGrpcServer(cfg.GrpcInnerAddress, grpcService)
}

func makeDeclaration(localConfig interface{}) bootstrap.ModuleInfo {
	cfg := localConfig.(*conf.Configuration)
	endpoints := helper.GetAllEndpoints(cfg.ModuleName)
	return bootstrap.ModuleInfo{
		ModuleName:       cfg.ModuleName,
		ModuleVersion:    version,
		GrpcOuterAddress: cfg.GrpcOuterAddress,
		Endpoints:        endpoints,
	}
}
