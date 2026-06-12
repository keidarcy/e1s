package view

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
)

var (
	daemonArn1 = "arn:aws:ecs:us-east-1:111111:daemon/cluster1/my-daemon-1"
	daemonArn2 = "arn:aws:ecs:us-east-1:111111:daemon/cluster1/my-daemon-2"
)

func getDaemonViews() []daemonView {
	now := time.Now()

	daemon1 := types.DaemonSummary{
		DaemonArn: aws.String(daemonArn1),
		Status:    types.DaemonStatusActive,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	daemon2 := types.DaemonSummary{
		DaemonArn: aws.String(daemonArn2),
		Status:    types.DaemonStatusDeleteInProgress,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	app, _ := newApp(Option{})
	app.cluster = &types.Cluster{
		ClusterName: aws.String(clusterName1),
		ClusterArn:  aws.String(clusterArn1),
	}

	v1 := newDaemonView([]types.DaemonSummary{daemon1}, app)
	v2 := newDaemonView([]types.DaemonSummary{daemon2}, app)

	return []daemonView{*v1, *v2}
}

func TestDaemonHeaderPageItems(t *testing.T) {
	views := getDaemonViews()

	testCases := []struct {
		name       string
		view       daemonView
		wantDaemon string
		wantStatus string
	}{
		{
			name:       "daemon1",
			view:       views[0],
			wantDaemon: utils.ArnToName(&daemonArn1),
			wantStatus: "ACTIVE",
		},
		{
			name:       "daemon2",
			view:       views[1],
			wantDaemon: utils.ArnToName(&daemonArn2),
			wantStatus: "DELETE_IN_PROGRESS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items := tc.view.headerPageItems(0)
			if items[0].value != tc.wantDaemon {
				t.Errorf("Daemon Got: %s, Want: %s", items[0].value, tc.wantDaemon)
			}
			if items[1].value != tc.wantStatus {
				t.Errorf("Status Got: %s, Want: %s", items[1].value, tc.wantStatus)
			}
		})
	}
}

func TestDaemonTableParamsBuilder(t *testing.T) {
	views := getDaemonViews()

	testCases := []struct {
		name       string
		view       daemonView
		wantDaemon string
	}{
		{
			name:       "daemon1",
			view:       views[0],
			wantDaemon: utils.ArnToName(&daemonArn1),
		},
		{
			name:       "daemon2",
			view:       views[1],
			wantDaemon: utils.ArnToName(&daemonArn2),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, headers, rowsBuilder := tc.view.tableParamsBuilder()
			if headers[0] != "Daemon" {
				t.Errorf("Header[0] Got: %s, Want: Daemon", headers[0])
			}
			matrix := rowsBuilder()
			if matrix[0][0] != tc.wantDaemon {
				t.Errorf("Name Got: %s, Want: %s", matrix[0][0], tc.wantDaemon)
			}
		})
	}
}
