package server

import (
	"context"
	"fmt"

	"github.com/yarlson/yarlang/analysis"
	"github.com/yarlson/yarlang/ast"
	"go.lsp.dev/protocol"
)

// Server implements the LSP server
type Server struct {
	documents          map[string]*Document
	DiagnosticCallback func(uri string, diagnostics []protocol.Diagnostic)
}

// New creates a new LSP server
func New() *Server {
	return &Server{
		documents: make(map[string]*Document),
	}
}

// Initialize handles the initialize request
func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{},
			},
			HoverProvider:      true,
			DefinitionProvider: true,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "yarlang-lsp",
			Version: "0.1.0",
		},
	}, nil
}

// DidOpen handles textDocument/didOpen notification
func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)

	doc := &Document{
		URI:     uri,
		Version: int(params.TextDocument.Version),
		Content: params.TextDocument.Text,
	}

	doc.Parse()
	doc.Analyze()

	s.documents[uri] = doc

	// Publish diagnostics
	s.publishDiagnostics(uri, doc)

	return nil
}

// DidChange handles textDocument/didChange notification
func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := string(params.TextDocument.URI)

	doc, ok := s.documents[uri]
	if !ok {
		return fmt.Errorf("document not found: %s", uri)
	}

	// Full sync - replace entire content
	if len(params.ContentChanges) > 0 {
		doc.Update(params.ContentChanges[0].Text, int(params.TextDocument.Version))
		s.publishDiagnostics(uri, doc)
	}

	return nil
}

// DidClose handles textDocument/didClose notification
func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	delete(s.documents, uri)

	return nil
}

// publishDiagnostics converts and publishes diagnostics
func (s *Server) publishDiagnostics(uri string, doc *Document) {
	if s.DiagnosticCallback == nil {
		return
	}

	diagnostics := []protocol.Diagnostic{}

	for _, diag := range doc.Diagnostics {
		severity := protocol.DiagnosticSeverityError

		switch diag.Severity {
		case analysis.SeverityWarning:
			severity = protocol.DiagnosticSeverityWarning
		case analysis.SeverityInfo:
			severity = protocol.DiagnosticSeverityInformation
		case analysis.SeverityHint:
			severity = protocol.DiagnosticSeverityHint
		}

		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(diag.Range.Start.Line - 1),
					Character: uint32(diag.Range.Start.Column - 1),
				},
				End: protocol.Position{
					Line:      uint32(diag.Range.End.Line - 1),
					Character: uint32(diag.Range.End.Column - 1),
				},
			},
			Severity: severity,
			Message:  diag.Message,
			Source:   "yarlang-lsp",
		})
	}

	s.DiagnosticCallback(uri, diagnostics)
}

// Completion handles textDocument/completion request
func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	uri := string(params.TextDocument.URI)

	doc, ok := s.documents[uri]
	if !ok || doc.Symbols == nil {
		return &protocol.CompletionList{Items: []protocol.CompletionItem{}}, nil
	}

	// Convert LSP position to AST position (LSP is 0-based, AST is 1-based)
	pos := ast.Position{
		Line:   int(params.Position.Line) + 1,
		Column: int(params.Position.Character) + 1,
	}

	// Find scope at cursor position
	scope := doc.Symbols.ScopeAt(pos)
	if scope == nil {
		// Fallback to global scope
		if len(doc.Symbols.Scopes()) > 0 {
			scope = doc.Symbols.Scopes()[0]
		} else {
			return &protocol.CompletionList{Items: []protocol.CompletionItem{}}, nil
		}
	}

	items := []protocol.CompletionItem{}

	// Add all visible symbols
	symbols := scope.AllSymbols()
	for _, sym := range symbols {
		kind := protocol.CompletionItemKindVariable

		switch sym.Kind {
		case analysis.SymbolKindFunction:
			kind = protocol.CompletionItemKindFunction
		case analysis.SymbolKindParameter:
			kind = protocol.CompletionItemKindVariable
		}

		items = append(items, protocol.CompletionItem{
			Label:  sym.Name,
			Kind:   kind,
			Detail: sym.Type,
		})
	}

	// Add keywords
	keywords := []string{"func", "if", "else", "for", "return", "break", "continue", "nil", "true", "false"}
	for _, kw := range keywords {
		items = append(items, protocol.CompletionItem{
			Label: kw,
			Kind:  protocol.CompletionItemKindKeyword,
		})
	}

	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// Hover handles textDocument/hover request
func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := string(params.TextDocument.URI)

	doc, ok := s.documents[uri]
	if !ok || doc.Symbols == nil {
		return nil, nil
	}

	pos := ast.Position{
		Line:   int(params.Position.Line) + 1,
		Column: int(params.Position.Character) + 1,
	}

	symbol := doc.Symbols.FindSymbolAt(pos)
	if symbol == nil {
		return nil, nil
	}

	// Format hover content
	var content string

	switch symbol.Kind {
	case analysis.SymbolKindFunction:
		// TODO: Format function signature with parameters
		content = fmt.Sprintf("```yar\nfunc %s\n```\n\nFunction", symbol.Name)
	case analysis.SymbolKindVariable:
		content = fmt.Sprintf("```yar\n%s: %s\n```\n\nVariable", symbol.Name, symbol.Type)
	case analysis.SymbolKindParameter:
		content = fmt.Sprintf("```yar\n%s: %s\n```\n\nParameter", symbol.Name, symbol.Type)
	default:
		content = symbol.Name
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: content,
		},
	}, nil
}

// Definition handles textDocument/definition request
func (s *Server) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	uri := string(params.TextDocument.URI)

	doc, ok := s.documents[uri]
	if !ok || doc.Symbols == nil {
		return nil, nil
	}

	pos := ast.Position{
		Line:   int(params.Position.Line) + 1,
		Column: int(params.Position.Character) + 1,
	}

	symbol := doc.Symbols.FindSymbolAt(pos)
	if symbol == nil {
		return nil, nil
	}

	// Return declaration location
	return []protocol.Location{{
		URI: protocol.DocumentURI(uri),
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(symbol.DeclRange.Start.Line - 1),
				Character: uint32(symbol.DeclRange.Start.Column - 1),
			},
			End: protocol.Position{
				Line:      uint32(symbol.DeclRange.End.Line - 1),
				Character: uint32(symbol.DeclRange.End.Column - 1),
			},
		},
	}}, nil
}

// Stub implementations for required protocol.Server interface methods

func (s *Server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Server) Exit(ctx context.Context) error {
	return nil
}

func (s *Server) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	return nil
}

func (s *Server) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	return nil
}

func (s *Server) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	return nil
}

func (s *Server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	return nil, nil
}

func (s *Server) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	return nil, nil
}

func (s *Server) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	return nil, nil
}

func (s *Server) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	return nil, nil
}

func (s *Server) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	return nil, nil
}

func (s *Server) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *Server) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	return nil
}

func (s *Server) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	return nil
}

func (s *Server) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	return nil
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (s *Server) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	return nil, nil
}

func (s *Server) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	return nil, nil
}

func (s *Server) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	return nil, nil
}

func (s *Server) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	return nil, nil
}

func (s *Server) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	return nil, nil
}

func (s *Server) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	return nil, nil
}

func (s *Server) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	return nil, nil
}

func (s *Server) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *Server) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *Server) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *Server) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	return nil, nil
}

func (s *Server) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *Server) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *Server) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	return nil, nil
}

func (s *Server) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	return nil, nil
}

func (s *Server) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	return nil, nil
}

func (s *Server) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *Server) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	return nil
}

func (s *Server) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *Server) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	return nil, nil
}

func (s *Server) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	return nil, nil
}

func (s *Server) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	return nil
}

func (s *Server) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	return nil, nil
}

func (s *Server) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	return nil
}

func (s *Server) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	return nil, nil
}

func (s *Server) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	return nil
}

func (s *Server) CodeLensRefresh(ctx context.Context) error {
	return nil
}

func (s *Server) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	return nil, nil
}

func (s *Server) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	return nil, nil
}

func (s *Server) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	return nil, nil
}

func (s *Server) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	return nil, nil
}

func (s *Server) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	return nil, nil
}

func (s *Server) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	return nil, nil
}

func (s *Server) SemanticTokensRefresh(ctx context.Context) error {
	return nil
}

func (s *Server) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	return nil, nil
}

func (s *Server) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	return nil, nil
}

func (s *Server) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	return nil, nil
}
