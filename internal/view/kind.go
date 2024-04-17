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
	case TaskKind, TaskDefinitionKind:
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
	case ServiceKind:
		return k.String() + "." + name
	case TaskKind:
		return k.String() + "." + name
	case ContainerKind:
		return k.String() + "." + name
	case TaskDefinitionKind:
		return k.String() + "." + name
	case DescriptionKind:
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
