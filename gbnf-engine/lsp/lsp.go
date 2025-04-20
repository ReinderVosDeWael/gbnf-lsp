package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var debugLogger *log.Logger

func Run() {
	reader := bufio.NewReader(os.Stdin)
	debugLogger = log.New(os.Stderr, "LSP: ", log.Ltime|log.Lshortfile)
	for {
		debugLogger.Printf("Number of open files: %v", len(openFiles))
		var contentLength int
		for {
			header, err := reader.ReadString('\n')
			debugLogger.Print(header)
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
				return
			}

			header = strings.TrimSpace(header)
			if header == "" {
				break
			}

			if strings.HasPrefix(header, "Content-Length:") {
				value := strings.TrimSpace(strings.TrimPrefix(header, "Content-Length:"))
				contentLength, err = strconv.Atoi(value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid Content-Length: %v\n", err)
					continue
				}
			}
		}

		if contentLength == 0 {
			fmt.Fprintf(os.Stderr, "Did not find a Content-Length\n")
			continue
		}

		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)

		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintf(os.Stderr, "Failed to read message body: %v\n", err)
			continue
		}
		var request Request

		err = json.Unmarshal(body, &request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON %v\n", err)
			continue
		}

		handleRequest(request)
	}
}

func handleRequest(request Request) {
	debugLogger.Printf("Started request %v with method %v", request.ID, request.Method)
	switch request.Method {
	case "initialize":
		handleInitialize(request)
	case "initialized":
		handleInitialized(request)
	case "shutdown":
		handleShutdown(request)
	case "exit":
		handleExit()
	case "textDocument/didOpen":
		handleTextDocumentDidOpen(request)
	case "textDocument/didChange":
		handleTextDocumentDidChange(request)
	case "textDocument/didSave":
		handleTextDocumentDidSave(request)
	case "textDocument/didClose":
		handleTextDocumentDidClose(request)
	case "textDocument/completion":
		handleTextDocumentCompletion(request)
	case "textDocument/rename":
		handleTextDocumentRename(request)
	case "textDocument/definition":
		handleTextDocumentDefinition(request)

	default:
		sendError(request.ID, -32601, "Method not found.")
	}
	debugLogger.Printf("Finished request %v", request.ID)
}
