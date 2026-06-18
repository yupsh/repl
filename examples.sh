#!/bin/bash
#
# Example commands for the yupsh REPL. Run them interactively, or pipe this
# file in:
#
#   ./yupsh/yupsh < examples.sh
#

# Emit a line (echo is a source command)
echo "Hello from yupsh REPL"

# Generate a sequence of numbers (seq is a source command)
seq 1 10

# Pipeline: generate numbers, drop one, keep the first three
seq 1 10 | grep -v 5 | head -n 3

# Count lines
seq 1 100 | wc -l

# Lowercase text
echo "HELLO WORLD" | tr A-Z a-z

# Reverse each line
echo hello | rev

# Reverse line order
seq 1 5 | tac

# Sort and de-duplicate
emit "banana" | sort | uniq

# Base64 round-trip
echo "secret message" | base64
echo "c2VjcmV0IG1lc3NhZ2U=" | base64 -d

# Number lines
seq 1 4 | nl

# Select a field
echo one,two,three | cut -d , -f 2

# Path helpers
basename /path/to/some/file.txt
dirname /path/to/some/file.txt

# Leave the REPL
exit
