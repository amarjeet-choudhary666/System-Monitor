package logs

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// LogLevel represents different log levels
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	DEBUG LogLevel = "DEBUG"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	Level   LogLevel
	Message string
	Time    string
}

// LogStats holds statistics about log analysis
type LogStats struct {
	LevelCounts  map[LogLevel]int `json:"level_counts"`
	TopErrors    []ErrorFrequency `json:"top_errors"`
	TotalEntries int              `json:"total_entries"`
}

// ErrorFrequency represents error message frequency
type ErrorFrequency struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// LogAnalyzer handles log file analysis
type LogAnalyzer struct {
	logPattern *regexp.Regexp
}

// NewLogAnalyzer creates a new log analyzer instance
func NewLogAnalyzer() *LogAnalyzer {
	// Pattern to match common log formats: [LEVEL] message or LEVEL: message
	pattern := regexp.MustCompile(`(?i)\[(INFO|WARN|ERROR|DEBUG)\]|^(INFO|WARN|ERROR|DEBUG):`)

	return &LogAnalyzer{
		logPattern: pattern,
	}
}

// ParseLogFile parses a log file and returns statistics
func (la *LogAnalyzer) ParseLogFile(filePath string) (*LogStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	stats := &LogStats{
		LevelCounts: make(map[LogLevel]int),
		TopErrors:   make([]ErrorFrequency, 0),
	}

	errorMessages := make(map[string]int)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entry := la.ParseLine(line)
		if entry != nil {
			stats.LevelCounts[entry.Level]++
			stats.TotalEntries++

			// Track error messages for frequency analysis
			if entry.Level == ERROR {
				errorMessages[entry.Message]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Calculate top 5 most frequent errors
	stats.TopErrors = la.getTopErrors(errorMessages, 5)

	return stats, nil
}

// ParseLine extracts log level and message from a single line
func (la *LogAnalyzer) ParseLine(line string) *LogEntry {
	matches := la.logPattern.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil
	}

	var level LogLevel
	var message string

	// Check which group matched
	if matches[1] != "" {
		level = LogLevel(strings.ToUpper(matches[1]))
		// Extract message after [LEVEL]
		parts := strings.SplitN(line, "]", 2)
		if len(parts) > 1 {
			message = strings.TrimSpace(parts[1])
		}
	} else if matches[2] != "" {
		level = LogLevel(strings.ToUpper(matches[2]))
		// Extract message after LEVEL:
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 1 {
			message = strings.TrimSpace(parts[1])
		}
	}

	if message == "" {
		message = line
	}

	return &LogEntry{
		Level:   level,
		Message: message,
	}
}

// getTopErrors returns the top N most frequent error messages
func (la *LogAnalyzer) getTopErrors(errorMessages map[string]int, topN int) []ErrorFrequency {
	// Convert map to slice for sorting
	errors := make([]ErrorFrequency, 0, len(errorMessages))
	for msg, count := range errorMessages {
		errors = append(errors, ErrorFrequency{
			Message: msg,
			Count:   count,
		})
	}

	// Sort by frequency (descending)
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Count > errors[j].Count
	})

	// Return top N errors
	if len(errors) > topN {
		return errors[:topN]
	}
	return errors
}

// PrintStats prints log statistics in a formatted way
func (la *LogAnalyzer) PrintStats(stats *LogStats) {
	fmt.Println("=== Log Analysis Results ===")
	fmt.Printf("Total log entries: %d\n\n", stats.TotalEntries)

	fmt.Println("Log Level Counts:")
	for level, count := range stats.LevelCounts {
		fmt.Printf("  %s: %d\n", level, count)
	}

	fmt.Println("\nTop 5 Most Frequent Errors:")
	for i, err := range stats.TopErrors {
		fmt.Printf("  %d. [%d times] %s\n", i+1, err.Count, err.Message)
	}
}
