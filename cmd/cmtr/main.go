package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/commitsmith/internal/config"
	"github.com/commitsmith/internal/git"
	"github.com/commitsmith/internal/ollama"
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

	client := ollama.NewClient(cfg.Host, cfg.Model, cfg.Temperature, cfg.MaxOptions, cfg.UseConventionalCommits)

	if err := client.CheckAvailability(); err != nil {
		fmt.Println("Unable to connect to Ollama.\nEnsure Ollama is running.")
		os.Exit(1)
	}

	if err := client.ValidateModel(); err != nil {
		models, listErr := client.ListModels()
		if listErr != nil || len(models) == 0 {
			fmt.Printf("Model not found and no alternatives available.\n")
			fmt.Printf("Configured model: %s\n", cfg.Model)
			os.Exit(1)
		}
		firstModel := models[0]
		fmt.Printf("Configured model '%s' not found. Auto-selecting '%s'.\n", cfg.Model, firstModel)
		client = ollama.NewClient(cfg.Host, firstModel, cfg.Temperature, cfg.MaxOptions, cfg.UseConventionalCommits)
		cfg.Model = firstModel
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
	fmt.Printf("0. Write your own commit message\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Aborted.")
		os.Exit(0)
	}()

	reader := bufio.NewReader(os.Stdin)
	var commitMsg string

	for {
		fmt.Printf("\nChoose (1-%d, 0 to write your own) or q: ", len(messages))
		input, _ := reader.ReadString('\n')
		input = input[:len(input)-1]

		if input == "q" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}

		if input == "" {
			commitMsg = messages[0]
			break
		}

		var selected int
		if _, err := fmt.Sscanf(input, "%d", &selected); err != nil || selected < 0 || selected > len(messages) {
			fmt.Println("Invalid selection.")
			continue
		}

		if selected == 0 {
			fmt.Print("\nEnter your commit message: ")
			customMsg, _ := reader.ReadString('\n')
			customMsg = customMsg[:len(customMsg)-1]
			if customMsg == "" {
				fmt.Println("Empty message. Try again.")
				continue
			}

			fmt.Println("Formatting commit message...")
			formatted, err := client.FormatCommitMessage(customMsg, diff)
			if err != nil {
				fmt.Printf("Error formatting message: %v\nUsing your message as-is.\n", err)
				commitMsg = customMsg
				break
			}

			fmt.Printf("\nFormatted: %s\n", formatted)
			fmt.Printf("Original:  %s\n", customMsg)

			for {
				fmt.Print("\nUse formatted? (Y/n): ")
				choice, _ := reader.ReadString('\n')
				choice = choice[:len(choice)-1]

				if choice == "" || choice == "y" || choice == "Y" {
					commitMsg = formatted
					break
				} else if choice == "n" || choice == "N" {
					commitMsg = customMsg
					break
				} else {
					fmt.Println("Invalid selection.")
				}
			}
			break
		}

		if selected >= 1 && selected <= len(messages) {
			commitMsg = messages[selected-1]
			break
		}
	}

	fmt.Printf("\nRunning\n\ngit commit -m \"%s\"\n\n", commitMsg)

	output, err := git.Commit(commitMsg)
	if err != nil {
		fmt.Println(output)
		os.Exit(1)
	}

	fmt.Println("✓ Commit successful")

	fmt.Print("\nPush now? (y/N): ")
	pushInput, _ := reader.ReadString('\n')
	pushInput = pushInput[:len(pushInput)-1]

	if pushInput == "y" || pushInput == "Y" {
		fmt.Println("Pushing...")
		pushOut, pushErr := git.Push()
		if pushErr != nil {
			fmt.Println(pushOut)
			os.Exit(1)
		}
		fmt.Println("✓ Push successful")
	}
}
