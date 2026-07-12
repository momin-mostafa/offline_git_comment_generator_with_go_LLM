package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/git-comment/internal/config"
	"github.com/git-comment/internal/git"
	"github.com/git-comment/internal/ollama"
)

func main() {
	showDiff := flag.Bool("d", false, "Show staged diff before generating messages")
	flag.Parse()
	if !git.IsRepo() {
		fmt.Println("Current directory is not a Git repository.")
		os.Exit(1)
	}

	if !git.HasStagedChanges() {
		fmt.Println("No staged changes found.\nRun:\n\ngit add .")
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	client := ollama.NewClient(cfg.Host, cfg.Model, cfg.Temperature)

	if err := client.CheckAvailability(); err != nil {
		fmt.Println("Unable to connect to Ollama.\nEnsure Ollama is running.")
		os.Exit(1)
	}

	if err := client.ValidateModel(); err != nil {
		models, _ := client.ListModels()
		fmt.Printf("Configured model not found.\n\nAvailable models:\n")
		for _, m := range models {
			fmt.Printf("- %s\n", m)
		}
		os.Exit(1)
	}

	diff, err := git.GetStagedDiff()
	if err != nil {
		fmt.Printf("Error reading staged changes: %v\n", err)
		os.Exit(1)
	}

	if len(diff) == 0 {
		fmt.Println("No staged changes found.\nRun:\n\ngit add .")
		os.Exit(0)
	}

	if *showDiff {
		fmt.Println("--- Staged Diff ---")
		fmt.Println(diff)
		fmt.Println("--- End Diff ---")
	}

	fmt.Println("Analyzing staged changes...")
	messages, err := client.GenerateCommitMessages(diff)
	if err != nil {
		fmt.Printf("Error generating commit messages: %v\n", err)
		os.Exit(1)
	}

	if len(messages) == 0 {
		fmt.Println("No commit messages could be generated.")
		os.Exit(1)
	}

	fmt.Println("\nSuggested commit messages:")
	for i, msg := range messages {
		fmt.Printf("%d. %s\n", i+1, msg)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Aborted.")
		os.Exit(0)
	}()

	fmt.Print("\nChoose (1-" + fmt.Sprintf("%d", len(messages)) + ") or q: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = input[:len(input)-1]

	if input == "q" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}

	var selected int
	if _, err := fmt.Sscanf(input, "%d", &selected); err != nil || selected < 1 || selected > len(messages) {
		fmt.Println("Invalid selection.")
		os.Exit(1)
	}

	commitMsg := messages[selected-1]
	fmt.Printf("\nRunning\n\ngit commit -m \"%s\"\n\n", commitMsg)

	output, err := git.Commit(commitMsg)
	if err != nil {
		fmt.Println(output)
		os.Exit(1)
	}

	fmt.Println("✓ Commit successful")
}
