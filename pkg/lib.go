package pkg

import (
	"time"

	"github.com/nextlinux/ecs-inventory/pkg/connection"
	"github.com/nextlinux/ecs-inventory/pkg/inventory"
	"github.com/nextlinux/ecs-inventory/pkg/logger"
)

var log logger.Logger

// PeriodicallyGetInventoryReport periodically retrieve image results and report/output them according to the configuration.
// Note: Errors do not cause the function to exit, since this is periodically running
func PeriodicallyGetInventoryReport(
	pollingIntervalSeconds int,
	nextlinuxDetails connection.NextlinuxInfo,
	region string,
	quiet, dryRun bool,
) {
	// Fire off a ticker that reports according to a configurable polling interval
	ticker := time.NewTicker(time.Duration(pollingIntervalSeconds) * time.Second)

	for {
		err := inventory.GetInventoryReportsForRegion(region, nextlinuxDetails, quiet, dryRun)
		if err != nil {
			log.Error("Failed to get Inventory Reports for region", err)
		}

		// Wait at least as long as the ticker
		log.Debugf("Start new gather %s", <-ticker.C)
	}
}

func SetLogger(logger logger.Logger) {
	log = logger
}
