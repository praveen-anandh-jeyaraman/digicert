package handler

import (
    "strings"
)

type ValidationErrors map[string]string

func trim(s string) string {
    return strings.TrimSpace(s)
}
