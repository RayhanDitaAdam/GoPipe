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

	"github.com/AlecAivazis/survey/v2"
)

const banner = `
  ____        ____  _             
 / ___| ___  |  _ \(_)_ __   ___ 
| |  _ / _ \ | |_) | | '_ \ / _ \
| |_| | (_) ||  __/| | |_) |  __/
 \____|\___/ |_|   |_| .__/ \___|
                     |_|         
`

func Execute() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "start":
		fmt.Print(banner)
		runStart()
	case "help", "-h", "--help":
		showHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("GoPipe - Auto CI/CD Generator")
	fmt.Println("I will help you generate GitHub Actions workflow automatically")
	fmt.Println("\nUsage:")
	fmt.Println("  gopipe start    Analyze project and generate deploy.yml")
	fmt.Println("  gopipe help     Show this help")
}

func runStart() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read directory: %v\n", err)
		os.Exit(1)
	}

	proj := detector.DetectProject(dir)
	if proj.Language == "" {
		fmt.Println("No supported project detected (Node/Next/Go)")
		os.Exit(0)
	}

	fmt.Printf("Project detected: %s\n", proj.Language)
	fmt.Printf("Port: %s\n", proj.Port)

	if err := workflow.GenerateGithubActions(proj); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate CI/CD: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("CI/CD workflow generated at .github/workflows/deploy.yml")

	// --- GitHub Integration ---
	var connect bool
	promptConnect := &survey.Confirm{
		Message: "Connect this project to GitHub?",
		Default: true,
	}
	survey.AskOne(promptConnect, &connect)

	if !connect {
		fmt.Println("CI/CD setup complete locally")
		return
	}

	// 1. Check GH CLI
	if err := exec.Command("gh", "--version").Run(); err != nil {
		fmt.Println("GitHub CLI (gh) not found Please install it first")
		return
	}

	// 2. Check Auth
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		var login bool
		promptLogin := &survey.Confirm{
			Message: "You are not logged in to GitHub Login now?",
			Default: true,
		}
		survey.AskOne(promptLogin, &login)

		if login {
			cmd := exec.Command("gh", "auth", "login")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Login failed: %v\n", err)
				return
			}
		} else {
			return
		}
	}

	// 3. Ask for repo status
	var hasRepo bool
	promptHasRepo := &survey.Confirm{
		Message: "Do you already have a repository on GitHub for this project?",
		Default: false,
	}
	survey.AskOne(promptHasRepo, &hasRepo)

	var repoName string
	if hasRepo {
		promptRepoName := &survey.Input{
			Message: "Enter repository name (e.g. username/repo-name):",
		}
		survey.AskOne(promptRepoName, &repoName)

		if repoName == "" {
			fmt.Println("Repository name cannot be empty")
			return
		}

		fmt.Println("Connecting to remote origin")
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

		fmt.Printf("Pushing branch '%s' to GitHub\n", currentBranch)
		cmd := exec.Command("git", "push", "-u", "origin", currentBranch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to push to GitHub: %v\n", err)
			return
		}
	} else {
		promptNewRepo := &survey.Input{
			Message: "Enter new repository name:",
		}
		survey.AskOne(promptNewRepo, &repoName)

		if repoName == "" {
			fmt.Println("Repository name cannot be empty")
			return
		}

		fmt.Println("Setting up local git")
		exec.Command("git", "init").Run()
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "chore: setup gopipe ci/cd").Run()

		fmt.Println("Creating new repository on GitHub and pushing")
		cmd := exec.Command("gh", "repo", "create", repoName, "--public", "--source=.", "--push")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to create repository: %v\n", err)
			return
		}
	}

	var setupVPS bool
	promptSetupVPS := &survey.Confirm{
		Message: "Setup CI/CD to VPS now?",
		Default: true,
	}
	survey.AskOne(promptSetupVPS, &setupVPS)

	if !setupVPS {
		fmt.Println("\nProject is now on GitHub with CI/CD enabled")
		return
	}

	var vpsIP, vpsUser, keyPath string
	appConfig, err := config.LoadConfig()
	useSaved := false

	if err == nil {
		fmt.Printf("Found saved VPS config: %s@%s\n", appConfig.VpsUser, appConfig.VpsIP)
		promptUseSaved := &survey.Confirm{
			Message: "Use this config?",
			Default: true,
		}
		survey.AskOne(promptUseSaved, &useSaved)
	}

	if useSaved {
		vpsIP = appConfig.VpsIP
		vpsUser = appConfig.VpsUser
		keyPath = appConfig.VpsKeyPath
	} else {
		promptVpsIP := &survey.Input{
			Message: "Enter VPS IP:",
		}
		survey.AskOne(promptVpsIP, &vpsIP)

		promptVpsUser := &survey.Input{
			Message: "Enter VPS Username:",
		}
		survey.AskOne(promptVpsUser, &vpsUser)

		promptKeyPath := &survey.Input{
			Message: "Enter path to SSH Private Key (e.g. /home/user/.ssh/id_rsa):",
		}
		survey.AskOne(promptKeyPath, &keyPath)

		if vpsIP == "" || vpsUser == "" || keyPath == "" {
			fmt.Println("VPS details cannot be empty")
			return
		}

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
		fmt.Printf("Failed to read SSH Key at %s: %v\n", resolvedPath, err)
		return
	}

	fmt.Println("Uploading secrets to GitHub")
	github.SetSecret("SSH_HOST", vpsIP)
	github.SetSecret("SSH_USER", vpsUser)
	github.SetSecret("SSH_KEY", string(keyContent))

	fmt.Println("\nSetup complete Every time you push, your project will be automatically deployed to ~/apps/gopipe on your VPS")
}
