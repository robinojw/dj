package agents

import (
	"sync"
	"testing"
)

func TestValidateDAG_NoCycle(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}
	dag, err := buildDAG(subtasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dag.inDegree["a"] != 0 {
		t.Errorf("a in-degree = %d, want 0", dag.inDegree["a"])
	}
	if dag.inDegree["b"] != 1 {
		t.Errorf("b in-degree = %d, want 1", dag.inDegree["b"])
	}
	if dag.inDegree["c"] != 1 {
		t.Errorf("c in-degree = %d, want 1", dag.inDegree["c"])
	}
}

func TestValidateDAG_Cycle(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a", DependsOn: []string{"b"}},
		{ID: "b", DependsOn: []string{"a"}},
	}
	_, err := buildDAG(subtasks)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestValidateDAG_SelfCycle(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a", DependsOn: []string{"a"}},
	}
	_, err := buildDAG(subtasks)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestValidateDAG_ThreeNodeCycle(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a", DependsOn: []string{"c"}},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}
	_, err := buildDAG(subtasks)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestValidateDAG_DiamondNoCycle(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"a"}},
		{ID: "d", DependsOn: []string{"b", "c"}},
	}
	dag, err := buildDAG(subtasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dag.inDegree["d"] != 2 {
		t.Errorf("d in-degree = %d, want 2", dag.inDegree["d"])
	}
}

func TestValidateDAG_MissingDependency(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a", DependsOn: []string{"nonexistent"}},
	}
	_, err := buildDAG(subtasks)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}

func TestSkipDependents(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}
	dag, _ := buildDAG(subtasks)

	workers := map[string]*Worker{
		"a": {ID: "a", Status: "error"},
		"b": {ID: "b", Status: "pending"},
		"c": {ID: "c", Status: "pending"},
	}

	updates := make(chan WorkerUpdate, 10)
	skipped := skipDependents(dag, "a", workers, updates)

	if skipped != 2 {
		t.Errorf("skipDependents returned %d, want 2", skipped)
	}
	if workers["b"].Status != "skipped" {
		t.Errorf("b status = %q, want skipped", workers["b"].Status)
	}
	if workers["c"].Status != "skipped" {
		t.Errorf("c status = %q, want skipped", workers["c"].Status)
	}
}

func TestSkipDependents_PartialDAG(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c"},
	}
	dag, _ := buildDAG(subtasks)

	workers := map[string]*Worker{
		"a": {ID: "a", Status: "error"},
		"b": {ID: "b", Status: "pending"},
		"c": {ID: "c", Status: "running"},
	}

	updates := make(chan WorkerUpdate, 10)
	skipped := skipDependents(dag, "a", workers, updates)

	if skipped != 1 {
		t.Errorf("skipDependents returned %d, want 1", skipped)
	}
	if workers["b"].Status != "skipped" {
		t.Errorf("b status = %q, want skipped", workers["b"].Status)
	}
	if workers["c"].Status != "running" {
		t.Errorf("c status = %q, want running (untouched)", workers["c"].Status)
	}
}

func TestExecutionOrder_LinearChain(t *testing.T) {
	var mu sync.Mutex
	var order []string

	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}

	dag, err := buildDAG(subtasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ready := dag.readySet()
	if len(ready) != 1 || ready[0] != "a" {
		t.Errorf("ready set = %v, want [a]", ready)
	}

	newlyReady := dag.markCompleted("a")
	mu.Lock()
	order = append(order, "a")
	mu.Unlock()
	if len(newlyReady) != 1 || newlyReady[0] != "b" {
		t.Errorf("after a completes, newly ready = %v, want [b]", newlyReady)
	}

	newlyReady = dag.markCompleted("b")
	mu.Lock()
	order = append(order, "b")
	mu.Unlock()
	if len(newlyReady) != 1 || newlyReady[0] != "c" {
		t.Errorf("after b completes, newly ready = %v, want [c]", newlyReady)
	}

	mu.Lock()
	order = append(order, "c")
	mu.Unlock()

	if order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Errorf("execution order = %v, want [a b c]", order)
	}
}

func TestExecutionOrder_Diamond(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"a"}},
		{ID: "d", DependsOn: []string{"b", "c"}},
	}

	dag, _ := buildDAG(subtasks)

	ready := dag.readySet()
	if len(ready) != 1 || ready[0] != "a" {
		t.Errorf("ready = %v, want [a]", ready)
	}

	newlyReady := dag.markCompleted("a")
	if len(newlyReady) != 2 {
		t.Errorf("after a, newly ready = %v, want [b c]", newlyReady)
	}

	newlyReady = dag.markCompleted("b")
	if len(newlyReady) != 0 {
		t.Errorf("after b, newly ready = %v, want []", newlyReady)
	}

	newlyReady = dag.markCompleted("c")
	if len(newlyReady) != 1 || newlyReady[0] != "d" {
		t.Errorf("after c, newly ready = %v, want [d]", newlyReady)
	}
}

func TestFlatTasks_AllReady(t *testing.T) {
	subtasks := []Subtask{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}
	dag, _ := buildDAG(subtasks)

	ready := dag.readySet()
	if len(ready) != 3 {
		t.Errorf("ready = %v, want 3 items", ready)
	}
}
