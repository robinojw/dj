package orchestrator

import "testing"

const (
	actionSpawn    = "spawn"
	actionMessage  = "message"
	actionComplete = "complete"
	personaArch    = "architect"
	personaTest    = "test"
	targetArch1    = "arch-1"
	taskDesignAPI  = "Design API"

	expectedOneCommand  = 1
	expectedTwoCommands = 2

	errExpectedDCommands = "expected %d command(s), got %d"
	errExpectedSGotS     = "expected %s, got %s"
)

func TestCommandParserSingleBlock(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("Some text before\n```dj-command\n")
	parser.Feed(`{"action":"spawn","persona":"architect","task":"Design API"}`)
	parser.Feed("\n```\nSome text after")

	commands := parser.Flush()
	if len(commands) != expectedOneCommand {
		testing.Fatalf(errExpectedDCommands, expectedOneCommand, len(commands))
	}
	if commands[0].Action != actionSpawn {
		testing.Errorf(errExpectedSGotS, actionSpawn, commands[0].Action)
	}
	if commands[0].Persona != personaArch {
		testing.Errorf(errExpectedSGotS, personaArch, commands[0].Persona)
	}
	if commands[0].Task != taskDesignAPI {
		testing.Errorf(errExpectedSGotS, taskDesignAPI, commands[0].Task)
	}
}

func TestCommandParserMultipleBlocks(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-command\n{\"action\":\"spawn\",\"persona\":\"architect\",\"task\":\"A\"}\n```\n")
	parser.Feed("```dj-command\n{\"action\":\"spawn\",\"persona\":\"test\",\"task\":\"B\"}\n```\n")

	commands := parser.Flush()
	if len(commands) != expectedTwoCommands {
		testing.Fatalf(errExpectedDCommands, expectedTwoCommands, len(commands))
	}
	if commands[0].Persona != personaArch {
		testing.Errorf(errExpectedSGotS, personaArch, commands[0].Persona)
	}
	if commands[1].Persona != personaTest {
		testing.Errorf(errExpectedSGotS, personaTest, commands[1].Persona)
	}
}

func TestCommandParserChunkedDelta(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-")
	parser.Feed("command\n{\"action\":")
	parser.Feed("\"message\",\"target\":\"arch-1\"")
	parser.Feed(",\"content\":\"hello\"}\n`")
	parser.Feed("``\n")

	commands := parser.Flush()
	if len(commands) != expectedOneCommand {
		testing.Fatalf(errExpectedDCommands, expectedOneCommand, len(commands))
	}
	if commands[0].Action != actionMessage {
		testing.Errorf(errExpectedSGotS, actionMessage, commands[0].Action)
	}
	if commands[0].Target != targetArch1 {
		testing.Errorf(errExpectedSGotS, targetArch1, commands[0].Target)
	}
}

func TestCommandParserNoCommands(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("Just regular text with no commands at all.")

	commands := parser.Flush()
	if len(commands) != 0 {
		testing.Errorf("expected 0 commands, got %d", len(commands))
	}
}

func TestCommandParserMalformedJSON(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-command\n{invalid json}\n```\n")

	commands := parser.Flush()
	if len(commands) != 0 {
		testing.Errorf("expected 0 commands for malformed JSON, got %d", len(commands))
	}
}

func TestCommandParserStripsCommands(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("Before\n```dj-command\n{\"action\":\"complete\",\"content\":\"done\"}\n```\nAfter")

	_ = parser.Flush()
	cleaned := parser.CleanedText()
	expectedCleaned := "Before\nAfter"
	if cleaned != expectedCleaned {
		testing.Errorf("expected %q, got %q", expectedCleaned, cleaned)
	}
}

func TestCommandParserCompleteAction(testing *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-command\n{\"action\":\"complete\",\"content\":\"Task finished with 2 findings\"}\n```\n")

	commands := parser.Flush()
	if len(commands) != expectedOneCommand {
		testing.Fatalf(errExpectedDCommands, expectedOneCommand, len(commands))
	}
	if commands[0].Action != actionComplete {
		testing.Errorf(errExpectedSGotS, actionComplete, commands[0].Action)
	}
	expectedContent := "Task finished with 2 findings"
	if commands[0].Content != expectedContent {
		testing.Errorf(errExpectedSGotS, expectedContent, commands[0].Content)
	}
}
