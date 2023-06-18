package api

import (
	"context"
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
//					--start-time "$(date -u -v -5M +'%Y-%m-%dT%H:%M:%SZ')" \
//					--end-time "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
//					--period 60 \
//					--dimensions Name=ClusterName,Value={clusterName} Name=ServiceName,Value={serviceName}
//
// Get last 5 minute, granularity 60s CPUUtilization
func (store *Store) getCPU(cluster, service *string) ([]types.Datapoint, error) {
	statisticsInput := store.getStatisticsInput(cluster, service)
	statisticsInput.MetricName = aws.String(CPU)
	metricOutput, err := store.cloudWatch.GetMetricStatistics(context.TODO(), statisticsInput)

	if err != nil {
		logger.Printf("aws failed to %s, cluster: \"%s\", service: \"%s\", err: %v\n", CPU, *cluster, *service, err)
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
//					--start-time "$(date -u -v -5M +'%Y-%m-%dT%H:%M:%SZ')" \
//					--end-time "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
//					--period 60 \
//					--dimensions Name=ClusterName,Value={clusterName} Name=ServiceName,Value={serviceName}
//
// Get last 5 minute, granularity 60s MemoryUtilization
func (store *Store) getMemory(cluster, service *string) ([]types.Datapoint, error) {
	statisticsInput := store.getStatisticsInput(cluster, service)
	statisticsInput.MetricName = aws.String(Memory)
	metricOutput, err := store.cloudWatch.GetMetricStatistics(context.TODO(), statisticsInput)

	if err != nil {
		logger.Printf("aws failed to %s, cluster: \"%s\", service: \"%s\", err: %v\n", Memory, *cluster, *service, err)
		return nil, err
	}

	return metricOutput.Datapoints, nil
}

func (store *Store) getStatisticsInput(cluster, service *string) *cloudwatch.GetMetricStatisticsInput {
	store.getCloudWatchClient()

	statistic := []types.Statistic{types.StatisticAverage}
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	now := time.Now()
	period := int32(60)
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
		StartTime:  aws.Time(fiveMinutesAgo),
		EndTime:    aws.Time(now),
		Period:     aws.Int32(period),
		Dimensions: dimensions,
	}
}

func (store *Store) getCloudWatchClient() {
	if store.cloudWatch == nil {
		store.cloudWatch = cloudwatch.NewFromConfig(*store.Config)
	}
}
