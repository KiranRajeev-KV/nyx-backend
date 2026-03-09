#!/bin/sh

if grep -nE 'fmt\.Print(ln|f)' "$@" | grep -v "main.go" ; then
  echo "Error: Found 'fmt.Println' or 'fmt.Printf' in staged Go files."
  echo "Please use structured logging (zerolog) instead."
  exit 1
fi
