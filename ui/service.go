package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

type ServiceView struct {
	View
	services []types.Service
}

func newServiceView(services []types.Service, app *App) *ServiceView {
	keys := append(basicKeyInputs, []KeyInput{
		{key: string(wKey), description: describeServiceEvents},
		{key: string(tKey), description: describeTaskDefinition},
		{key: string(vKey), description: describeTaskDefinitionRevisions},
		{key: string(mKey), description: showMetrics},
		{key: string(aKey), description: showAutoScaling},
		{key: string(eKey), description: editService},
		{key: string(oKey), description: showLogs},
	}...)
	return &ServiceView{
		View:     *newView(app, ServicePage, keys),
		services: services,
	}
}

func (app *App) showServicesPage() error {
	services, err := app.Store.ListServices(app.cluster.ClusterName)
	if err != nil {
		logger.Printf("e1s - show services page failed, error: %v\n", err)
		return err
	}

	// no services exists do nothing
	if len(services) == 0 {
		return nil
	}

	view := newServiceView(services, app)
	page := buildAppPage(view)
	view.addAppPage(page)
	return nil
}

// Build info pages for service page
func (v *ServiceView) infoBuilder() *tview.Pages {
	for _, s := range v.services {
		items := v.infoPagesParam(s)
		infoFlex := v.buildInfoFlex(*s.ServiceName, items, v.keys)
		v.infoPages.AddPage(*s.ServiceArn, infoFlex, true, true)
	}
	// prevent empty services
	if len(v.services) > 0 && v.services[0].ServiceArn != nil {
		// show first when enter
		v.infoPages.SwitchToPage(*v.services[0].ServiceArn)
		v.changeSelectedValues()
	}
	return v.infoPages
}

// Build table for service page
func (v *ServiceView) tableBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.tablePages
}

// Build footer for service page
func (v *ServiceView) footerBuilder() *tview.Flex {
	v.footer.service.SetText(fmt.Sprintf(footerSelectedItemFmt, v.kind))
	v.addFooterItems()
	return v.footer.footer
}

// Handlers for service table
func (v *ServiceView) tableHandler() {
	for row, service := range v.services {
		s := service
		// Events are too long show in separate view
		events := s.Events
		s.Events = []types.ServiceEvent{}
		v.table.GetCell(row+1, 0).SetReference(Entity{service: &s, events: events, entityName: *s.ServiceArn})
	}
}

// Generate info pages params
func (v *ServiceView) infoPagesParam(s types.Service) (items []InfoItem) {
	// publicIP
	ip := util.EmptyText
	// security groups
	sgs := []string{}
	if s.NetworkConfiguration != nil && s.NetworkConfiguration.AwsvpcConfiguration != nil {
		sgs = append(sgs, s.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups...)
		ip = string(s.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp)
	}

	// target groups
	tgs := []string{}
	for _, lb := range s.LoadBalancers {
		tgs = append(tgs, util.ArnToFullName(lb.TargetGroupArn))
	}

	// capacity provider strategy
	cps := []string{}
	for _, p := range s.CapacityProviderStrategy {
		cps = append(cps, *p.CapacityProvider)
	}
	cpsString := strings.Join(cps, ",")
	if len(cpsString) == 0 {
		cpsString = util.EmptyText
	}

	// deployment circuit breaker enable
	dcbe := util.EmptyText
	// deployment circuit breaker rollback
	dcbr := util.EmptyText
	// deployment max percent
	dmaxp := util.EmptyText
	// deployment min percent
	dminp := util.EmptyText

	if s.DeploymentConfiguration != nil {
		dmaxp = strconv.Itoa(int(*s.DeploymentConfiguration.MaximumPercent)) + "%"
		dminp = strconv.Itoa(int(*s.DeploymentConfiguration.MinimumHealthyPercent)) + "%"
		if s.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			dcbe = strconv.FormatBool(s.DeploymentConfiguration.DeploymentCircuitBreaker.Enable)
			dcbr = strconv.FormatBool(s.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback)
		}
	}

	// deployment controller
	dc := util.EmptyText
	if s.DeploymentController != nil {
		dc = string(s.DeploymentController.Type)
	}

	items = []InfoItem{
		{name: "Name", value: util.ShowString(s.ServiceName)},
		{name: "Cluster", value: util.ArnToName(s.ClusterArn)},
		{name: "Capacity provider strategy", value: cpsString},
		{name: "RoleArn", value: util.ArnToName(s.RoleArn)},
		{name: "Task Definition", value: util.ArnToName(s.TaskDefinition)},
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
		{name: "Created", value: util.ShowTime(s.CreatedAt)},
		{name: "Created by", value: util.ArnToName(s.CreatedBy)},
		{name: "Platform family", value: util.ShowString(s.PlatformFamily)},
		{name: "Platform version", value: util.ShowString(s.PlatformVersion)},
	}
	return
}

// Generate table params
func (v *ServiceView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, "Services", *v.app.cluster.ClusterName, len(v.services))
	headers = []string{
		"Name",
		"Status",
		"Tasks ▾",
		"Last deployment",
		"Task definition",
		"Revision",
	}
	dataBuilder = func() (data [][]string) {
		for _, s := range v.services {
			row := []string{}

			// tasks
			tasks := fmt.Sprintf(serviceTasksFmt, s.RunningCount, s.DesiredCount)

			// last deployment
			lastDeployment := ""
			if len(s.Deployments) > 0 {
				rollout := string(s.Deployments[0].RolloutState)
				lastDeployment += fmt.Sprintf("%s - %s", util.ShowGreenGrey(&rollout, "completed"), s.Deployments[0].UpdatedAt.Format(time.RFC3339))
			}

			// task definition family and revision
			family, revision := getTaskDefinitionInfo(s.TaskDefinition)

			row = append(row, util.ShowString(s.ServiceName))
			row = append(row, util.ShowGreenGrey(s.Status, "active"))
			row = append(row, tasks)
			row = append(row, lastDeployment)
			row = append(row, family)
			row = append(row, revision)
			data = append(data, row)
		}
		return data
	}
	return
}
