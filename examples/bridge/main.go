package main

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: clearbridge <clearnode_ws_url>\n")
		return
	}

	clearnodeWSURL := os.Args[1]
	clearnode, err := NewClearnodeClient(clearnodeWSURL)
	if err != nil {
		fmt.Printf("Failed to connect to Clearnode WebSocket: %s\n", err.Error())
		return
	}

	store, err := NewStorage(os.Getenv("CLEARBRIDGE_STORE_PATH"))
	if err != nil {
		fmt.Printf("Failed to initialize storage: %s\n", err.Error())
		return
	}

	operator, err := NewOperator(clearnode, store)
	if err != nil {
		fmt.Printf("Failed to create operator: %s\n", err.Error())
		return
	}

	for {
		t := prompt.Input(">>> ", operator.Complete,
			prompt.OptionTitle("ClearBridge CLI"),
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
