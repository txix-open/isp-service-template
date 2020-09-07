package model

import (
	"github.com/integration-system/isp-lib/v2/database"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
)

var (
	DbClient = database.NewRxDbClient(
		database.WithSchemaEnsuring(),
		database.WithSchemaAutoInjecting(),
		database.WithMigrationsEnsuring(),
		database.WithInitializingErrorHandler(func(err *database.ErrorEvent) {
			log.WithMetadata(map[string]interface{}{
				"message": err.Error(),
			}).Fatal(stdcodes.InitializingDbError, "error when initializing db connection")
		}),
	)
	//UNCOMMENT BELOW LINE - IF YOU DON'T USE msp-ctl generation tools
	//ObjectRep = ObjectRepository{client: DbClient}

	//**- create "repos" **/
)
