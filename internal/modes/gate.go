package modes

// Gate controls tool access through allow/deny lists and mode rules.
type Gate struct {
	mode      ExecutionMode
	allowList []string
	denyList  []string
	registry  ToolClassifier
}

// NewGate creates a gate with the given mode and lists.
func NewGate(mode ExecutionMode, allowList, denyList []string) *Gate {
	return &Gate{
		mode:      mode,
		allowList: allowList,
		denyList:  denyList,
	}
}

// NewGateWithRegistry creates a gate that consults tool annotations for classification.
func NewGateWithRegistry(mode ExecutionMode, allowList, denyList []string, registry ToolClassifier) *Gate {
	return &Gate{
		mode:      mode,
		allowList: allowList,
		denyList:  denyList,
		registry:  registry,
	}
}

// SetMode updates the gate's execution mode.
func (g *Gate) SetMode(mode ExecutionMode) {
	g.mode = mode
}

// Evaluate determines whether a tool call should be allowed.
func (g *Gate) Evaluate(toolName string, args map[string]any) GateDecision {
	if g.isDenied(toolName) {
		return GateDeny
	}

	if g.isAllowed(toolName) {
		return GateAllow
	}

	class := ClassifyToolWithRegistry(toolName, g.registry)

	switch g.mode {
	case ModeTurbo:
		return GateAllow
	case ModePlan:
		if class == ToolRead || class == ToolMCPRead {
			return GateAllow
		}
		return GateDeny
	case ModeConfirm:
		if class == ToolRead || class == ToolMCPRead {
			return GateAllow
		}
		return GateAskUser
	default:
		return GateDeny
	}
}

// isDenied checks if tool matches deny list (with glob).
func (g *Gate) isDenied(toolName string) bool {
	for _, pattern := range g.denyList {
		if MatchGlob(pattern, toolName) {
			return true
		}
	}
	return false
}

// isAllowed checks if tool matches allow list (with glob).
func (g *Gate) isAllowed(toolName string) bool {
	for _, pattern := range g.allowList {
		if MatchGlob(pattern, toolName) {
			return true
		}
	}
	return false
}

// AllowForSession adds a tool to the session allow list.
func (g *Gate) AllowForSession(toolName string) {
	g.allowList = append(g.allowList, toolName)
}
