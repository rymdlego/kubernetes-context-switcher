package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	// Check if kubectl is installed
	if !commandExists("kubectl") {
		fmt.Println("kubectl could not be found. Please install kubectl to use this script.")
		os.Exit(1)
	}

	// Check if fzf is installed
	if !commandExists("fzf") {
		fmt.Println("fzf could not be found. Please install fzf to use this script.")
		os.Exit(1)
	}

	// Parse flags
	sketchybarNotify := false
	for _, arg := range os.Args {
		if arg == "--sketchybar" {
			// Check if fzf is installed
			if !commandExists("sketchybar") {
				fmt.Println("sketchybar could not be found. Please install sketchybar to use this script.")
				os.Exit(1)
			}
			sketchybarNotify = true
			// Remove the flag from args
			newArgs := []string{}
			for _, a := range os.Args {
				if a != "--sketchybar" {
					newArgs = append(newArgs, a)
				}
			}
			os.Args = newArgs
		}
	}

	if len(os.Args) > 1 && os.Args[1] == "help" {
		printHelp()
		return
	}

	contexts := getContexts()
	contexts = append([]string{"- Unset Context -"}, contexts...)

	if len(os.Args) > 1 {
		searchTerm := os.Args[1]
		contexts = filterContexts(contexts, searchTerm)
		if len(contexts) == 0 {
			fmt.Printf("No contexts found matching \"%s\"\n", searchTerm)
			os.Exit(1)
		}
	}

	if len(contexts) == 1 {
		selected := contexts[0]
		handleSelection(selected)
	} else {
		selected := selectContext(contexts)
		if selected == "" {
			fmt.Println("No context selected")
			return
		}
		handleSelection(selected)
	}
	// Notify sketchybar if flag was set
	if sketchybarNotify {
		notifySketchybar()
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getContexts() []string {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error getting contexts:", err)
		os.Exit(1)
	}
	return strings.Split(strings.TrimSpace(out.String()), "\n")
}

func filterContexts(contexts []string, searchTerm string) []string {
	var filtered []string
	for _, context := range contexts {
		if strings.Contains(strings.ToLower(context), strings.ToLower(searchTerm)) {
			filtered = append(filtered, context)
		}
	}
	return filtered
}

func selectContext(contexts []string) string {
	cmd := exec.Command("fzf-tmux", "-p", "-h", fmt.Sprintf("%d", len(contexts)+5), "--info", "hidden", "--border-label= Kubernetes Contexts ")
	cmd.Stdin = strings.NewReader(strings.Join(contexts, "\n"))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// Check if the error is due to SIGINT or SIGTSTP
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 130 {
					// Exit gracefully if Ctrl-C or Ctrl-Z was pressed
					return ""
				}
			}
		}
		fmt.Println("Error selecting context:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(out.String())
}

func handleSelection(selected string) {
	if selected == "- Unset Context -" {
		unsetContext()
		fmt.Println("Current context unset.")
	} else {
		switchContext(selected)
		fmt.Printf("Switched to context \"%s\".\n", selected)
	}
}

func switchContext(context string) {
	cmd := exec.Command("kubectl", "config", "use-context", context)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error switching context:", err)
		os.Exit(1)
	}
}

func unsetContext() {
	cmd := exec.Command("kubectl", "config", "unset", "current-context")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error unsetting context:", err)
		os.Exit(1)
	}
}

func notifySketchybar() {
	cmd := exec.Command("sketchybar", "--trigger", "kubernetes_context_switch")
	err := cmd.Run()
	if err != nil {
		// Silently ignore sketchybar errors
		return
	}
}

func printHelp() {
	appName := filepath.Base(os.Args[0])
	fmt.Printf(`%s - Kubernetes Context Switcher

Usage:
  %s [search-term]
  %s help

Flags:
  --sketchybar   Sends a sketchybar trigger (kubernetes_context_switch)

Options:
  search-term    Filter contexts based on the search term.
  help           Show this help message.

Description:
  This tool allows you to switch between Kubernetes contexts interactively using fzf.
  - If a search term is provided, it filters the contexts based on the term.
  - If the search term is 'unset', it unsets the current context.
  - If only one context matches, it switches to that context automatically.
  - If multiple contexts match, it allows you to select one interactively using fzf.
`, appName, appName, appName)
}
