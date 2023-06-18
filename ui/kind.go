package ui

type Kind int

const (
	ClusterPage Kind = iota
	ServicePage
	TaskPage
	ContainerPage
	JsonPage
	TaskDefinitionPage
	TaskDefinitionRevisionsPage
	ServiceEventsPage
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
	case JsonPage:
		return "json"
	case TaskDefinitionPage:
		return "taskDefinition"
	case TaskDefinitionRevisionsPage:
		return "taskDefinitionRevisions"
	case ServiceEventsPage:
		return "serviceEvents"
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

func (k Kind) getAppPageName(name string) string {
	switch k {
	case ClusterPage:
		return k.String()
	case ServicePage:
		return k.String() + "." + name
	case TaskPage:
		return k.String() + "." + name
	case JsonPage:
		return k.String() + "." + name
	default:
		return k.String()
	}
}

func (k Kind) getTablePageName(name string) string {
	pageName := k.getAppPageName(name)
	return pageName + ".table"
}

func (k Kind) getJsonPageName(name string) string {
	pageName := k.getAppPageName(name)
	return pageName + ".json"
}
