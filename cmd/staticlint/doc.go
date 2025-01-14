// Package staticlint is a custom linter for Go code
// Contains the following checks:
//   - checks for shadowed variables.
//   - detects common mistakes involving boolean operators
//   - checks for common mistakes in defer statements
//   - inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons.
//   - various misuses of the standard library
//   - stylistic issues
//   - code simplifications
//   - quickfixes
//   - check for direct call os.Exit in main functions
//
// Usage:
// ./staticlint ./...
package main
