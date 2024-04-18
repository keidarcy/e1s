package view

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var (
	clusterName1 = "cluster1"

	clusterName2 = "cluster2"
)

type want struct {
	title         string
	clusterName   string
	serviceName   string
	taskID        string
	containerName string
}

func getClusterViews() []clusterView {
	cluster1 := types.Cluster{}
	cluster1.ClusterName = aws.String(clusterName1)

	cluster2 := types.Cluster{}
	cluster2.ClusterName = aws.String(clusterName2)

	app, _ := newApp(Option{})
	clusterView1 := newClusterView([]types.Cluster{cluster1}, app)
	clusterView2 := newClusterView([]types.Cluster{cluster2}, app)

	return []clusterView{*clusterView1, *clusterView2}
}

func TestClusterPageParams(t *testing.T) {
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
			},
		},
		{
			name: clusterName2,
			view: clusterViews[1],
			want: want{
				title:       clusterName2,
				clusterName: clusterName2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, c := range tc.view.clusters {
				items := tc.view.headerPagesParam(c)
				if items[0].value != tc.want.clusterName {
					t.Errorf("%s Got: %s, Want: %s\n", items[0].name, items[0].value, tc.want.clusterName)
				}
			}
		})
	}
}

func TestClusterTableParam(t *testing.T) {
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
			},
		},
		{
			name: clusterName2,
			view: clusterViews[1],
			want: want{
				title:       clusterName2,
				clusterName: clusterName2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, dataBuilder := tc.view.tableParam()
			matrix := dataBuilder()
			if matrix[0][0] != tc.want.clusterName {
				t.Errorf("Name Got: %s, Want: %s\n", matrix[0][0], tc.want.clusterName)
			}
		})
	}
}
