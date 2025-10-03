package types

import (
    "fmt"
    "strings"
)

type StringSlice []string

func (s *StringSlice) String() string {
    return fmt.Sprintf("%v", *s)
}

func (s *StringSlice) Set(value string) error {
    // Разбиваем по запятой
    parts := strings.Split(value, ",")
    for _, p := range parts {
        trimmed := strings.TrimSpace(p)
        if trimmed != "" {
            *s = append(*s, trimmed)
        }
    }
    return nil
}

func (s *StringSlice) Get() interface{} {
    return *s
}
