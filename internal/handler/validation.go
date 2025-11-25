package handler

import (
    "strings"
)

type ValidationErrors map[string]string

// trim string safely
func trim(s string) string {
    return strings.TrimSpace(s)
}
