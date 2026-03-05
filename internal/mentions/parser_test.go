package mentions

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseFileMention(t *testing.T) {
	mentions := Parse("Fix the bug in @src/auth/handler.go please")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionFile {
		t.Errorf("Expected file mention, got %s", mentions[0].Type)
	}
	if mentions[0].Value != "src/auth/handler.go" {
		t.Errorf("Expected 'src/auth/handler.go', got %q", mentions[0].Value)
	}
}

func TestParseURLMention(t *testing.T) {
	mentions := Parse("Check @https://docs.stripe.com/api for reference")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionURL {
		t.Errorf("Expected URL mention, got %s", mentions[0].Type)
	}
}

func TestParseFunctionMention(t *testing.T) {
	mentions := Parse("Look at @fn:CreateSession")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionFunction {
		t.Errorf("Expected function mention, got %s", mentions[0].Type)
	}
	if mentions[0].Value != "CreateSession" {
		t.Errorf("Expected 'CreateSession', got %q", mentions[0].Value)
	}
}

func TestParseGitMention(t *testing.T) {
	mentions := Parse("Show changes from @git:HEAD~3")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionGit {
		t.Errorf("Expected git mention, got %s", mentions[0].Type)
	}
	if mentions[0].Value != "HEAD~3" {
		t.Errorf("Expected 'HEAD~3', got %q", mentions[0].Value)
	}
}

func TestParseTestMention(t *testing.T) {
	mentions := Parse("Run @test:TestAuthFlow and show me the output")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionTest {
		t.Errorf("Expected test mention, got %s", mentions[0].Type)
	}
}

func TestParseMultipleMentions(t *testing.T) {
	mentions := Parse("Compare @src/old.go with @src/new.go")
	if len(mentions) != 2 {
		t.Fatalf("Expected 2 mentions, got %d", len(mentions))
	}
}

func TestParseNoMentions(t *testing.T) {
	mentions := Parse("Just a normal message with email user@example.com")
	if len(mentions) != 0 {
		t.Errorf("Expected 0 mentions, got %d", len(mentions))
	}
}

func TestStripMentions(t *testing.T) {
	input := "Fix @src/auth.go and run @test:TestAuth"
	stripped := StripMentions(input)
	if stripped != "Fix  and run " {
		t.Errorf("Expected mentions stripped, got %q", stripped)
	}
}

func TestFormatResolved(t *testing.T) {
	resolved := []ResolvedMention{
		{
			Mention: Mention{Type: MentionFile, Value: "main.go"},
			Content: "package main\n\nfunc main() {}",
		},
		{
			Mention: Mention{Type: MentionGit, Value: "HEAD~1"},
			Error:   fmt.Errorf("not a git repo"),
		},
	}

	output := FormatResolved(resolved)
	if !strings.Contains(output, "package main") {
		t.Error("Expected file content in output")
	}
	if !strings.Contains(output, "not a git repo") {
		t.Error("Expected error message in output")
	}
}
