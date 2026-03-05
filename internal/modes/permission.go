package modes

// PermissionRequest is sent from worker to TUI when permission is needed.
type PermissionRequest struct {
	ID       string             // unique request ID
	WorkerID string             // worker making the request
	Tool     string             // tool name (e.g. "write_file")
	Args     map[string]any     // tool arguments
	RespCh   chan PermissionResp // response channel (worker blocks on this)
}

// PermissionResp is sent from TUI to worker after user decision.
type PermissionResp struct {
	Allowed     bool          // true if user approved
	RememberFor RememberScope // how to remember this decision
}

// RememberScope determines how long a permission decision lasts.
type RememberScope int

const (
	RememberOnce    RememberScope = iota // allow this single call
	RememberSession                      // allow for current session
	RememberAlways                       // persist to config
)

func (r RememberScope) String() string {
	switch r {
	case RememberOnce:
		return "Once"
	case RememberSession:
		return "Session"
	case RememberAlways:
		return "Always"
	default:
		return "Unknown"
	}
}
