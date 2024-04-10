package ui

type Kind int

const (
	ClusterKind Kind = iota
	ServiceKind
	TaskKind
	ContainerKind
	TaskDefinitionKind
	DescriptionKind
	ServiceEventsKind
	LogKind
	AutoScalingKind
	ModalKind
	EmptyKind
)

func (k Kind) String() string {
	switch k {
	case ClusterKind:
		return "clusters"
	case ServiceKind:
		return "services"
	case TaskKind:
		return "tasks"
	case ContainerKind:
		return "containers"
	case DescriptionKind:
		return "description"
	case TaskDefinitionKind:
		return "task definitions"
	case ServiceEventsKind:
		return "service events"
	case LogKind:
		return "logs"
	case AutoScalingKind:
		return "autoscaling"
	case ModalKind:
		return "modal"
	default:
		return "unknownKind"
	}
}

func (k Kind) nextKind() Kind {
	switch k {
	case ClusterKind:
		return ServiceKind
	case ServiceKind:
		return TaskKind
	case TaskKind:
		return ContainerKind
	default:
		return ClusterKind
	}
}

func (k Kind) prevKind() Kind {
	switch k {
	case ClusterKind:
		return ClusterKind
	case ServiceKind:
		return ClusterKind
	case TaskKind, TaskDefinitionKind:
		return ServiceKind
	case ContainerKind:
		return TaskKind
	default:
		return ClusterKind
	}
}

// App page name is kind string + "." + cluster arn
func (k Kind) getAppPageName(name string) string {
	switch k {
	case ClusterKind:
		return k.String()
	case ServiceKind:
		return k.String() + "." + name
	case TaskKind:
		return k.String() + "." + name
	case ContainerKind:
		return k.String() + "." + name
	case DescriptionKind:
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
	return pageName + "." + DescriptionKind.String()
}

type secondaryPageKeyMap = map[Kind][]KeyInput

var descriptionPageKeys = []KeyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: string(eKey), description: openInEditor},
	{key: ctrlZ, description: backToPrevious},
}

var logPageKeys = []KeyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: string(rKey), description: realtimeLog},
	{key: ctrlR, description: reloadResource},
	{key: ctrlZ, description: backToPrevious},
}
