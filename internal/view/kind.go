package view

type kind int

const (
	ClusterKind kind = iota
	ServiceKind
	TaskKind
	ContainerKind
	TaskDefinitionKind
	HelpKind
	DescriptionKind
	ServiceEventsKind
	ServiceDeployment
	LogKind
	AutoScalingKind
	ModalKind
	EmptyKind
)

func (k kind) String() string {
	switch k {
	case ClusterKind:
		return "clusters"
	case ServiceKind:
		return "services"
	case TaskKind:
		return "tasks"
	case ContainerKind:
		return "containers"
	case HelpKind:
		return "help"
	case DescriptionKind:
		return "description"
	case TaskDefinitionKind:
		return "task definitions"
	case ServiceEventsKind:
		return "service events"
	case ServiceDeployment:
		return "service deployments"
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

func (k kind) nextKind() kind {
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

func (k kind) prevKind() kind {
	switch k {
	case ClusterKind:
		return ClusterKind
	case ServiceKind:
		return ClusterKind
	case TaskKind, TaskDefinitionKind, ServiceDeployment:
		return ServiceKind
	case ContainerKind:
		return TaskKind
	default:
		return ClusterKind
	}
}

// App page name is kind string + "." + cluster arn
func (k kind) getAppPageName(name string) string {
	switch k {
	case ClusterKind:
		return k.String()
	case ServiceKind, TaskKind, ContainerKind, TaskDefinitionKind, ServiceDeployment, DescriptionKind:
		return k.String() + "." + name
	default:
		return k.String()
	}
}

func (k kind) getTablePageName(name string) string {
	pageName := k.getAppPageName(name)
	return pageName + ".table"
}

func (k kind) getSecondaryPageName(name string) string {
	pageName := k.getAppPageName(name)
	return pageName + "." + DescriptionKind.String()
}
