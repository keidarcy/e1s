package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type clusterView struct {
	view
	clusters []types.Cluster
}

func newClusterView(clusters []types.Cluster, app *App) *clusterView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["n"],
		hotKeyMap["N"],
	}...)
	return &clusterView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		clusters: clusters,
	}
}

func (app *App) showClustersPage(reload bool) error {
	app.kind = ClusterKind
	if switched := app.switchPage(reload); switched {
		return nil
	}

	resources, err := app.Store.ListClusters()
	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		return newClusterView(resources, app)
	})
	return err
}

func (v *clusterView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.cluster
}

func (v *clusterView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.clusters))
	for i, c := range v.clusters {
		params = append(params, headerPageParam{
			title:      *c.ClusterName,
			entityName: *c.ClusterArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *clusterView) headerPageItems(index int) (items []headerItem) {
	c := v.clusters[index]
	containerInsights := "disabled"
	if len(c.Settings) > 0 && c.Settings[0].Name == "containerInsights" {
		containerInsights = *c.Settings[0].Value
	}
	// ServiceConnectDefaults
	scd := utils.EmptyText
	if c.ServiceConnectDefaults != nil {
		scd = *c.ServiceConnectDefaults.Namespace
	}

	// executeCommandConfiguration
	ecc := utils.EmptyText
	if c.Configuration != nil && c.Configuration.ExecuteCommandConfiguration != nil {
		ecc = "Exists"
	}
	// managedStorageConfiguration
	msc := utils.EmptyText
	if c.Configuration != nil && c.Configuration.ManagedStorageConfiguration != nil {
		msc = "Exists"
	}
	active, draining, running, pending, activeEC2, activeFargate := 0, 0, 0, 0, 0, 0
	for _, statistic := range c.Statistics {
		n, err := strconv.Atoi(*statistic.Value)
		if err != nil {
			n = 0
		}
		if strings.HasPrefix(*statistic.Name, "active") {
			active += n
		}
		if strings.HasPrefix(*statistic.Name, "draining") {
			draining += n
		}
		if strings.HasPrefix(*statistic.Name, "running") {
			running += n
		}
		if strings.HasPrefix(*statistic.Name, "pending") {
			pending += n
		}
		if strings.HasPrefix(*statistic.Name, "activeEC2") {
			activeEC2 += n
		}
		if strings.HasPrefix(*statistic.Name, "activeFargate") {
			activeFargate += n
		}
	}
	items = []headerItem{
		{name: "Name", value: *c.ClusterName},
		{name: "Active services count", value: strconv.Itoa(active)},
		{name: "Active EC2 services count", value: strconv.Itoa(activeEC2)},
		{name: "Active Fargate services count", value: strconv.Itoa(activeFargate)},
		{name: "Draining services count", value: strconv.Itoa(draining)},
		{name: "Running tasks count", value: strconv.Itoa(running)},
		{name: "Pending tasks count", value: strconv.Itoa(pending)},
		{name: "Capacity providers", value: utils.ShowArray(c.CapacityProviders)},
		{name: "Capacity providers count", value: strconv.Itoa(len(c.CapacityProviders))},
		{name: "Container insights", value: containerInsights},
		{name: "Service connect defaults", value: scd},
		{name: "Attachments status", value: utils.ShowString(c.AttachmentsStatus)},
		{name: "Registered container instances", value: utils.ShowInt(&c.RegisteredContainerInstancesCount)},
		{name: "Execute command configuration", value: ecc},
		{name: "Managed storage configuration", value: msc},
		{name: "Tags count", value: strconv.Itoa(len(c.Tags))},
	}
	return
}

// Generate table params
func (v *clusterView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, "all", len(v.clusters))
	headers = []string{
		"Name",
		"Status",
		"Services",
		"Tasks",
		"Container instances",
		"Capacity providers",
	}

	rowsBuilder = func() (data [][]string) {
		for _, c := range v.clusters {
			// calculate tasks
			tasks := fmt.Sprintf(color.TableClusterTasksFmt, c.PendingTasksCount, c.RunningTasksCount)

			row := []string{}
			row = append(row, utils.ShowString(c.ClusterName))
			row = append(row, utils.ShowGreenGrey(c.Status, "active"))
			row = append(row, utils.ShowInt(&c.ActiveServicesCount))
			row = append(row, tasks)
			row = append(row, utils.ShowInt(&c.RegisteredContainerInstancesCount)+" EC2")
			row = append(row, utils.ShowArray(c.CapacityProviders))

			data = append(data, row)

			entity := Entity{cluster: &c, entityName: *c.ClusterArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}
	return
}
