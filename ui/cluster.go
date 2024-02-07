package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

type ClusterView struct {
	View
	clusters []types.Cluster
}

func newClusterView(clusters []types.Cluster, app *App) *ClusterView {
	return &ClusterView{
		View: *newView(app, ClusterPage, basicKeyInputs, secondaryPageKeyMap{
			JsonPage: jsonPageKeys,
		}),
		clusters: clusters,
	}
}

func (app *App) showClustersPage(reload bool) error {
	pageName := ContainerPage.getAppPageName("")
	if app.Pages.HasPage(pageName) && app.StaleData && !reload {
		app.Pages.SwitchToPage(pageName)
		return nil
	}

	clusters, err := app.Store.ListClusters()
	if err != nil {
		logger.Printf("e1s - show clusters failed, error: %v\n", err)
		return err
	}

	view := newClusterView(clusters, app)

	if len(clusters) == 0 {
		fmt.Printf("There is no valid clusters in \033[31m%s\033[0m. Please check you ecs cluster via `AWS_REGION=%s aws ecs list-clusters`.\n", app.Region, app.Region)
		os.Exit(0)
	}

	page := buildAppPage(view)
	view.addAppPage(page)
	return nil
}

// Build info pages for cluster page
func (v *ClusterView) infoBuilder() *tview.Pages {
	for _, c := range v.clusters {
		title := *c.ClusterName
		entityName := *c.ClusterArn
		items := v.infoPagesParam(c)

		v.buildInfoPages(items, title, entityName)
	}
	// prevent empty clusters
	if len(v.clusters) > 0 && v.clusters[0].ClusterArn != nil {
		// show first when enter
		v.infoPages.SwitchToPage(*v.clusters[0].ClusterArn)
		v.changeSelectedValues()
	}
	return v.infoPages
}

// Build table for cluster page
func (v *ClusterView) tableBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.tablePages
}

// Build footer for cluster page
func (v *ClusterView) footerBuilder() *tview.Flex {
	v.footer.cluster.SetText(fmt.Sprintf(footerSelectedItemFmt, v.kind))
	v.addFooterItems()
	return v.footer.footer
}

// Handlers for cluster table
func (v *ClusterView) tableHandler() {
	for row, cluster := range v.clusters {
		c := cluster
		v.table.GetCell(row+1, 0).SetReference(Entity{cluster: &c, entityName: *c.ClusterArn})
	}
}

// Generate info pages params
func (v *ClusterView) infoPagesParam(c types.Cluster) (items []InfoItem) {
	containerInsights := "disabled"
	if len(c.Settings) > 0 && c.Settings[0].Name == "containerInsights" {
		containerInsights = *c.Settings[0].Value
	}
	// ServiceConnectDefaults
	scd := util.EmptyText
	if c.ServiceConnectDefaults != nil {
		scd = *c.ServiceConnectDefaults.Namespace
	}
	active, draining, running, pending := 0, 0, 0, 0
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
	}
	items = []InfoItem{
		{name: "Name", value: *c.ClusterName},
		{name: "Active services count", value: strconv.Itoa(active)},
		{name: "Draining services count", value: strconv.Itoa(draining)},
		{name: "Running tasks count", value: strconv.Itoa(running)},
		{name: "Pending tasks count", value: strconv.Itoa(pending)},
		{name: "Capacity providers", value: util.ShowArray(c.CapacityProviders)},
		{name: "Container insights", value: containerInsights},
		{name: "Service connect defaults", value: scd},
		{name: "Attachments status", value: util.ShowString(c.AttachmentsStatus)},
		{name: "Registered containers", value: util.ShowInt(&c.RegisteredContainerInstancesCount)},
	}
	return
}

// Generate table params
func (v *ClusterView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.kind, "all", len(v.clusters))
	headers = []string{
		"Name",
		"Status",
		"Services",
		"Tasks ▾",
		"Capacity providers",
		"Registered containers",
	}
	dataBuilder = func() (data [][]string) {
		for _, c := range v.clusters {
			// calculate tasks
			tasks := fmt.Sprintf(clusterTasksFmt, c.PendingTasksCount, c.RunningTasksCount)

			row := []string{}
			row = append(row, util.ShowString(c.ClusterName))
			row = append(row, util.ShowGreenGrey(c.Status, "active"))
			row = append(row, util.ShowInt(&c.ActiveServicesCount))
			row = append(row, tasks)
			row = append(row, util.ShowArray(c.CapacityProviders))
			row = append(row, util.ShowInt(&c.RegisteredContainerInstancesCount))

			data = append(data, row)
		}
		return data
	}

	return
}
