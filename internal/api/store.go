package api

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Store struct {
	*aws.Config
	Region         string
	Profile        string
	ecs            *ecs.Client
	cloudwatch     *cloudwatch.Client
	cloudwatchlogs *cloudwatchlogs.Client
	autoScaling    *applicationautoscaling.Client
	ssm            *ssm.Client
}

func NewStore() (*Store, error) {
	profile := os.Getenv("AWS_PROFILE")
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		slog.Error("failed to load aws SDK config", "error", err)
		return nil, err
	}
	ecsClient := ecs.NewFromConfig(cfg)
	slog.Info("load config", slog.String("AWS_PROFILE", profile), slog.String("AWS_REGION", cfg.Region))
	return &Store{
		Config:  &cfg,
		Region:  cfg.Region,
		Profile: profile,
		ecs:     ecsClient,
	}, nil
}

func (store *Store) initCloudwatchClient() {
	if store.cloudwatch == nil {
		store.cloudwatch = cloudwatch.NewFromConfig(*store.Config)
	}
}

func (store *Store) initCloudwatchlogsClient() {
	if store.cloudwatchlogs == nil {
		store.cloudwatchlogs = cloudwatchlogs.NewFromConfig(*store.Config)
	}
}

func (store *Store) initSsmClient() {
	if store.ssm == nil {
		store.ssm = ssm.NewFromConfig(*store.Config)
	}
}

// SwitchProfile switches to a different AWS profile and reinitializes clients
func (store *Store) SwitchProfile(profileName string) error {
	// Set the profile in environment variable
	if profileName == "default" {
		os.Unsetenv("AWS_PROFILE")
	} else {
		os.Setenv("AWS_PROFILE", profileName)
	}

	// Load new configuration with the updated profile
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load aws SDK config with new profile", "profile", profileName, "error", err)
		return err
	}

	// Update store with new configuration
	store.Config = &cfg
	store.Region = cfg.Region
	store.Profile = profileName
	if profileName == "default" {
		store.Profile = ""
	}

	// Reinitialize all clients with new configuration
	store.ecs = ecs.NewFromConfig(cfg)
	store.cloudwatch = nil       // Will be lazy-loaded with new config
	store.cloudwatchlogs = nil   // Will be lazy-loaded with new config
	store.ssm = nil              // Will be lazy-loaded with new config
	store.autoScaling = nil      // Will be lazy-loaded with new config

	slog.Info("switched AWS profile", slog.String("AWS_PROFILE", store.Profile), slog.String("AWS_REGION", store.Region))
	return nil
}

func (store *Store) initAutoScalingClient() {
	if store.autoScaling == nil {
		store.autoScaling = applicationautoscaling.NewFromConfig(*store.Config)
	}
}
