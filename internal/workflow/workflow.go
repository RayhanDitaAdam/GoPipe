package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopipe/internal/detector"
)

func GenerateGithubActions(proj detector.Project) error {
	workDir := filepath.Join(proj.RootDir, ".github", "workflows")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("gagal buat direktori: %w", err)
	}

	action := buildWorkflow(proj)
	path := filepath.Join(workDir, "deploy.yml")
	return os.WriteFile(path, []byte(action), 0644)
}

func buildWorkflow(proj detector.Project) string {
	var buildSteps string
	var deploySteps string

	switch proj.Language {
	case "Go":
		{
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
		}

	case "Next.js":
		{
			installCmd := "npm ci"
			buildCmd := "npm run build"
			pnpmStep := ""
			if proj.PackageManager == "pnpm" {
				installCmd = "pnpm install"
				buildCmd = "pnpm run build"
				pnpmStep = `      - name: Install pnpm
        uses: pnpm/action-setup@v4
        with:
          version: 9
`
			} else if proj.PackageManager == "yarn" {
				installCmd = "yarn install"
				buildCmd = "yarn run build"
			}

			buildSteps = fmt.Sprintf(`      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

%s
      - name: Install dependencies
        run: %s

      - name: Build
        run: %s

      - name: Export static
        if: hashFiles('next.config.js') != ''
        run: PM_REPLACE run export || true`, pnpmStep, installCmd, buildCmd)

			buildSteps = strings.ReplaceAll(buildSteps, "PM_REPLACE", proj.PackageManager)

			deploySteps = `      - name: Create Target Directory
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: mkdir -p ~/apps/gopipe

      - name: Deploy via Rsync
        uses: burnett01/rsync-deployments@v8
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
            export PATH=$PATH:/usr/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.npm-global/bin
            cd ~/apps/gopipe
            npm install
            pm2 restart gopipe || pm2 start "npm run start -- --port $PORT" --name gopipe || pm2 start "npm run preview -- --port $PORT" --name gopipe`

			deploySteps = strings.ReplaceAll(deploySteps, "PM_REPLACE", proj.PackageManager)
		}

	case "Node.js":
		{
			installCmd := "npm ci"
			buildCmd := "npm run build"
			pnpmStep := ""
			if proj.PackageManager == "pnpm" {
				installCmd = "pnpm install"
				buildCmd = "pnpm run build"
				pnpmStep = `      - name: Install pnpm
        uses: pnpm/action-setup@v4
        with:
          version: 9
`
			} else if proj.PackageManager == "yarn" {
				installCmd = "yarn install"
				buildCmd = "yarn run build"
			}

			buildSteps = fmt.Sprintf(`      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

%s
      - name: Install dependencies
        run: %s

      - name: Build
        run: %s`, pnpmStep, installCmd, buildCmd)

			deploySteps = `      - name: Create Target Directory
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: mkdir -p ~/apps/gopipe

      - name: Deploy via Rsync
        uses: burnett01/rsync-deployments@v8
        with:
          switches: -avzr --delete
          path: ./ --exclude node_modules --exclude .git
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
            export PATH=$PATH:/usr/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.npm-global/bin
            cd ~/apps/gopipe
            npm install
            pm2 restart gopipe || pm2 start "npm run start -- --port $PORT" --name gopipe || pm2 start "npm run preview -- --port $PORT" --name gopipe`

			deploySteps = strings.ReplaceAll(deploySteps, "PM_REPLACE", proj.PackageManager)
		}
	}

	return fmt.Sprintf(`name: Deploy

on:
  push:
    branches: [ main, master ]

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
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        