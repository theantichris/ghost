#!/bin/bash
# Test script for Ghost features
# Usage: ./test_features.sh

set -e

echo "Testing Ghost features"
echo "================================================"
echo ""

echo "1. Testing normal output:"
echo "------------------------"
echo "$ go run . \"say hello\""
echo ""
go run . "say hello"
echo ""

echo "2. Testing piping content in:"
echo "------------------------"
echo "$ echo \"package main\" | go run . \"explain this code\""
echo ""
echo "package main" | go run . "explain this code"
echo ""

echo "3. Testing piping content out:"
echo "------------------------"
echo "$ go run . \"say hello\" | cat"
echo ""
go run . "say hello" | cat
echo ""

echo "4. Testing JSON output:"
echo "------------------------"
echo "$ go run . -f json \"list 3 colors\""
echo ""
go run . -f json "list 3 colors"
echo ""

echo "5. Testing markdown output:"
echo "------------------------"
echo "$ go run . -f markdown \"write markdown showing all these elements: heading 1, heading 2, bold, italic, code block with go code, inline code, link, list, blockquote, horizontal rule\""
echo ""
go run . -f markdown "write markdown showing all these elements: heading 1, heading 2, bold, italic, code block with go code, inline code, link, list, blockquote, horizontal rule"
echo ""

echo "================================================"
echo "All feature tests completed!"
