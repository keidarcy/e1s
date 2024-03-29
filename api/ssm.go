package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"golang.org/x/sync/errgroup"
)

type SsmStartSessionInput struct {
	ClusterName string
	TaskId      string
	RuntimeId   string
	RemoteHost  bool
	Host        string
	Port        string
	LocalPort   string
}

// Equivalent to
// aws ssm start-session
// --target ecs:${cluster_id}_${task_id}_${runtime_id}
// --document-name AWS-StartPortForwardingSession
// --parameters {"portNumber":["${port}"], "localPortNumber":["${local_port}"]}
// OR
// aws ssm start-session
// --target ecs:${cluster_id}_${task_id}_${runtime_id}
// --document-name AWS-StartPortForwardingSession
// --parameters {"portNumber":["${port}"], "localPortNumber":["${local_port}"]}
func (store *Store) StartSession(input *SsmStartSessionInput) (*string, error) {
	store.initSsmClient()
	smpCi := "session-manager-plugin"

	target := fmt.Sprintf("ecs:%s_%s_%s", input.ClusterName, input.TaskId, input.RuntimeId)

	documentName := "AWS-StartPortForwardingSession"
	params := map[string][]string{
		"portNumber":      {input.Port},
		"localPortNumber": {input.LocalPort},
	}
	if input.RemoteHost {
		documentName = "AWS-StartPortForwardingSessionToRemoteHost"
		params["host"] = []string{input.Host}
	}

	startInput := &ssm.StartSessionInput{
		Target:       aws.String(target),
		DocumentName: aws.String(documentName),
		Parameters:   params,
		Reason:       aws.String("session started via e1s"),
	}

	result, err := store.ssm.StartSession(context.Background(), startInput)
	if err != nil {
		return nil, err
	}

	type sessionManagerPluginParameter struct {
		Target     string
		Parameters map[string][]string
	}

	bin, err := exec.LookPath(smpCi)
	if err != nil {
		logger.Warnf("Failed to find %s path, please check %s", smpCi, "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
		return nil, fmt.Errorf("failed to find %s path, please check %s", smpCi, "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
	}

	region := store.Config.Region

	sessionJson, _ := json.Marshal(result)

	profile := "default"
	if p := os.Getenv("AWS_PROFILE"); p != "" {
		profile = p
	}
	pluginParameter := &sessionManagerPluginParameter{Target: target, Parameters: startInput.Parameters}
	parameterJson, _ := json.Marshal(pluginParameter)

	args := []string{
		string(sessionJson),
		region,
		"StartSession",
		profile,
		string(parameterJson),
		fmt.Sprintf("https://ssm.%v.amazonaws.com", region),
	}

	logger.Infof("Exec: `%s %s`", bin, strings.Join(args, " "))
	// start process
	cmd := exec.Command(bin, args...)
	err = cmd.Start()

	return result.SessionId, err
}

func (store *Store) TerminateSessions(sessionIds []*string) error {
	g := new(errgroup.Group)

	for _, id := range sessionIds {
		g.Go(func() error {
			input := &ssm.TerminateSessionInput{
				SessionId: id,
			}
			_, err := store.ssm.TerminateSession(context.Background(), input)
			return err
		})
	}
	err := g.Wait()
	return err
}
