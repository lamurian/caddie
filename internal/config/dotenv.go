package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ReadDotEnv parses a .env file into a map. A missing file is not an error
// and yields an empty map. Errors name the file and the line.
func ReadDotEnv(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfig, err)
	}
	return ParseDotEnv(path, data)
}

// ParseDotEnv parses .env content: KEY=VALUE lines, optional `export `,
// single or double quotes, full-line comments and ` # inline` comments.
func ParseDotEnv(name string, data []byte) (map[string]string, error) {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; sc.Scan(); n++ {
		key, val, ok, err := parseDotEnvLine(sc.Text())
		if err != nil {
			return nil, fmt.Errorf("%w: %s:%d: %s", ErrConfig, name, n, err)
		}
		if ok {
			out[key] = val
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrConfig, name, err)
	}
	return out, nil
}

func parseDotEnvLine(line string) (key, val string, ok bool, err error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}
	line = strings.TrimPrefix(line, "export ")
	key, raw, found := strings.Cut(line, "=")
	key = strings.TrimSpace(key)
	if !found {
		return "", "", false, fmt.Errorf("expected KEY=VALUE, got %q", line)
	}
	if !keyPattern.MatchString(key) {
		return "", "", false, fmt.Errorf("invalid key %q", key)
	}
	val, err = parseDotEnvValue(strings.TrimSpace(raw))
	return key, val, err == nil, err
}

func parseDotEnvValue(raw string) (string, error) {
	if raw == "" || raw[0] == '#' {
		return "", nil // a value starting with # must be quoted
	}
	if q := raw[0]; q == '"' || q == '\'' {
		end := strings.IndexByte(raw[1:], q)
		if end < 0 {
			return "", fmt.Errorf("unterminated %c quote", q)
		}
		rest := strings.TrimSpace(raw[end+2:])
		if rest != "" && !strings.HasPrefix(rest, "#") {
			return "", fmt.Errorf("unexpected text after quoted value: %q", rest)
		}
		return raw[1 : end+1], nil
	}
	if i := strings.Index(raw, " #"); i >= 0 {
		raw = raw[:i]
	}
	if i := strings.Index(raw, "\t#"); i >= 0 {
		raw = raw[:i]
	}
	return strings.TrimSpace(raw), nil
}
