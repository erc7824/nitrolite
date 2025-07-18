package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/c-bata/go-prompt"
	"golang.org/x/term"

	"github.com/erc7824/nitrolite/examples/cerebro/clearnet"
	"github.com/erc7824/nitrolite/examples/cerebro/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: cerebro <clearnode_ws_url>\n")
		return
	}

	clearnodeWSURL := os.Args[1]
	clearnode, err := clearnet.NewClearnodeClient(clearnodeWSURL)
	if err != nil {
		fmt.Printf("Failed to connect to Clearnode WebSocket: %s\n", err.Error())
		return
	}

	userConfDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Failed to get user config directory: %s\n", err.Error())
		return
	}
	configDir := path.Join(userConfDir, "cerebro")
	if customDir := os.Getenv("CEREBRO_CONFIG_DIR"); customDir != "" {
		configDir = customDir
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Failed to create config directory: %s\n", err.Error())
		return
	}

	storagePath := path.Join(configDir, "storage.db")
	store, err := storage.NewStorage(storagePath)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %s\n", err.Error())
		return
	}

	operator, err := NewOperator(clearnode, store)
	if err != nil {
		fmt.Printf("Failed to create operator: %s\n", err.Error())
		return
	}

	initialState, _ := term.GetState(int(os.Stdin.Fd()))
	handleExit := func() {
		term.Restore(int(os.Stdin.Fd()), initialState)
		exec.Command("stty", "sane").Run()
	}

	options := append(getStyleOptions(),
		prompt.OptionPrefix(">>> "),

		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Println("Exiting Cerebro CLI.")
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
	case <-clearnode.WaitCh():
		fmt.Println("Clearnode client disconnected.")
	case <-operator.Wait():
		fmt.Println("Operator exited.")
	case <-promptExitCh:
		fmt.Println("Prompt exited.")
	}
	handleExit()
	fmt.Println("Exiting Cerebro CLI.")
}

func emptyCompleter(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func getStyleOptions() []prompt.Option {
	return []prompt.Option{
		prompt.OptionTitle("Cerebro CLI"),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Cyan),

		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.DarkBlue),

		prompt.OptionDescriptionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.Yellow),

		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionSelectedSuggestionBGColor(prompt.Yellow),

		prompt.OptionSelectedDescriptionTextColor(prompt.White),
		prompt.OptionSelectedDescriptionBGColor(prompt.DarkBlue),

		prompt.OptionShowCompletionAtStart(),
	}
}
