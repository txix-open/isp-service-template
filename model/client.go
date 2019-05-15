package model

import (
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/logger"
)

var (
	DbClient = database.NewRxDbClient(
		database.WithSchemaEnsuring(),
		database.WithSchemaAutoInjecting(),
		database.WithMigrationsEnsuring(),
		database.WithInitializingErrorHandler(func(err *database.ErrorEvent) {
			logger.Fatal(err)
		}),
	)
	ObjectRep = ObjectRepository{DbClient}
)
