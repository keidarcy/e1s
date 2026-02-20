package view

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
)

var (
	clusterName1     = "cluster1"
	clusterArn1      = "arn:aws:ecs:us-east-1:111111:cluster/cluster1"
	clusterServices1 = aws.Int32(10)
	clusterName2     = "cluster2"
	clusterArn2      = "arn:aws:ecs:us-east-1:111111:cluster/cluster2"
	clusterServices2 = aws.Int32(0)
)

type want struct {
	title             string
	clusterName       string
	serviceName       string
	taskID            string
	containerName     string
	containerInsights string
	services          string
}

func getClusterViews() []clusterView {
	cluster1 := types.Cluster{}
	cluster1.ClusterName = aws.String(clusterName1)
	cluster1.ClusterArn = aws.String(clusterArn1)
	cluster1.Settings = []types.ClusterSetting{
		{
			Name:  types.ClusterSettingNameContainerInsights,
			Value: aws.String("enabled"),
		},
	}
	cluster1.ActiveServicesCount = *clusterServices1

	cluster2 := types.Cluster{}
	cluster2.ClusterName = aws.String(clusterName2)
	cluster2.ClusterArn = aws.String(clusterArn2)
	cluster2.ActiveServicesCount = *clusterServices2

	app, _ := newApp(Option{})
	clusterView1 := newClusterView([]types.Cluster{cluster1}, app)
	clusterView2 := newClusterView([]types.Cluster{cluster2}, app)

	return []clusterView{*clusterView1, *clusterView2}
}

func TestClusterHeaderPageItems(t *testing.T) {
	clusterViews := getClusterViews()

	testCases := []struct {
		name string
		view clusterView
		want want
	}{
		{
			name: clusterName1,
			view: clusterViews[0],
			want: want{
				title:             clusterName1,
				clusterName:       clusterName1,
				containerInsights: "enabled",
			},
		},
		{
			name: clusterName2,
			view: clusterViews[1],
			want: want{
				title:             clusterName2,
				clusterName:       clusterName2,
				containerInsights: "disabled",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i := range tc.view.clusters {
				items := tc.view.headerPageItems(i)
				if items[0].value != tc.want.clusterName {
					t.Errorf("%s Got: %s, Want: %s\n", items[0].name, items[0].value, tc.want.clusterName)
				}
				if items[9].value != tc.want.containerInsights {
					t.Errorf("%s Got: %s, Want: %s\n", items[9].name, items[9].value, tc.want.containerInsights)
				}
			}
		})
	}
}

func TestClusterTableParamsBuilder(t *testing.T) {
	clusterViews := getClusterViews()

	testCases := []struct {
		name string
		view clusterView
		want want
	}{
		{
			name: clusterName1,
			view: clusterViews[0],
			want: want{
				title:       clusterName1,
				clusterName: clusterName1,
				services:    utils.ShowInt(clusterServices1),
			},
		},
		{
			name: clusterName2,
			view: clusterViews[1],
			want: want{
				title:       clusterName2,
				clusterName: clusterName2,
				services:    utils.ShowInt(clusterServices2),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, rowsBuilder := tc.view.tableParamsBuilder()
			matrix := rowsBuilder()
			if matrix[0][0] != tc.want.clusterName {
				t.Errorf("Name Got: %s, Want: %s\n", matrix[0][0], tc.want.clusterName)
			}
			if matrix[0][2] != tc.want.services {
				t.Errorf("Name Got: %s, Want: %s\n", matrix[0][2], tc.want.services)
			}
		})
	}
}
