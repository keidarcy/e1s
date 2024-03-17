package ui

type Kind int

const (
	ClusterPage Kind = iota
	ServicePage
	TaskPage
	ContainerPage
	TaskDefinitionPage
	TaskDefinitionRevisionsPage
	DescriptionPage
	ServiceEventsPage
	LogPage
	AutoScalingPage
	EmptyKind
)

func (k Kind) String() string {
	switch k {
	case ClusterPage:
		return "clusters"
	case ServicePage:
		return "services"
	case TaskPage:
		return "tasks"
	case ContainerPage:
		return "containers"
	case DescriptionPage:
		return "description"
	case TaskDefinitionPage:
		return "taskDefinition"
	case TaskDefinitionRevisionsPage:
		return "taskDefinitionRevisions"
	case ServiceEventsPage:
		return "serviceEvents"
	case LogPage:
		return "logs"
	case AutoScalingPage:
		return "autoscaling"
	default:
		return "unknownKind"
	}
}

func (k Kind) nextKind() Kind {
	switch k {
	case ClusterPage:
		return ServicePage
	case ServicePage:
		return TaskPage
	case TaskPage:
		return ContainerPage
	default:
		return ClusterPage
	}
}

func (k Kind) prevKind() Kind {
	switch k {
	case ClusterPage:
		return ClusterPage
	case ServicePage:
		return ClusterPage
	case TaskPage:
		return ServicePage
	case ContainerPage:
		return TaskPage
	default:
		return ClusterPage
	}
}

// App page name is kind string + "." + cluster arn
func (k Kind) getAppPageName(name string) string {
	switch k {
	case ClusterPage:
		return k.String()
	case ServicePage:
		return k.String() + "." + name
	case TaskPage:
		return k.String() + "." + name
	case ContainerPage:
		return k.String() + "." + name
	case DescriptionPage:
		return k.String() + "." + name
	default:
		return k.String()
	}
}

func (k Kind) getTablePageName(name string) string {
	pageName := k.getAppPageName(name)
	return pageName + ".table"
}

func (k Kind) getContentPageName(name string) string {
	pageName := k.getAppPageName(name)
	return pageName + "." + DescriptionPage.String()
}

type secondaryPageKeyMap = map[Kind][]KeyInput

var descriptionPageKeys = []KeyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: ctrlZ, description: backToPrevious},
}

var logPageKeys = []KeyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: string(rKey), description: realtimeLog},
	{key: ctrlR, description: reloadResource},
	{key: ctrlZ, description: backToPrevious},
}
