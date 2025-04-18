package view

import (
	"fmt"
	"log/slog"
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

	clusters, err := app.Store.ListClusters()
	if err != nil {
		slog.Error("failed to load clusters", "region", app.Region, "error", err.Error())
		return err
	}

	if len(clusters) == 0 {
		m := fmt.Sprintf("there is no valid clusters in %s region", app.Region)
		slog.Warn("failed start", "reason", m)
		return fmt.Errorf(m)
	}

	view := newClusterView(clusters, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for cluster page
func (v *clusterView) headerBuilder() *tview.Pages {
	for _, c := range v.clusters {
		title := *c.ClusterName
		entityName := *c.ClusterArn
		items := v.headerPagesParam(c)

		v.buildHeaderPages(items, title, entityName)
	}
	// prevent empty clusters
	if len(v.clusters) > 0 && v.clusters[0].ClusterArn != nil {
		// show first when enter
		v.headerPages.SwitchToPage(*v.clusters[0].ClusterArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Build table for cluster page
func (v *clusterView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.bodyPages
}

// Build footer for cluster page
func (v *clusterView) footerBuilder() *tview.Flex {
	v.footer.cluster.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Handlers for cluster table
func (v *clusterView) tableHandler() {
	for row, cluster := range v.clusters {
		c := cluster
		v.table.GetCell(row+1, 0).SetReference(Entity{cluster: &c, entityName: *c.ClusterArn})
	}
}

// Generate info pages params
func (v *clusterView) headerPagesParam(c types.Cluster) (items []headerItem) {
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
		v, err := strconv.Atoi(*statistic.Value)
		if err != nil {
			v = 0
		}
		if strings.HasPrefix(*statistic.Name, "active") {
			active += v
		}
		if strings.HasPrefix(*statistic.Name, "draining") {
			draining += v
		}
		if strings.HasPrefix(*statistic.Name, "running") {
			running += v
		}
		if strings.HasPrefix(*statistic.Name, "pending") {
			pending += v
		}
		if strings.HasPrefix(*statistic.Name, "activeEC2") {
			activeEC2 += v
		}
		if strings.HasPrefix(*statistic.Name, "activeFargate") {
			activeFargate += v
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
func (v *clusterView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, "all", len(v.clusters))
	headers = []string{
		"Name",
		"Status",
		"Services",
		"Tasks ▾",
		"Container instances",
		"Capacity providers",
	}
	dataBuilder = func() (data [][]string) {
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
		}
		return data
	}

	return
}
