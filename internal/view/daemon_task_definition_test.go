package view

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var (
	daemonTdArn1 = "arn:aws:ecs:us-east-1:111111:daemon-task-definition/my-daemon:1"
	daemonTdArn2 = "arn:aws:ecs:us-east-1:111111:daemon-task-definition/my-daemon:2"
)

func getDaemonTaskDefinitionViews() []daemonTaskDefinitionView {
	now := time.Now()

	td1 := types.DaemonTaskDefinition{
		DaemonTaskDefinitionArn: aws.String(daemonTdArn1),
		Family:                  aws.String("my-daemon"),
		Revision:                1,
		Status:                  types.DaemonTaskDefinitionStatusActive,
		Cpu:                     aws.String("256"),
		Memory:                  aws.String("512"),
		RegisteredAt:            &now,
		ContainerDefinitions: []types.DaemonContainerDefinition{
			{Name: aws.String("agent")},
		},
	}
	td2 := types.DaemonTaskDefinition{
		DaemonTaskDefinitionArn: aws.String(daemonTdArn2),
		Family:                  aws.String("my-daemon"),
		Revision:                2,
		Status:                  types.DaemonTaskDefinitionStatusActive,
		Cpu:                     aws.String("512"),
		Memory:                  aws.String("1024"),
		RegisteredAt:            &now,
		ContainerDefinitions: []types.DaemonContainerDefinition{
			{Name: aws.String("agent")},
			{Name: aws.String("sidecar")},
		},
	}

	app, _ := newApp(Option{})
	app.cluster = &types.Cluster{
		ClusterName: aws.String(clusterName1),
		ClusterArn:  aws.String(clusterArn1),
	}
	app.daemonSummary = &types.DaemonSummary{
		DaemonArn: aws.String(daemonArn1),
	}

	v1 := newDaemonTaskDefinitionView([]types.DaemonTaskDefinition{td1, td2}, app)
	return []daemonTaskDefinitionView{*v1}
}

func TestDaemonTaskDefinitionHeaderPageItems(t *testing.T) {
	views := getDaemonTaskDefinitionViews()
	v := views[0]

	items := v.headerPageItems(0)
	if items[0].value != "my-daemon" {
		t.Errorf("Family Got: %s, Want: my-daemon", items[0].value)
	}
	if items[1].value != "1" {
		t.Errorf("Revision Got: %s, Want: 1", items[1].value)
	}
	if items[3].value != "256" {
		t.Errorf("CPU Got: %s, Want: 256", items[3].value)
	}

	items = v.headerPageItems(1)
	if items[1].value != "2" {
		t.Errorf("Revision Got: %s, Want: 2", items[1].value)
	}
	if items[5].value != "agent,sidecar" {
		t.Errorf("Containers Got: %s, Want: agent,sidecar", items[5].value)
	}
}

func TestDaemonTaskDefinitionTableParamsBuilder(t *testing.T) {
	views := getDaemonTaskDefinitionViews()
	v := views[0]

	_, headers, rowsBuilder := v.tableParamsBuilder()
	if headers[0] != "Revision" {
		t.Errorf("Header[0] Got: %s, Want: Revision", headers[0])
	}

	matrix := rowsBuilder()
	if len(matrix) != 2 {
		t.Fatalf("Rows Got: %d, Want: 2", len(matrix))
	}
	// First row: my-daemon:1
	if matrix[0][0] != "my-daemon:1" {
		t.Errorf("Revision Got: %s, Want: my-daemon:1", matrix[0][0])
	}
	if matrix[0][2] != "256" {
		t.Errorf("CPU Got: %s, Want: 256", matrix[0][2])
	}
	// Second row: my-daemon:2
	if matrix[1][0] != "my-daemon:2" {
		t.Errorf("Revision Got: %s, Want: my-daemon:2", matrix[1][0])
	}
	if matrix[1][4] != "2" {
		t.Errorf("Containers Got: %s, Want: 2", matrix[1][4])
	}
}
