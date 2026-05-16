package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	Language string
	Port     string
	RootDir  string
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "start":
		runStart()
	case "help", "-h", "--help":
		showHelp()
	default:
		fmt.Printf("❌ Command tidak dikenal: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("🚀 GoPipe - Auto CI/CD Generator")
	fmt.Println("Gue bakal bantu lu nge-generate workflow GitHub Actions secara otomatis.")
	fmt.Println("\nUsage:")
	fmt.Println("  gopipe start    Analisa project dan generate deploy.yml")
	fmt.Println("  gopipe help     Tampilkan bantuan ini")
}

func runStart() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Gagal baca direktori: %v\n", err)
		os.Exit(1)
	}

	proj := detectProject(dir)
	if proj.Language == "" {
		fmt.Println("❌ Tidak terdeteksi project yang dikenal (Node/Next/Go).")
		os.Exit(0)
	}

	fmt.Printf("🚀 Project terdeteksi: %s\n", proj.Language)
	fmt.Printf("🔌 Port: %s\n", proj.Port)

	if err := generateGithubActions(proj); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Gagal generate CI/CD: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ CI/CD workflow berhasil digenerate di .github/workflows/deploy.yml")

	// --- GitHub Integration ---
	fmt.Print("\n🌐 Hubungkan project ini ke GitHub? (y/n): ")
	var connect string
	fmt.Scanln(&connect)
	if strings.ToLower(connect) != "y" {
		fmt.Println("Siaap, CI/CD udah siap lokal aja. Mantap, bre! 🔥")
		return
	}

	// 1. Check GH CLI
	if err := exec.Command("gh", "--version").Run(); err != nil {
		fmt.Println("❌ GitHub CLI (gh) tidak ditemukan. Install dulu ya biar makin sakti!")
		return
	}

	// 2. Check Auth
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		fmt.Println("🔑 Kamu belum login ke GitHub. Silakan jalankan: gh auth login")
		return
	}

	// 3. Ask for repo status
	fmt.Print("📂 Udah punya repo di GitHub buat project ini? (y/n): ")
	var hasRepo string
	fmt.Scanln(&hasRepo)

	var repoName string
	if strings.ToLower(hasRepo) == "y" {
		fmt.Print("🔗 Masukkan nama repo (misal: username/repo-name): ")
		fmt.Scanln(&repoName)

		fmt.Println("🔗 Menghubungkan ke remote origin...")
		exec.Command("git", "init").Run()
		exec.Command("git", "remote", "add", "origin", "https://github.com/"+repoName+".git").Run()
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "chore: setup gopipe ci/cd").Run()

		fmt.Println("🚀 Pushing to GitHub...")
		cmd := exec.Command("git", "push", "-u", "origin", "main")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Fallback to master if main fails
			exec.Command("git", "push", "-u", "origin", "master").Run()
		}
	} else {
		fmt.Print("🆕 Masukkan nama repo baru yang mau dibuat: ")
		fmt.Scanln(&repoName)

		fmt.Println("🔨 Menyiapkan git lokal...")
		exec.Command("git", "init").Run()
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "chore: setup gopipe ci/cd").Run()

		fmt.Println("🔨 Membuat repository baru di GitHub dan pushing...")
		// Use gh repo create with source and push
		cmd := exec.Command("gh", "repo", "create", repoName, "--public", "--source=.", "--push")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Gagal membuat repo: %v\n", err)
			return
		}
	}

	fmt.Println("\n✨ Mantap, bre! Project lu udah aman di GitHub dengan CI/CD yang nyala! 🔥")
}

func detectProject(dir string) Project {
	p := Project{RootDir: dir}

	if hasFile(dir, "package.json") {
		p.Language = "Node.js"
		if hasFile(dir, "next.config.js") || hasFile(dir, "next.config.ts") || hasDir(dir, "pages") || hasDir(dir, "app") || hasDir(dir, ".next") {
			p.Language = "Next.js"
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

func generateGithubActions(proj Project) error {
	workDir := filepath.Join(proj.RootDir, ".github", "workflows")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("gagal buat direktori: %w", err)
	}

	action := buildWorkflow(proj)
	path := filepath.Join(workDir, "deploy.yml")
	return os.WriteFile(path, []byte(action), 0644)
}

func buildWorkflow(proj Project) string {
	var buildSteps string
	var deploySteps string

	switch proj.Language {
	case "Go":
		buildSteps = `      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: go build -o app main.go`

		deploySteps = `      - name: Deploy to Server
        uses: appleboy/scp-action@v0.1.7
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          source: "app"
          target: "~/apps/gopipe"

      - name: Restart Service
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd ~/apps/gopipe
            chmod +x app
            sudo systemctl restart gopipe || true
            nohup ./app > app.log 2>&1 &`

	case "Next.js":
		buildSteps = `      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Export static
        if: hashFiles('next.config.js') != ''
        run: npm run export || true`

		deploySteps = `      - name: Deploy via Rsync
        uses: burnett01/rsync-deployments@7.0.2
        with:
          switches: -avzr --delete
          path: .next/
          remote_path: ~/apps/gopipe
          remote_host: ${{ secrets.SSH_HOST }}
          remote_user: ${{ secrets.SSH_USER }}
          remote_key: ${{ secrets.SSH_KEY }}

      - name: Restart PM2
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd ~/apps/gopipe
            npm run start -- -p $PORT &
            pm2 restart gopipe || pm2 start npm --name gopipe -- start`

	case "Node.js":
		buildSteps = `      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build || true`

		deploySteps = `      - name: Deploy via Rsync
        uses: burnett01/rsync-deployments@7.0.2
        with:
          switches: -avzr --delete
          path: ./ --exclude node_modules
          remote_path: ~/apps/gopipe
          remote_host: ${{ secrets.SSH_HOST }}
          remote_user: ${{ secrets.SSH_USER }}
          remote_key: ${{ secrets.SSH_KEY }}

      - name: Restart App
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd ~/apps/gopipe
            npm install
            pm2 restart gopipe || pm2 start npm --name gopipe -- start`
	}

	return fmt.Sprintf(`name: Deploy

on:
  push:
    branches: [ main ]

env:
  PORT: %s

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

%s

%s
`, proj.Port, buildSteps, deploySteps)
}

func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}

func hasDir(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && info.IsDir()
}
