package ecg

import (
	"fmt"
	"os"
	"time"

	"github.com/anchore/elastic-container-gatherer/ecg/inventory"
	"github.com/anchore/elastic-container-gatherer/ecg/logger"
	"github.com/anchore/elastic-container-gatherer/ecg/presenter"
	"github.com/anchore/elastic-container-gatherer/ecg/reporter"
	"github.com/anchore/elastic-container-gatherer/internal/config"
	"github.com/anchore/elastic-container-gatherer/internal/log"
)

type channels struct {
	reportItem chan inventory.ReportItem
	errors     chan error
	stopper    chan struct{}
}

func HandleReport(report inventory.Report, cfg *config.Application) error {
	if cfg.AnchoreDetails.IsValid() {
		if err := reporter.Post(report, cfg.AnchoreDetails, cfg); err != nil {
			return fmt.Errorf("unable to report Inventory to Anchore: %w", err)
		}
	} else {
		log.Debug("Anchore details not specified, not reporting inventory")
	}

	if err := presenter.GetPresenter(cfg.PresenterOpt, report).Present(os.Stdout); err != nil {
		return fmt.Errorf("unable to show inventory: %w", err)
	}
	return nil
}

// PeriodicallyGetInventoryReport periodically retrieve image results and report/output them according to the configuration.
// Note: Errors do not cause the function to exit, since this is periodically running
func PeriodicallyGetInventoryReport(cfg *config.Application) {
	// Fire off a ticker that reports according to a configurable polling interval
	ticker := time.NewTicker(time.Duration(cfg.PollingIntervalSeconds) * time.Second)

	for {
		report, err := GetInventoryReport(cfg)
		if err != nil {
			log.Errorf("Failed to get Inventory Report: %w", err)
		} else {
			err := HandleReport(report, cfg)
			if err != nil {
				log.Errorf("Failed to handle Inventory Report: %w", err)
			}
		}

		// Wait at least as long as the ticker
		log.Debugf("Start new gather: %s", <-ticker.C)
	}
}

// GetInventoryReport is an atomic method for getting in-use image results, in parallel for multiple clusters
func GetInventoryReport(cfg *config.Application) (inventory.Report, error) {
	return inventory.Report{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Results:       []inventory.ReportItem{},
		InventoryType: "ecs",
	}, nil
}

func SetLogger(logger logger.Logger) {
	log.Log = logger
}