package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		handleSelection(selected)
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
