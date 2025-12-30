#!/bin/bash
# Test script for Ghost output formats
# Usage: ./test_formats.sh

set -e

echo "Testing Ghost output formats"
echo "================================================"
echo ""

echo "1. Testing normal output:"
echo "------------------------"
echo "$ go run . \"say hello\""
go run . "say hello"
echo ""

echo "2. Testing JSON output:"
echo "------------------------"
echo "$ go run . -f json \"list 3 colors\""
go run . -f json "list 3 colors"
echo ""

echo "3. Testing markdown output:"
echo "------------------------"
echo "$ go run . -f markdown \"write markdown showing all these elements: heading 1, heading 2, bold, italic, code block with go code, inline code, link, list, blockquote, horizontal rule\""
go run . -f markdown "write markdown showing all these elements: heading 1, heading 2, bold, italic, code block with go code, inline code, link, list, blockquote, horizontal rule"
echo ""

echo "================================================"
echo "All format tests completed!"
