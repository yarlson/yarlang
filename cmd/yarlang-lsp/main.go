package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/yarlson/yarlang/server"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// stdinStdout wraps stdin and stdout into a single ReadWriteCloser
type stdinStdout struct {
	io.Reader
	io.Writer
}

func (s stdinStdout) Close() error {
	// stdin/stdout don't need explicit closing
	return nil
}

func main() {
	// Redirect logs to file to avoid corrupting LSP protocol stream on stdout
	logFile, err := os.OpenFile("/tmp/yarlang-lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)

		defer func() {
			if err := logFile.Close(); err != nil {
				log.Printf("Failed to close log file: %v", err)
			}
		}()
	}

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	// LSP communicates over stdin/stdout
	// Combine stdin/stdout into a single ReadWriteCloser
	rwc := stdinStdout{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(rwc))

	srv := server.New()

	// Set up diagnostic callback to publish via LSP
	srv.DiagnosticCallback = func(uri string, diagnostics []protocol.Diagnostic) {
		if err := conn.Notify(context.Background(), "textDocument/publishDiagnostics", &protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentURI(uri),
			Diagnostics: diagnostics,
		}); err != nil {
			log.Printf("Failed to publish diagnostics: %v", err)
		}
	}

	// Handle incoming requests
	handler := protocol.ServerHandler(srv, nil)

	ctx := context.Background()
	conn.Go(ctx, handler)

	// Wait for connection to close
	<-conn.Done()

	if err := conn.Err(); err != nil {
		log.Printf("Connection error: %v", err)
		os.Exit(1)
	}
}
