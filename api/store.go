package api

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/keidarcy/e1s/util"
)

var logger = util.Logger

type Store struct {
	*aws.Config
	ecs            *ecs.Client
	cloudwatch     *cloudwatch.Client
	cloudwatchlogs *cloudwatchlogs.Client
	autoScaling    *applicationautoscaling.Client
}

func NewStore() *Store {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		logger.Printf("e1s - aws unable to load SDK config, error: %v\n", err)
	}
	ecsClient := ecs.NewFromConfig(cfg)
	return &Store{
		Config: &cfg,
		ecs:    ecsClient,
	}
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
