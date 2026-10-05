package api

import (
	"log/slog"
	"path"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/spf13/viper"
)

const defaultStreamTemplate = "{container}/{taskId}"

// CloudwatchLogOverride is a user configured cloudwatch log location, it makes logs
// readable for containers whose log driver is not awslogs(awsfirelens, fluentd
// and so on) or whose log group can not be derived from the task definition.
type CloudwatchLogOverride struct {
	// Task definition family glob, empty matches every family
	Family string `mapstructure:"family"`
	// Container name glob, empty matches every container
	Container string `mapstructure:"container"`
	// Cloudwatch log group name, empty keeps the log driver defaults
	Group string `mapstructure:"group"`
	// Cloudwatch log stream name template, supports {family}, {container} and
	// {taskId}, defaults to "{container}/{taskId}"
	Stream string `mapstructure:"stream"`
}

var (
	logOverridesOnce sync.Once
	logOverrides     []CloudwatchLogOverride
)

func configuredLogOverrides() []CloudwatchLogOverride {
	logOverridesOnce.Do(func() {
		if err := viper.UnmarshalKey("cloudwatch-log-overrides", &logOverrides); err != nil {
			slog.Warn("failed to read cloudwatch-log-overrides config", "error", err)
			logOverrides = nil
		}
		slog.Debug("cloudwatch log overrides", "logOverrides", logOverrides)
	})
	return logOverrides
}

// ResolveLogLocation returns the cloudwatch log group and log stream name of a
// container. A matching user configured log override wins, otherwise fall back
// to the awslogs log driver options. ok is false when the container has no
// readable cloudwatch logs.
func ResolveLogLocation(family string, c types.ContainerDefinition, taskId string) (group string, stream string, ok bool) {
	return resolveLogLocation(configuredLogOverrides(), family, c, taskId)
}

func resolveLogLocation(overrides []CloudwatchLogOverride, family string, c types.ContainerDefinition, taskId string) (group string, stream string, ok bool) {
	containerName := ""
	if c.Name != nil {
		containerName = *c.Name
	}

	for _, l := range overrides {
		if !globMatch(l.Family, family) || !globMatch(l.Container, containerName) {
			continue
		}
		// a matching entry without a group keeps the log driver defaults, it
		// makes it possible to exclude a container from a broader entry
		if l.Group == "" {
			break
		}
		streamTemplate := l.Stream
		if streamTemplate == "" {
			streamTemplate = defaultStreamTemplate
		}
		replacer := strings.NewReplacer("{family}", family, "{container}", containerName, "{taskId}", taskId)
		return l.Group, replacer.Replace(streamTemplate), true
	}

	if c.LogConfiguration == nil {
		return "", "", false
	}
	if c.LogConfiguration.LogDriver != types.LogDriverAwslogs {
		return "", "", false
	}
	group = c.LogConfiguration.Options["awslogs-group"]
	if group == "" {
		return "", "", false
	}

	streamPrefix := containerName
	if prefix, exist := c.LogConfiguration.Options["awslogs-stream-prefix"]; exist {
		streamPrefix = prefix
	}
	return group, strings.Join([]string{streamPrefix, containerName, taskId}, "/"), true
}

// globMatch reports whether value matches pattern, an empty pattern matches
// everything.
func globMatch(pattern, value string) bool {
	if pattern == "" {
		return true
	}
	matched, err := path.Match(pattern, value)
	if err != nil {
		slog.Warn("invalid cloudwatch-log-overrides pattern", "pattern", pattern, "error", err)
		return false
	}
	return matched
}
