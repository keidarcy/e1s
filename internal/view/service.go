package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type serviceView struct {
	view
	services []types.Service
}

func newServiceView(services []types.Service, app *App) *serviceView {
	keys := append(basicKeyInputs, []keyInput{
		hotKeyMap["U"],
		hotKeyMap["w"],
		hotKeyMap["t"],
		hotKeyMap["l"],
		hotKeyMap["m"],
		hotKeyMap["a"],
	}...)
	return &serviceView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind:   describePageKeys,
			LogKind:           logPageKeys,
			AutoScalingKind:   describePageKeys,
			ServiceEventsKind: otherDescribePageKeys,
		}),
		services: services,
	}
}

func (app *App) showServicesPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	services, err := app.Store.ListServices(app.cluster.ClusterName)
	if err != nil {
		logger.Warnf("Failed to show services page, error: %v", err)
		app.back()
		return err
	}

	// no services exists do nothing
	if len(services) == 0 {
		app.back()
		return fmt.Errorf("no valid service")
	}

	view := newServiceView(services, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for service page
func (v *serviceView) headerBuilder() *tview.Pages {
	for _, s := range v.services {
		title := *s.ServiceName
		entityName := *s.ServiceArn
		items := v.headerPagesParam(s)

		v.buildHeaderPages(items, title, entityName)
	}
	// prevent empty services
	if len(v.services) > 0 && v.services[0].ServiceArn != nil {
		// show first when enter
		v.headerPages.SwitchToPage(*v.services[0].ServiceArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Build table for service page
func (v *serviceView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.bodyPages
}

// Build footer for service page
func (v *serviceView) footerBuilder() *tview.Flex {
	v.footer.service.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Handlers for service table
func (v *serviceView) tableHandler() {
	for row, service := range v.services {
		s := service
		// Events are too long show in separate view
		events := s.Events
		s.Events = []types.ServiceEvent{}
		v.table.GetCell(row+1, 0).SetReference(Entity{service: &s, events: events, entityName: *s.ServiceArn})
	}
}

// Generate info pages params
func (v *serviceView) headerPagesParam(s types.Service) (items []headerItem) {
	// publicIP
	ip := utils.EmptyText
	// security groups
	sgs := []string{}
	if s.NetworkConfiguration != nil && s.NetworkConfiguration.AwsvpcConfiguration != nil {
		sgs = append(sgs, s.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups...)
		ip = string(s.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp)
	}

	// target groups
	tgs := []string{}
	for _, lb := range s.LoadBalancers {
		tgs = append(tgs, utils.ArnToFullName(lb.TargetGroupArn))
	}

	// capacity provider strategy
	cps := []string{}
	for _, p := range s.CapacityProviderStrategy {
		cps = append(cps, *p.CapacityProvider)
	}
	cpsString := strings.Join(cps, ",")
	if len(cpsString) == 0 {
		cpsString = utils.EmptyText
	}

	// deployment circuit breaker enable
	dcbe := utils.EmptyText
	// deployment circuit breaker rollback
	dcbr := utils.EmptyText
	// deployment max percent
	dmaxp := utils.EmptyText
	// deployment min percent
	dminp := utils.EmptyText

	if s.DeploymentConfiguration != nil {
		dmaxp = strconv.Itoa(int(*s.DeploymentConfiguration.MaximumPercent)) + "%"
		dminp = strconv.Itoa(int(*s.DeploymentConfiguration.MinimumHealthyPercent)) + "%"
		if s.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			dcbe = strconv.FormatBool(s.DeploymentConfiguration.DeploymentCircuitBreaker.Enable)
			dcbr = strconv.FormatBool(s.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback)
		}
	}

	// deployment controller
	dc := utils.EmptyText
	if s.DeploymentController != nil {
		dc = string(s.DeploymentController.Type)
	}

	items = []headerItem{
		{name: "Name", value: utils.ShowString(s.ServiceName)},
		{name: "Cluster", value: utils.ArnToName(s.ClusterArn)},
		{name: "Capacity provider strategy", value: cpsString},
		{name: "RoleArn", value: utils.ArnToName(s.RoleArn)},
		{name: "Task Definition", value: utils.ArnToName(s.TaskDefinition)},
		{name: "Scheduling strategy", value: string(s.SchedulingStrategy)},
		{name: "Deployment controller", value: dc},
		{name: "Deployment circuitBreaker enable", value: dcbe},
		{name: "Deployment circuitBreaker rollback", value: dcbr},
		{name: "Deployment maximum", value: dmaxp},
		{name: "Deployment minimum", value: dminp},
		{name: "Public IP", value: ip},
		{name: "Security groups", value: strings.Join(sgs, ",")},
		{name: "Target groups", value: strings.Join(tgs, ",")},
		{name: "Execute command", value: strconv.FormatBool(s.EnableExecuteCommand)},
		{name: "Created", value: utils.ShowTime(s.CreatedAt)},
		{name: "Created by", value: utils.ArnToName(s.CreatedBy)},
		{name: "Platform family", value: utils.ShowString(s.PlatformFamily)},
		{name: "Platform version", value: utils.ShowString(s.PlatformVersion)},
	}
	return
}

// Generate table params
func (v *serviceView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, "Services", *v.app.cluster.ClusterName, len(v.services))
	headers = []string{
		"Name",
		"Status",
		"Tasks ▾",
		"Pending",
		"Last deployment",
		"Execute command",
		"Task definition",
		"Age",
	}
	dataBuilder = func() (data [][]string) {
		for _, s := range v.services {
			row := []string{}

			// tasks
			tasks := fmt.Sprintf(serviceTasksFmt, s.RunningCount, s.DesiredCount)

			// last deployment
			lastDeployment := ""
			// last update time
			var lastUpdateTime *time.Time

			if len(s.Deployments) > 0 {
				rollout := string(s.Deployments[0].RolloutState)
				lastDeployment += fmt.Sprintf("%s - %s", utils.ShowGreenGrey(&rollout, "completed"), s.Deployments[0].UpdatedAt.Format(time.RFC3339))

				lastUpdateTime = s.Deployments[0].UpdatedAt
			}

			// enable execute command
			enableExecuteCommand := "False"
			if s.EnableExecuteCommand {
				enableExecuteCommand = "True"
			}

			row = append(row, utils.ShowString(s.ServiceName))
			row = append(row, utils.ShowGreenGrey(s.Status, "active"))
			row = append(row, tasks)
			row = append(row, utils.ShowInt(&s.PendingCount))
			row = append(row, lastDeployment)
			row = append(row, utils.ShowGreenGrey(&enableExecuteCommand, "true"))
			row = append(row, utils.ArnToName(s.TaskDefinition))
			row = append(row, utils.Age(lastUpdateTime))
			data = append(data, row)
		}
		return data
	}
	return
}
