package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	Language       string
	Port           string
	RootDir        string
	PackageManager string
}

func DetectProject(dir string) Project {
	p := Project{RootDir: dir}

	if hasFile(dir, "package.json") {
		p.Language = "Node.js"
		if hasFile(dir, "next.config.js") || hasFile(dir, "next.config.ts") || hasDir(dir, "pages") || hasDir(dir, "app") || hasDir(dir, ".next") {
			p.Language = "Next.js"
		}

		// Detect Package Manager
		if hasFile(dir, "pnpm-lock.yaml") {
			p.PackageManager = "pnpm"
		} else if hasFile(dir, "yarn.lock") {
			p.PackageManager = "yarn"
		} else {
			p.PackageManager = "npm"
		}
	} else if hasFile(dir, "go.mod") {
		p.Language = "Go"
	}

	p.Port = extractPort(dir, p.Language)
	return p
}

func extractPort(dir, lang string) string {
	// 1. Check .env variants
	envFiles := []string{".env", ".env.local", ".env.example"}
	for _, f := range envFiles {
		if content, err := os.ReadFile(filepath.Join(dir, f)); err == nil {
			re := regexp.MustCompile(`(?i)PORT\s*[=:]\s*(\d+)`)
			matches := re.FindStringSubmatch(string(content))
			if len(matches) >= 2 {
				return matches[1]
			}
		}
	}

	// 2. For Node/Next, check package.json scripts
	if lang == "Node.js" || lang == "Next.js" {
		if content, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(content, &pkg); err == nil {
				if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
					re := regexp.MustCompile(`-p\s+(\d+)`)
					for _, cmd := range scripts {
						if s, ok := cmd.(string); ok {
							matches := re.FindStringSubmatch(s)
							if len(matches) >= 2 {
								return matches[1]
							}
						}
					}
				}
			}
		}
	}

	// 3. For Go, scan source files
	if lang == "Go" {
		foundPort := ""
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
				return nil
			}
			if strings.Contains(path, "vendor") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			re := regexp.MustCompile(`(?i)(?:ListenAndServe|Run|Start)\s*\(\s*".*?:(\d{4,5})"`)
			matches := re.FindStringSubmatch(string(content))
			if len(matches) >= 2 {
				foundPort = matches[1]
				return fmt.Errorf("found") // Stop walking
			}
			return nil
		})
		if foundPort != "" {
			return foundPort
		}
	}

	// Default ports
	switch lang {
	case "Node.js", "Next.js":
		return "3000"
	case "Go":
		return "8080"
	}

	return "8080"
}

func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}

func hasDir(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && info.IsDir()
}
