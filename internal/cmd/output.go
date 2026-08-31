package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	colorStep    = color.New(color.Faint)
	colorInfo    = color.New(color.FgCyan)
	colorSuccess = color.New(color.FgGreen)
	colorWarn    = color.New(color.FgYellow)
	colorError   = color.New(color.FgRed, color.Bold)
	colorPrompt  = color.New(color.Bold)
	colorAdded   = color.New(color.FgGreen)
	colorRemoved = color.New(color.FgRed)
)

func shouldLog(msgLevel string) bool {
	cl, ok := verbosityLevels[verbosity]
	if !ok {
		cl = 2
	}
	ml, ok := verbosityLevels[msgLevel]
	if !ok {
		ml = 2
	}
	return ml <= cl
}

func outStep(msg string) {
	outStepLevel("info", msg)
}

func outStepLevel(level, msg string) {
	if !shouldLog(level) {
		return
	}
	colorStep.Println("→ " + msg)
}

func outInfo(msg string) {
	if !shouldLog("info") {
		return
	}
	colorInfo.Println(msg)
}

func outSuccess(msg string) {
	if !shouldLog("info") {
		return
	}
	colorSuccess.Println("✓ " + msg)
}

func outWarn(msg string) {
	if !shouldLog("warning") {
		return
	}
	colorWarn.Println("⚠ " + msg)
}

func outError(msg string) {
	if !shouldLog("error") {
		return
	}
	colorError.Println("✗ " + msg)
}

func outPrompt(msg string) {
	colorPrompt.Print(msg + " ")
}

func formatACLList(current []string, changed string, act action) string {
	var sb strings.Builder
	for _, entry := range current {
		if entry == changed {
			if act == actionAdd {
				sb.WriteString("  ")
				colorAdded.Fprint(&sb, "+ "+entry)
				sb.WriteString("\n")
			} else {
				sb.WriteString("  ")
				colorRemoved.Fprint(&sb, "- "+entry)
				sb.WriteString("\n")
			}
		} else {
			sb.WriteString(fmt.Sprintf("  • %s\n", entry))
		}
	}
	return sb.String()
}
