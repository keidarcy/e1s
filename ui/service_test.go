package ui

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var (
	serviceName1 = "service1"
	serviceName2 = "service2"
)

func getServiceViews() []ServiceView {
	now := time.Now()
	service1 := types.Service{
		CreatedAt: &now,
	}
	service1.ServiceName = aws.String(serviceName1)

	service2 := types.Service{
		CreatedAt: &now,
	}
	service2.ServiceName = aws.String(serviceName2)

	app, _ := newApp(false)
	app.cluster = &types.Cluster{
		ClusterName: aws.String(clusterName1),
	}
	serviceView1 := newServiceView([]types.Service{service1}, app)
	serviceView2 := newServiceView([]types.Service{service2}, app)

	return []ServiceView{*serviceView1, *serviceView2}
}

func TestServicePageParams(t *testing.T) {
	serviceViews := getServiceViews()

	testCases := []struct {
		name string
		view ServiceView
		want want
	}{
		{
			name: serviceName1,
			view: serviceViews[0],
			want: want{
				serviceName: serviceName1,
			},
		},
		{
			name: clusterName2,
			view: serviceViews[1],
			want: want{
				serviceName: serviceName2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, c := range tc.view.services {
				items := tc.view.infoPagesParam(c)
				if items[0].value != tc.want.serviceName {
					t.Errorf("%s Got: %s, Want: %s\n", items[0].name, items[0].value, tc.want.serviceName)
				}
			}
		})
	}
}

func TestServiceTableParam(t *testing.T) {
	serviceViews := getServiceViews()

	testCases := []struct {
		name string
		view ServiceView
		want want
	}{
		{
			name: serviceName1,
			view: serviceViews[0],
			want: want{
				title:       serviceName1,
				serviceName: serviceName1,
			},
		},
		{
			name: clusterName2,
			view: serviceViews[1],
			want: want{
				title:       serviceName2,
				serviceName: serviceName2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, dataBuilder := tc.view.tableParam()
			matrix := dataBuilder()
			if matrix[0][0] != tc.want.serviceName {
				t.Errorf("Name Got: %s, Want: %s\n", matrix[0][0], tc.want.serviceName)
			}
		})
	}
}
