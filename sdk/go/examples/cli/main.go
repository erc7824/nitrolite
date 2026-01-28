package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/c-bata/go-prompt"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: clearnode-cli <clearnode_ws_url>\n")
		fmt.Printf("Example: clearnode-cli wss://clearnode.example.com/ws\n")
		return
	}

	wsURL := os.Args[1]

	// Get config directory
	userConfDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Failed to get user config directory: %s\n", err.Error())
		return
	}
	configDir := path.Join(userConfDir, "clearnode-cli")
	if customDir := os.Getenv("CLEARNODE_CLI_CONFIG_DIR"); customDir != "" {
		configDir = customDir
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Failed to create config directory: %s\n", err.Error())
		return
	}

	// Initialize storage
	storagePath := path.Join(configDir, "config.db")
	store, err := NewStorage(storagePath)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %s\n", err.Error())
		return
	}

	// Create operator
	operator, err := NewOperator(wsURL, store)
	if err != nil {
		fmt.Printf("Failed to create operator: %s\n", err.Error())
		return
	}

	fmt.Println("🚀 Clearnode CLI - Developer Tool for Clearnode SDK")
	fmt.Printf("📡 Connected to: %s\n", wsURL)
	fmt.Printf("💾 Config directory: %s\n", configDir)
	fmt.Println("\n💡 Type 'help' for available commands or 'exit' to quit\n")

	// Terminal handling
	initialState, _ := term.GetState(int(os.Stdin.Fd()))
	handleExit := func() {
		term.Restore(int(os.Stdin.Fd()), initialState)
		exec.Command("stty", "sane").Run()
	}

	options := append(getStyleOptions(),
		prompt.OptionPrefix("clearnode> "),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Println("\n👋 Exiting Clearnode CLI")
				handleExit()
				os.Exit(0)
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlD,
			Fn:  func(buf *prompt.Buffer) {},
		}),
	)

	p := prompt.New(
		operator.Execute,
		operator.Complete,
		options...,
	)

	promptExitCh := make(chan struct{})
	go func() {
		p.Run()
		close(promptExitCh)
	}()

	select {
	case <-operator.Wait():
		fmt.Println("Operator exited.")
	case <-promptExitCh:
		fmt.Println("Prompt exited.")
	}

	handleExit()
	fmt.Println("👋 Goodbye!")
}

func getStyleOptions() []prompt.Option {
	return []prompt.Option{
		prompt.OptionTitle("Clearnode CLI"),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),

		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),

		prompt.OptionDescriptionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.Cyan),

		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionSelectedSuggestionBGColor(prompt.Green),

		prompt.OptionSelectedDescriptionTextColor(prompt.White),
		prompt.OptionSelectedDescriptionBGColor(prompt.DarkBlue),

		prompt.OptionShowCompletionAtStart(),
	}
}
