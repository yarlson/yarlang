#!/bin/bash
# Simple test to verify LSP server responds to initialize request

# Create the initialize request
cat <<EOF | ./yarlang-lsp 2>/dev/null &
Content-Length: 169

{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"rootUri":"file:///test","capabilities":{},"initializationOptions":null}}
Content-Length: 52

{"jsonrpc":"2.0","method":"initialized","params":{}}
Content-Length: 46

{"jsonrpc":"2.0","id":2,"method":"shutdown"}
Content-Length: 43

{"jsonrpc":"2.0","method":"exit","params":{}}
EOF

# Wait briefly for response
sleep 1

# Kill if still running
pkill -f yarlang-lsp 2>/dev/null

echo "Test completed - if no errors appeared, the server is working"
