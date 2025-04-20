package lsp

import "os"

func handleInitialize(request Request) {
	result := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"textDocumentSync": 1,
			"completionProvider": map[string]interface{}{
				"resolveProvider":   false,
				"triggerCharacters": []string{"|", "=", " "},
			},
			"renameProvider":     true,
			"definitionProvider": true,
		},
	}
	sendResponse(request.ID, result)
}

func handleInitialized(request Request) {
	// No response required, no actions taken currently.
}

func handleShutdown(request Request) {
	shutdownRequested = true
	sendResponse(request.ID, nil)
}

func handleExit() {
	if shutdownRequested {
		os.Exit(0)
	}
	os.Exit(1)
}
