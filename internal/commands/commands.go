package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopipe/internal/config"
	"gopipe/internal/detector"
	"gopipe/internal/github"
	"gopipe/internal/workflow"
)

func Execute() {
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

	proj := detector.DetectProject(dir)
	if proj.Language == "" {
		fmt.Println("❌ Tidak terdeteksi project yang dikenal (Node/Next/Go).")
		os.Exit(0)
	}

	fmt.Printf("🚀 Project terdeteksi: %s\n", proj.Language)
	fmt.Printf("🔌 Port: %s\n", proj.Port)

	if err := workflow.GenerateGithubActions(proj); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Gagal generate CI/CD: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ CI/CD workflow berhasil digenerate di .github/workflows/deploy.yml")

	// --- GitHub Integration ---
	fmt.Print("\n🌐 Hubungkan project ini ke GitHub? (y/n): ")
	var connect string
	fmt.Scan(&connect)
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
	fmt.Scan(&hasRepo)

	var repoName string
	if strings.ToLower(hasRepo) == "y" {
		fmt.Print("🔗 Masukkan nama repo (misal: username/repo-name): ")
		fmt.Scan(&repoName)

		fmt.Println("🔗 Menghubungkan ke remote origin...")
		exec.Command("git", "init").Run()
		exec.Command("git", "remote", "add", "origin", "https://github.com/"+repoName+".git").Run()
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "chore: setup gopipe ci/cd").Run()

		// Detect current branch
		branchCmd := exec.Command("git", "branch", "--show-current")
		branchOut, _ := branchCmd.Output()
		currentBranch := strings.TrimSpace(string(branchOut))
		if currentBranch == "" {
			currentBranch = "main" // Fallback
		}

		fmt.Printf("🚀 Pushing branch '%s' to GitHub...\n", currentBranch)
		cmd := exec.Command("git", "push", "-u", "origin", currentBranch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Gagal push ke GitHub: %v\n", err)
			return
		}
	} else {
		fmt.Print("🆕 Masukkan nama repo baru yang mau dibuat: ")
		fmt.Scan(&repoName)

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

	fmt.Print("\n🖥️  Setup CI/CD ke VPS sekarang? (y/n): ")
	var setupVPS string
	fmt.Scan(&setupVPS)
	if strings.ToLower(setupVPS) != "y" {
		fmt.Println("\n✨ Mantap, bre! Project lu udah aman di GitHub dengan CI/CD yang nyala! 🔥")
		return
	}

	var vpsIP, vpsUser, keyPath string
	appConfig, err := config.LoadConfig()
	useSaved := "n"

	if err == nil {
		fmt.Printf("📦 Nemu config VPS tersimpan: %s@%s\n", appConfig.VpsUser, appConfig.VpsIP)
		fmt.Print("🖥️  Gunakan config ini, bre? (y/n): ")
		fmt.Scan(&useSaved)
	}

	if strings.ToLower(useSaved) == "y" {
		vpsIP = appConfig.VpsIP
		vpsUser = appConfig.VpsUser
		keyPath = appConfig.VpsKeyPath
	} else {
		fmt.Print("🌐 Masukkan IP VPS: ")
		fmt.Scan(&vpsIP)

		fmt.Print("👤 Masukkan Username VPS: ")
		fmt.Scan(&vpsUser)

		fmt.Print("🔑 Masukkan path ke SSH Private Key (misal: /home/user/.ssh/id_rsa): ")
		fmt.Scan(&keyPath)

		// Save new config
		config.SaveConfig(config.AppConfig{VpsIP: vpsIP, VpsUser: vpsUser, VpsKeyPath: keyPath})
	}

	// Resolve path (handle ~ if any)
	resolvedPath := keyPath
	if strings.HasPrefix(resolvedPath, "~") {
		resolvedPath = filepath.Join(os.Getenv("HOME"), resolvedPath[1:])
	}

	keyContent, err := os.ReadFile(resolvedPath)
	if err != nil {
		fmt.Printf("❌ Gagal baca SSH Key di %s: %v\n", resolvedPath, err)
		return
	}

	fmt.Println("🔐 Mengupload secrets ke GitHub...")
	github.SetSecret("SSH_HOST", vpsIP)
	github.SetSecret("SSH_USER", vpsUser)
	github.SetSecret("SSH_KEY", string(keyContent))

	fmt.Println("\n✨ SEMUA SET! Sekarang tiap lu push, project lu bakal otomatis deploy ke VPS! Mantap, bre! 🔥🚀")
}
