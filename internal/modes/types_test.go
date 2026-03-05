package modes

import "testing"

func TestExecutionMode_String(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		want string
	}{
		{ModeConfirm, "Confirm"},
		{ModePlan, "Plan"},
		{ModeTurbo, "Turbo"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionMode_StatusLabel(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		want string
	}{
		{ModeConfirm, "⏸ CONFIRM"},
		{ModePlan, "◎ PLAN"},
		{ModeTurbo, "⚡ TURBO"},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			if got := tt.mode.StatusLabel(); got != tt.want {
				t.Errorf("StatusLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
