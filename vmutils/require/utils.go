// From https://github.com/luau-lang/luau/blob/master/CLI/src/FileUtils.cpp
//
// Unix-like path normalization and absolute path checking
package require

import (
	"strings"
	"unicode"
)

func splitPath(path string) []string {
	var components []string

	pos := 0
	nextPos := strings.IndexAny(path[pos:], "\\/")

	for nextPos != -1 {
		components = append(components, path[pos:pos+nextPos])
		pos += nextPos + 1
		nextPos = strings.IndexAny(path[pos:], "\\/")
	}
	components = append(components, path[pos:])

	return components
}

func unixisAbsolutePath(path string) bool {
	if len(path) >= 3 && unicode.IsLetter(rune(path[0])) && path[1] == ':' && (path[2] == '/' || path[2] == '\\') {
		return true
	}
	return len(path) >= 1 && (path[0] == '/' || path[0] == '\\')
}

func unixnormalizePath(path string) string {
	components := splitPath(path)
	var normalizedComponents []string

	isAbsolute := unixisAbsolutePath(path)

	// 1. Normalize path components
	startIndex := 0
	if isAbsolute {
		startIndex = 1
	}
	for i := startIndex; i < len(components); i++ {
		component := components[i]
		if component == ".." {
			if len(normalizedComponents) == 0 {
				if !isAbsolute {
					normalizedComponents = append(normalizedComponents, "..")
				}
			} else if normalizedComponents[len(normalizedComponents)-1] == ".." {
				normalizedComponents = append(normalizedComponents, "..")
			} else {
				normalizedComponents = normalizedComponents[:len(normalizedComponents)-1]
			}
		} else if component != "" && component != "." {
			normalizedComponents = append(normalizedComponents, component)
		}
	}

	var normalizedPath strings.Builder

	// 2. Add correct prefix to formatted path
	if isAbsolute {
		normalizedPath.WriteString(components[0])
		normalizedPath.WriteString("/")
	} else if len(normalizedComponents) == 0 || normalizedComponents[0] != ".." {
		normalizedPath.WriteString("./")
	}

	// 3. Join path components to form the normalized path
	for i, component := range normalizedComponents {
		if i != 0 {
			normalizedPath.WriteString("/")
		}
		normalizedPath.WriteString(component)
	}
	if len(normalizedPath.String()) >= 2 && normalizedPath.String()[len(normalizedPath.String())-1] == '.' && normalizedPath.String()[len(normalizedPath.String())-2] == '.' {
		normalizedPath.WriteString("/")
	}

	return normalizedPath.String()
}
