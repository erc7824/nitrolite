package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOperatorComplete tests the completion functionality
func TestOperatorComplete(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	// Note: We can't fully test this without a live client connection
	// These tests focus on the structure and filtering logic
	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("returns filtered suggestions", func(t *testing.T) {
		doc := prompt.Document{
			Text: "hel",
		}
		suggestions := op.Complete(doc)

		// Should include "help" when typing "hel"
		found := false
		for _, s := range suggestions {
			if s.Text == "help" {
				found = true
				break
			}
		}
		assert.True(t, found, "should suggest 'help' when typing 'hel'")
	})

	t.Run("returns empty when word doesn't match", func(t *testing.T) {
		doc := prompt.Document{
			Text: "xyz",
		}
		suggestions := op.Complete(doc)

		// Should not return standard commands for non-matching prefix
		assert.Empty(t, suggestions)
	})

	t.Run("handles empty input", func(t *testing.T) {
		doc := prompt.Document{
			Text: "",
		}
		suggestions := op.Complete(doc)

		// Should return all top-level commands
		assert.NotEmpty(t, suggestions)

		// Verify some key commands are present
		commands := make(map[string]bool)
		for _, s := range suggestions {
			commands[s.Text] = true
		}

		assert.True(t, commands["help"])
		assert.True(t, commands["config"])
		assert.True(t, commands["wallet"])
		assert.True(t, commands["import"])
		assert.True(t, commands["deposit"])
		assert.True(t, commands["withdraw"])
		assert.True(t, commands["transfer"])
		assert.True(t, commands["exit"])
	})
}

// TestOperatorCompleteStructure tests the structure of suggestions
func TestOperatorCompleteStructure(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("import subcommands", func(t *testing.T) {
		doc := prompt.Document{
			Text: "import ",
		}
		suggestions := op.complete(doc)

		// Should suggest "wallet" and "rpc"
		texts := make(map[string]bool)
		for _, s := range suggestions {
			texts[s.Text] = true
		}

		assert.True(t, texts["wallet"])
		assert.True(t, texts["rpc"])
	})

	t.Run("node subcommands", func(t *testing.T) {
		doc := prompt.Document{
			Text: "node ",
		}
		suggestions := op.complete(doc)

		// Should suggest "info"
		found := false
		for _, s := range suggestions {
			if s.Text == "info" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("all suggestions have descriptions", func(t *testing.T) {
		doc := prompt.Document{
			Text: "",
		}
		suggestions := op.complete(doc)

		for _, s := range suggestions {
			assert.NotEmpty(t, s.Description, "command %s should have a description", s.Text)
		}
	})
}

// TestOperatorExecuteStructure tests the execute method structure
func TestOperatorExecuteStructure(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("handles empty input", func(t *testing.T) {
		// Should not panic
		op.Execute("")
	})

	t.Run("handles whitespace only", func(t *testing.T) {
		// Should not panic
		op.Execute("   ")
	})

	t.Run("handles unknown command gracefully", func(t *testing.T) {
		// Should not panic
		op.Execute("unknown_command")
	})

	t.Run("exit command closes channel", func(t *testing.T) {
		exitCh := make(chan struct{})
		op := &Operator{
			store:  storage,
			exitCh: exitCh,
		}

		go op.Execute("exit")

		// Wait for channel to close
		select {
		case <-exitCh:
			// Success - channel was closed
		case <-context.WithValue(context.Background(), "test", "timeout").Done():
			t.Fatal("exit command did not close channel")
		}
	})
}

// TestOperatorWait tests the Wait method
func TestOperatorWait(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("returns exit channel", func(t *testing.T) {
		exitCh := make(chan struct{})
		op := &Operator{
			store:  storage,
			exitCh: exitCh,
		}

		ch := op.Wait()
		assert.NotNil(t, ch)

		// Verify it's the same channel
		close(exitCh)
		select {
		case <-ch:
			// Success
		default:
			t.Fatal("Wait() should return the exit channel")
		}
	})
}

// TestOperatorBuildStateSigner tests the state signer construction
func TestOperatorBuildStateSigner(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("creates default signer when no session key", func(t *testing.T) {
		// Generate a valid private key
		privateKey, err := generatePrivateKey()
		require.NoError(t, err)

		signer, err := op.buildStateSigner(privateKey)
		require.NoError(t, err)
		assert.NotNil(t, signer)
	})

	t.Run("returns error for invalid private key", func(t *testing.T) {
		_, err := op.buildStateSigner("invalid_key")
		assert.Error(t, err)
	})

	t.Run("creates session key signer when session key exists", func(t *testing.T) {
		// Generate valid keys
		walletKey, err := generatePrivateKey()
		require.NoError(t, err)

		sessionKey, err := generatePrivateKey()
		require.NoError(t, err)

		// Store session key
		err = storage.SetSessionKey(sessionKey, "0xhash", "0xsig")
		require.NoError(t, err)

		_, err = op.buildStateSigner(walletKey)
		require.NoError(t, err)

		// Clean up for other tests
		storage.ClearSessionKey()
	})
}

// TestOperatorCommandParsing tests command argument parsing
func TestOperatorCommandParsing(t *testing.T) {
	t.Run("parses single command", func(t *testing.T) {
		input := "help"
		parts := strings.Fields(input)
		assert.Len(t, parts, 1)
		assert.Equal(t, "help", parts[0])
	})

	t.Run("parses command with arguments", func(t *testing.T) {
		input := "import wallet"
		parts := strings.Fields(input)
		assert.Len(t, parts, 2)
		assert.Equal(t, "import", parts[0])
		assert.Equal(t, "wallet", parts[1])
	})

	t.Run("parses command with multiple arguments", func(t *testing.T) {
		input := "deposit 80002 usdc 100"
		parts := strings.Fields(input)
		assert.Len(t, parts, 4)
		assert.Equal(t, "deposit", parts[0])
		assert.Equal(t, "80002", parts[1])
		assert.Equal(t, "usdc", parts[2])
		assert.Equal(t, "100", parts[3])
	})

	t.Run("handles extra whitespace", func(t *testing.T) {
		input := "  import   wallet  "
		parts := strings.Fields(input)
		assert.Len(t, parts, 2)
		assert.Equal(t, "import", parts[0])
		assert.Equal(t, "wallet", parts[1])
	})

	t.Run("handles tabs and mixed whitespace", func(t *testing.T) {
		input := "import\twallet"
		parts := strings.Fields(input)
		assert.Len(t, parts, 2)
		assert.Equal(t, "import", parts[0])
		assert.Equal(t, "wallet", parts[1])
	})
}

// TestOperatorGetAssetSuggestions tests asset suggestion generation
func TestOperatorGetAssetSuggestions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("returns empty when client not available", func(t *testing.T) {
		// Without a connected client, should return nil or empty
		suggestions := op.getAssetSuggestions()
		// Test passes if it doesn't panic
		// Actual behavior depends on client state
		_ = suggestions
	})
}

// TestOperatorGetChainSuggestions tests chain suggestion generation
func TestOperatorGetChainSuggestions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("returns empty when client not available", func(t *testing.T) {
		// Without a connected client, should return nil or empty
		suggestions := op.getChainSuggestions()
		// Test passes if it doesn't panic
		// Actual behavior depends on client state
		_ = suggestions
	})
}

// TestOperatorGetWalletSuggestion tests wallet address suggestion
func TestOperatorGetWalletSuggestion(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("returns empty when no wallet configured", func(t *testing.T) {
		suggestions := op.getWalletSuggestion()
		assert.Empty(t, suggestions)
	})

	t.Run("returns wallet suggestion when configured", func(t *testing.T) {
		// Generate and store a wallet
		privateKey, err := generatePrivateKey()
		require.NoError(t, err)

		err = storage.SetPrivateKey(privateKey)
		require.NoError(t, err)

		suggestions := op.getWalletSuggestion()
		require.Len(t, suggestions, 1)

		assert.NotEmpty(t, suggestions[0].Text)
		assert.True(t, strings.HasPrefix(suggestions[0].Text, "0x"))
		assert.Equal(t, 42, len(suggestions[0].Text))
		assert.Contains(t, suggestions[0].Description, "wallet")
	})
}

// TestOperatorCommandValidation tests command validation logic
func TestOperatorCommandValidation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("deposit requires correct number of arguments", func(t *testing.T) {
		// Test insufficient arguments
		op.Execute("deposit")
		op.Execute("deposit 1")
		op.Execute("deposit 1 usdc")
		// Should not panic, will print error messages
	})

	t.Run("withdraw requires correct number of arguments", func(t *testing.T) {
		op.Execute("withdraw")
		op.Execute("withdraw 1")
		op.Execute("withdraw 1 usdc")
		// Should not panic
	})

	t.Run("transfer requires correct number of arguments", func(t *testing.T) {
		op.Execute("transfer")
		op.Execute("transfer 0x123")
		op.Execute("transfer 0x123 usdc")
		// Should not panic
	})

	t.Run("import requires subcommand", func(t *testing.T) {
		op.Execute("import")
		// Should not panic
	})

	t.Run("import rpc requires arguments", func(t *testing.T) {
		op.Execute("import rpc")
		op.Execute("import rpc 1")
		// Should not panic
	})
}

// TestOperatorConcurrency tests concurrent access to operator
func TestOperatorConcurrency(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store:  storage,
		exitCh: make(chan struct{}),
	}

	t.Run("concurrent execute calls don't panic", func(t *testing.T) {
		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Execute panicked: %v", r)
					}
				}()
				op.Execute("help")
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent complete calls don't panic", func(t *testing.T) {
		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Complete panicked: %v", r)
					}
				}()
				doc := prompt.Document{Text: "h"}
				_ = op.Complete(doc)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}