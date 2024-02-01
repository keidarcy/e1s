package api

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

var logger *log.Logger

type Store struct {
	*aws.Config
	ecs            *ecs.Client
	cloudwatch     *cloudwatch.Client
	cloudwatchlogs *cloudwatchlogs.Client
	autoScaling    *applicationautoscaling.Client
	logger         *log.Logger
}

func NewStore(logr *log.Logger) (*Store, error) {
	logger = logr
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		logger.Printf("e1s - aws unable to load SDK config, error: %v\n", err)
		return nil, err
	}
	ecsClient := ecs.NewFromConfig(cfg)
	return &Store{
		Config: &cfg,
		ecs:    ecsClient,
	}, nil
}

func (store *Store) getCloudwatchClient() {
	if store.cloudwatch == nil {
		store.cloudwatch = cloudwatch.NewFromConfig(*store.Config)
	}
}

func (store *Store) getCloudwatchlogsClient() {
	if store.cloudwatchlogs == nil {
		store.cloudwatchlogs = cloudwatchlogs.NewFromConfig(*store.Config)
	}
}
