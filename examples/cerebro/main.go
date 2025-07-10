package main

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
	"golang.org/x/term"

	"github.com/erc7824/nitrolite/examples/bridge/clearnet"
	"github.com/erc7824/nitrolite/examples/bridge/storage"
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

	store, err := storage.NewStorage(os.Getenv("CEREBRO_STORE_PATH"))
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
	defer term.Restore(int(os.Stdin.Fd()), initialState)

	for {
		t := prompt.Input(">>> ", operator.Complete,
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
		)

		operator.Execute(t)
	}
}

func emptyCompleter(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}
