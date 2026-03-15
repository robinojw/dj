package tui

type agentLayer []AgentStatus

func buildDAGLayers(agents []AgentStatus) []agentLayer {
	depthMap := make(map[string]int, len(agents))
	byID := make(map[string]AgentStatus, len(agents))

	for _, agent := range agents {
		byID[agent.ID] = agent
	}

	var assignDepth func(id string, depth int)
	assignDepth = func(id string, depth int) {
		if existing, seen := depthMap[id]; seen && existing >= depth {
			return
		}
		depthMap[id] = depth
		for _, agent := range agents {
			if agent.ParentID == id {
				assignDepth(agent.ID, depth+1)
			}
		}
	}

	for _, agent := range agents {
		isRoot := agent.ParentID == "" || agent.ParentID == "root"
		if isRoot {
			assignDepth(agent.ID, 0)
		}
	}

	maxDepth := 0
	for _, depth := range depthMap {
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	layers := make([]agentLayer, maxDepth+1)
	for _, agent := range agents {
		depth := depthMap[agent.ID]
		layers[depth] = append(layers[depth], agent)
	}
	return layers
}

type cursorPos struct {
	Col int
	Row int
}
