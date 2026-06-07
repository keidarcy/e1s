package view

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/rivo/tview"
)

type testResourceViewBuilder struct {
	v            *view
	footer       *tview.TextView
	headerParams []headerPageParam
	title        string
	headers      []string
	rows         [][]string
}

func (b *testResourceViewBuilder) getViewAndFooter() (*view, *tview.TextView) {
	return b.v, b.footer
}

func (b *testResourceViewBuilder) headerParamsBuilder() []headerPageParam {
	return b.headerParams
}

func (b *testResourceViewBuilder) headerPageItems(index int) []headerItem {
	_ = index
	return nil
}

func (b *testResourceViewBuilder) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	return b.title, b.headers, func() [][]string {
		return b.rows
	}
}

func TestResourceViewPreHandler(t *testing.T) {
	resources := []string{"resource"}

	t.Run("error input", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = TaskKind
		expectErr := errors.New("failed to load resources")

		gotErr := resourceViewPreHandler(resources, app, expectErr)
		if !errors.Is(gotErr, expectErr) {
			t.Errorf("Got: %v, Want: %v\n", gotErr, expectErr)
		}
		if app.kind != ServiceKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ServiceKind)
		}
	})

	t.Run("empty resources", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = ServiceKind

		gotErr := resourceViewPreHandler([]string{}, app, nil)
		if gotErr == nil {
			t.Errorf("Got nil error, Want non nil error\n")
		}
		if gotErr != nil && gotErr.Error() != "no services found" {
			t.Errorf("Got: %v, Want: %v\n", gotErr, "no services found")
		}
		if app.kind != ClusterKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ClusterKind)
		}
	})

	t.Run("success", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = ServiceKind

		gotErr := resourceViewPreHandler(resources, app, nil)
		if gotErr != nil {
			t.Errorf("Got: %v, Want: %v\n", gotErr, nil)
		}
		if app.kind != ServiceKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ServiceKind)
		}
	})
}

func TestBuildResourcePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = ClusterKind

		cluster := &types.Cluster{
			ClusterName: aws.String(clusterName1),
			ClusterArn:  aws.String(clusterArn1),
		}

		v := newView(app, basicKeyInputs, nil)
		v.originalRowReferences = []Entity{
			{
				cluster:    cluster,
				entityName: clusterArn1,
			},
		}
		footer := tview.NewTextView().SetDynamicColors(true)

		builder := &testResourceViewBuilder{
			v:      v,
			footer: footer,
			headerParams: []headerPageParam{
				{
					title:      clusterName1,
					entityName: clusterArn1,
					items: []headerItem{
						{name: "Name", value: clusterName1},
					},
				},
			},
			title:   clusterName1,
			headers: []string{"Name"},
			rows: [][]string{
				{clusterName1},
			},
		}

		err := buildResourcePage([]types.Cluster{*cluster}, app, nil, func() resourceViewBuilder {
			return builder
		})
		if err != nil {
			t.Errorf("Got: %v, Want: %v\n", err, nil)
		}
		if !app.Pages.HasPage(ClusterKind.getAppPageName("")) {
			t.Errorf("Page clusters should exist\n")
		}
		if v.mainFlex == nil {
			t.Errorf("mainFlex should be initialized\n")
		}
		if v.filterInput == nil {
			t.Errorf("filterInput should be initialized\n")
		}
		if v.table.GetRowCount() != 2 {
			t.Errorf("RowCount Got: %d, Want: %d\n", v.table.GetRowCount(), 2)
		}
		if app.cluster == nil || aws.ToString(app.cluster.ClusterArn) != clusterArn1 {
			t.Errorf("ClusterArn Got: %v, Want: %v\n", app.cluster, clusterArn1)
		}
		footerText := footer.GetText(false)
		if !strings.Contains(footerText, app.kind.String()) {
			t.Errorf("Footer Got: %s, Want contains: %s\n", footerText, app.kind.String())
		}
	})

	t.Run("pre handler error", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = TaskKind

		expectErr := errors.New("failed before builder")
		called := false
		err := buildResourcePage([]string{"resource"}, app, expectErr, func() resourceViewBuilder {
			called = true
			return nil
		})

		if !errors.Is(err, expectErr) {
			t.Errorf("Got: %v, Want: %v\n", err, expectErr)
		}
		if called {
			t.Errorf("builder should not be called when pre handler returns error\n")
		}
		if app.kind != ServiceKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ServiceKind)
		}
	})
}

func TestViewStateKeyIncludesPageHandle(t *testing.T) {
	app, _ := newApp(Option{})
	app.kind = ServiceKind
	app.cluster.ClusterArn = aws.String(clusterArn1)
	cluster1Key := app.viewStateKey()

	app.cluster.ClusterArn = aws.String(clusterArn2)
	cluster2Key := app.viewStateKey()

	if cluster1Key == cluster2Key {
		t.Errorf("view state keys should differ between clusters, got %q", cluster1Key)
	}
}

func TestCanAutoRefreshSkipsWhileFilterInputActive(t *testing.T) {
	app, _ := newApp(Option{})
	if !app.canAutoRefresh() {
		t.Errorf("canAutoRefresh should allow refresh by default")
	}

	app.filterInputActive = true
	if app.canAutoRefresh() {
		t.Errorf("canAutoRefresh should skip while filter input is active")
	}

	app.filterInputActive = false
	app.isSuspended = true
	if app.canAutoRefresh() {
		t.Errorf("canAutoRefresh should skip while app is suspended")
	}

	app.isSuspended = false
	app.secondaryKind = DescriptionKind
	if app.canAutoRefresh() {
		t.Errorf("canAutoRefresh should skip while a secondary view is active")
	}
}

func TestShowAndHideFilterInputTogglesAutoRefreshGuard(t *testing.T) {
	app, _ := newApp(Option{})
	v := newView(app, basicKeyInputs, nil)
	v.initFilterInput()

	err := v.showFilterInput()
	if err != nil {
		t.Errorf("Got: %v, Want: %v\n", err, nil)
	}
	if !app.filterInputActive {
		t.Errorf("filterInputActive should be true after showing the filter input")
	}
	if app.canAutoRefresh() {
		t.Errorf("canAutoRefresh should skip while the filter input is shown")
	}

	v.hideFilterInput()
	if app.filterInputActive {
		t.Errorf("filterInputActive should be false after hiding the filter input")
	}
}

func TestFilterInputChangeSavesViewStateBeforeApply(t *testing.T) {
	app, _ := newApp(Option{})
	app.kind = ClusterKind
	v := newView(app, basicKeyInputs, nil)
	v.initFilterInput()

	v.filterInput.SetText("bravo")

	state, ok := app.viewStates[app.viewStateKey()]
	if !ok {
		t.Fatalf("view state should be saved when filter text changes")
	}
	if state.filterText != "bravo" {
		t.Errorf("filterText Got: %q, Want: %q", state.filterText, "bravo")
	}
	if state.sortColumn != -1 {
		t.Errorf("sortColumn Got: %d, Want: %d", state.sortColumn, -1)
	}
}

func TestApplyFilterSavesNoSortState(t *testing.T) {
	app, _ := newApp(Option{})
	app.kind = ClusterKind
	v := newView(app, basicKeyInputs, nil)
	v.headers = []string{"Name"}
	v.originalRowData = [][]string{
		{"alpha"},
		{"bravo"},
	}
	v.originalRowReferences = []Entity{
		{entityName: "alpha"},
		{entityName: "bravo"},
	}
	v.initFilterInput()
	v.filterInput.SetText("a")

	v.applyFilter()

	state, ok := app.viewStates[app.viewStateKey()]
	if !ok {
		t.Fatalf("view state should be saved")
	}
	if state.filterText != "a" {
		t.Errorf("filterText Got: %q, Want: %q", state.filterText, "a")
	}
	if state.sortColumn != -1 {
		t.Errorf("sortColumn Got: %d, Want: %d", state.sortColumn, -1)
	}
	if state.sortOrder != "desc" {
		t.Errorf("sortOrder Got: %q, Want: %q", state.sortOrder, "desc")
	}
}

func TestBuildResourcePageRestoresFilterOnlyStateWithoutSorting(t *testing.T) {
	app, _ := newApp(Option{})
	app.kind = ClusterKind
	app.viewStates[app.viewStateKey()] = viewState{
		sortColumn: -1,
		sortOrder:  "desc",
		filterText: "br",
	}

	v := newView(app, basicKeyInputs, nil)
	v.originalRowReferences = []Entity{
		{entityName: "alpha"},
		{entityName: "bravo"},
		{entityName: "charlie"},
	}
	footer := tview.NewTextView().SetDynamicColors(true)
	builder := &testResourceViewBuilder{
		v:      v,
		footer: footer,
		title:  "clusters",
		headers: []string{
			"Name",
		},
		rows: [][]string{
			{"alpha"},
			{"bravo"},
			{"charlie"},
		},
	}

	err := buildResourcePage([]string{"resource"}, app, nil, func() resourceViewBuilder {
		return builder
	})
	if err != nil {
		t.Errorf("Got: %v, Want: %v\n", err, nil)
	}
	if v.filterInput.GetText() != "br" {
		t.Errorf("filter input Got: %q, Want: %q", v.filterInput.GetText(), "br")
	}
	if v.table.GetRowCount() != 2 {
		t.Errorf("RowCount Got: %d, Want: %d", v.table.GetRowCount(), 2)
	}
	firstRow := v.table.GetCell(1, 0).Text
	if firstRow != "bravo" {
		t.Errorf("first row Got: %q, Want filtered row %q", firstRow, "bravo")
	}
	if v.sortColumn != -1 {
		t.Errorf("sortColumn Got: %d, Want: %d", v.sortColumn, -1)
	}
}
