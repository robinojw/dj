package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EditFileHandler returns a ToolHandler for edit_file / str_replace.
// Supports 3-tier whitespace-tolerant matching:
//  1. Exact match
//  2. Leading/trailing whitespace trimmed per line
//  3. All whitespace normalized to single spaces
func EditFileHandler(workspaceRoot string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		filePath, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("edit_file: missing required argument 'file_path'")
		}

		oldStr, ok := stringArg(args, "old_string")
		if !ok {
			return "", fmt.Errorf("edit_file: missing required argument 'old_string'")
		}

		newStr, hasNew := args["new_string"]
		newString := ""
		if hasNew {
			if s, ok := newStr.(string); ok {
				newString = s
			}
		}

		absPath, err := safePath(workspaceRoot, filePath)
		if err != nil {
			return "", fmt.Errorf("edit_file: %w", err)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("edit_file: %w", err)
		}

		content := string(data)
		replaced, count := replaceWithWhitespaceTolerance(content, oldStr, newString)
		if count == 0 {
			return "", fmt.Errorf("edit_file: old_string not found in %s", filePath)
		}

		if err := os.WriteFile(absPath, []byte(replaced), 0644); err != nil {
			return "", fmt.Errorf("edit_file: write failed: %w", err)
		}

		return fmt.Sprintf("Replaced %d occurrence(s) in %s", count, filePath), nil
	}
}

// replaceWithWhitespaceTolerance attempts replacement in 3 tiers:
//  1. Exact string match
//  2. Trimmed-line match (leading/trailing whitespace per line)
//  3. Normalized whitespace match (all runs of whitespace → single space)
func replaceWithWhitespaceTolerance(content, oldStr, newStr string) (string, int) {
	// Tier 1: exact match
	if strings.Contains(content, oldStr) {
		result := strings.Replace(content, oldStr, newStr, 1)
		return result, 1
	}

	// Tier 2: trimmed-line match
	if result, ok := trimmedLineReplace(content, oldStr, newStr); ok {
		return result, 1
	}

	// Tier 3: normalized whitespace
	if result, ok := normalizedReplace(content, oldStr, newStr); ok {
		return result, 1
	}

	return content, 0
}

// trimmedLineReplace finds oldStr by comparing lines with leading/trailing whitespace trimmed.
func trimmedLineReplace(content, oldStr, newStr string) (string, bool) {
	contentLines := strings.Split(content, "\n")
	oldLines := strings.Split(oldStr, "\n")

	if len(oldLines) > len(contentLines) {
		return "", false
	}

	for i := 0; i <= len(contentLines)-len(oldLines); i++ {
		match := true
		for j, oldLine := range oldLines {
			if strings.TrimSpace(contentLines[i+j]) != strings.TrimSpace(oldLine) {
				match = false
				break
			}
		}
		if match {
			// Replace the matched lines, preserving original indentation for new content
			before := strings.Join(contentLines[:i], "\n")
			after := strings.Join(contentLines[i+len(oldLines):], "\n")

			var result string
			if before != "" && after != "" {
				result = before + "\n" + newStr + "\n" + after
			} else if before != "" {
				result = before + "\n" + newStr
			} else if after != "" {
				result = newStr + "\n" + after
			} else {
				result = newStr
			}
			return result, true
		}
	}
	return "", false
}

// normalizedReplace normalizes all whitespace runs to single spaces for comparison.
func normalizedReplace(content, oldStr, newStr string) (string, bool) {
	normContent := normalizeWhitespace(content)
	normOld := normalizeWhitespace(oldStr)

	idx := strings.Index(normContent, normOld)
	if idx == -1 {
		return "", false
	}

	// Map the normalized index back to the original content.
	// Walk the original content tracking normalized position.
	origStart := mapNormIdx(content, idx)
	origEnd := mapNormIdx(content, idx+len(normOld))

	result := content[:origStart] + newStr + content[origEnd:]
	return result, true
}

// normalizeWhitespace collapses all whitespace runs to a single space and trims.
func normalizeWhitespace(s string) string {
	var sb strings.Builder
	inSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !inSpace {
				sb.WriteByte(' ')
				inSpace = true
			}
		} else {
			sb.WriteRune(r)
			inSpace = false
		}
	}
	return strings.TrimSpace(sb.String())
}

// mapNormIdx maps a position in the normalized string back to the original.
func mapNormIdx(original string, normIdx int) int {
	normPos := 0
	inSpace := false
	for i, r := range original {
		if normPos >= normIdx {
			return i
		}
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !inSpace {
				normPos++
				inSpace = true
			}
		} else {
			normPos++
			inSpace = false
		}
	}
	return len(original)
}
