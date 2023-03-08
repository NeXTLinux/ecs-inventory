package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/anchore/anchore-ecs-inventory/internal/logger"
	"github.com/anchore/anchore-ecs-inventory/pkg/connection"
	"github.com/anchore/anchore-ecs-inventory/pkg/reporter"
)

// Output the JSON formatted report to stdout
func reportToStdout(report reporter.Report) error {
	enc := json.NewEncoder(os.Stdout)
	// prevent > and < from being escaped in the payload
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("unable to show inventory: %w", err)
	}
	return nil
}

func HandleReport(report reporter.Report, anchoreDetails connection.AnchoreInfo) error {
	if anchoreDetails.IsValid() {
		if err := reporter.Post(report, anchoreDetails); err != nil {
			return fmt.Errorf("unable to report Inventory to Anchore: %w", err)
		}
	} else {
		logger.Log.Debug("Anchore details not specified, not reporting inventory")
	}

	// Encode the report to JSON and output to stdout (maintains same behaviour as when multiple presenters were supported)
	return reportToStdout(report)
}

func GetInventoryReportsForRegion(region string, anchoreDetails connection.AnchoreInfo) error {
	sessConfig := &aws.Config{}
	if region != "" {
		sessConfig.Region = aws.String(region)
	}
	sess, err := session.NewSession(sessConfig)
	if err != nil {
		logger.Log.Error("Failed to create AWS session", err)
	}

	err = checkAWSCredentials(sess)
	if err != nil {
		return err
	}

	ecsClient := ecs.New(sess)

	clusters, err := fetchClusters(ecsClient)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		report, err := GetInventoryReportForCluster(*cluster, ecsClient)
		if err != nil {
			return err
		}

		// Only report if there are images present in the cluster
		if len(report.Results) != 0 {
			err = HandleReport(report, anchoreDetails)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetInventoryReportForCluster is an atomic method for getting in-use image results, for a cluster
func GetInventoryReportForCluster(cluster string, ecsClient *ecs.ECS) (reporter.Report, error) {
	logger.Log.Debug("Found cluster", "cluster", cluster)

	// Fetch tasks in cluster
	tasks, err := fetchTasksFromCluster(ecsClient, cluster)
	if err != nil {
		return reporter.Report{}, err
	}

	results := []reporter.ReportItem{}

	// Must be at least one task to continue
	if len(tasks) == 0 {
		logger.Log.Debug("No tasks found in cluster", "cluster", cluster)
	} else {
		images, err := fetchImagesFromTasks(ecsClient, cluster, tasks)
		if err != nil {
			return reporter.Report{}, err
		}
		results = append(results, reporter.ReportItem{
			Namespace: "", // NOTE The key is Namespace to match the Anchore API but it's actually the cluster ARN
			Images:    images,
		})
	}

	// NOTE: clusterName not used for ECS as the clusternARN (used as the namespace in results payload) provides sufficient
	// unique location data (account, region, clustername)
	return reporter.Report{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Results:       results,
		ClusterName:   cluster,
		InventoryType: "ecs",
	}, nil
}
