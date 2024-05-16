package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

const (
	Namespace = "AWS/ECS"
	CPU       = "CPUUtilization"
	Memory    = "MemoryUtilization"
)

type MetricsData struct {
	CPUUtilization    []types.Datapoint
	MemoryUtilization []types.Datapoint
}

// Get ECS service metrics(CPU, Memory)
func (store *Store) GetMetrics(cluster, service *string) (*MetricsData, error) {
	cpu, err := store.getCPU(cluster, service)
	if err != nil {
		return nil, err
	}

	memory, err := store.getMemory(cluster, service)
	if err != nil {
		return nil, err
	}

	return &MetricsData{
		CPUUtilization:    cpu,
		MemoryUtilization: memory,
	}, nil

}

// Equivalent to
//
//	aws cloudwatch get-metric-statistics \
//					--namespace AWS/ECS \
//					--metric-name CPUUtilization \
//					--statistics Average \
//					--start-time "$(date -u -v -30M +'%Y-%m-%dT%H:%M:%SZ')" \
//					--end-time "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
//					--period 1800 \
//					--dimensions Name=ClusterName,Value=${clusterName} Name=ServiceName,Value=${serviceName}
//
// Get last 30 minute, granularity 1800 seconds CPUUtilization
func (store *Store) getCPU(cluster, service *string) ([]types.Datapoint, error) {
	statisticsInput := store.getStatisticsInput(cluster, service)
	statisticsInput.MetricName = aws.String(CPU)
	metricOutput, err := store.cloudwatch.GetMetricStatistics(context.TODO(), statisticsInput)

	if err != nil {
		slog.Warn("failed to run aws api", "metrics", CPU, "cluster", *cluster, "service", *service, "error", err)
		return nil, err
	}

	return metricOutput.Datapoints, nil
}

// Equivalent to
//
//	aws cloudwatch get-metric-statistics \
//					--namespace AWS/ECS \
//					--metric-name MemoryUtilization \
//					--statistics Average \
//					--start-time "$(date -u -v -30M +'%Y-%m-%dT%H:%M:%SZ')" \
//					--end-time "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
//					--period 1800 \
//					--dimensions Name=ClusterName,Value=${clusterName} Name=ServiceName,Value=${serviceName}
//
// Get last 30 minute, granularity 1800 seconds CPUUtilization
func (store *Store) getMemory(cluster, service *string) ([]types.Datapoint, error) {
	statisticsInput := store.getStatisticsInput(cluster, service)
	statisticsInput.MetricName = aws.String(Memory)
	metricOutput, err := store.cloudwatch.GetMetricStatistics(context.TODO(), statisticsInput)

	if err != nil {
		slog.Warn("failed to run aws api", "metrics", Memory, "cluster", *cluster, "service", *service, "error", err)
		return nil, err
	}

	return metricOutput.Datapoints, nil
}

func (store *Store) getStatisticsInput(cluster, service *string) *cloudwatch.GetMetricStatisticsInput {
	store.initCloudwatchClient()

	// period := 30
	// granularity := 1800
	statistic := []types.Statistic{types.StatisticAverage}
	halfHourAgo := time.Now().Add(-30 * time.Minute)
	now := time.Now()
	period := int32(1800)
	dimensions := []types.Dimension{
		{
			Name:  aws.String("ClusterName"),
			Value: cluster,
		},
		{
			Name:  aws.String("ServiceName"),
			Value: service,
		},
	}
	return &cloudwatch.GetMetricStatisticsInput{
		MetricName: aws.String("CPUUtilization"),
		Namespace:  aws.String("AWS/ECS"),
		Statistics: statistic,
		StartTime:  aws.Time(halfHourAgo),
		EndTime:    aws.Time(now),
		Period:     aws.Int32(period),
		Dimensions: dimensions,
	}
}
