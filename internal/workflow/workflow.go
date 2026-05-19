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
		return fmt.Errorf("failed to create directory: %w", err)
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

			deploySteps = fmt.Sprintf(`      - name: Deploy to Server
        uses: appleboy/scp-action@v0.1.7
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          source: "app"
          target: "~/apps/%[1]s"

      - name: Restart Service
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd ~/apps/%[1]s
            chmod +x app
            sudo systemctl restart %[1]s || true
            nohup ./app > app.log 2>&1 &`, proj.AppName())
		}

	case "Next.js":
		{
			installCmd := "npm install"
			buildCmd := "npm run build"
			pnpmStep := ""
			pmName := "npm"

			switch proj.PackageManager {
			case "pnpm":
				installCmd = "pnpm install"
				buildCmd = "pnpm run build"
				pmName = "pnpm"
				pnpmStep = `      - name: Install pnpm
        uses: pnpm/action-setup@v4
        with:
          version: 9
`
			case "yarn":
				installCmd = "yarn install"
				buildCmd = "yarn run build"
				pmName = "yarn"
			case "npm-ci":
				installCmd = "npm ci"
				buildCmd = "npm run build"
				pmName = "npm"
			case "npm":
				installCmd = "npm install"
				buildCmd = "npm run build"
				pmName = "npm"
			}

			buildSteps = fmt.Sprintf(`      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

%s
      - name: Install dependencies
        run: %s`, pnpmStep, installCmd)

			if proj.HasBuildScript {
				buildSteps += fmt.Sprintf(`

      - name: Build
        run: %s`, buildCmd)
			}

			buildSteps += `

      - name: Export static
        if: hashFiles('next.config.js') != ''
        run: PM_REPLACE run export || true`

			buildSteps = strings.ReplaceAll(buildSteps, "PM_REPLACE", pmName)

			deploySteps = fmt.Sprintf(`      - name: Create Target Directory
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: mkdir -p ~/apps/%[1]s

      - name: Deploy via Rsync
        uses: burnett01/rsync-deployments@v8
        with:
          switches: -avzr --delete
          path: .next/
          remote_path: ~/apps/%[1]s
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
            export NVM_DIR="$HOME/.nvm"
            [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
            export PATH=$PATH:/usr/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.npm-global/bin
            cd ~/apps/%[1]s
            PM_REPLACE install
            pm2 restart %[1]s || pm2 start "PM_REPLACE run start -- --port $PORT" --name %[1]s || pm2 start "PM_REPLACE run preview -- --port $PORT" --name %[1]s`, proj.AppName())

			deploySteps = strings.ReplaceAll(deploySteps, "PM_REPLACE", pmName)
		}

	case "Node.js":
		{
			installCmd := "npm install"
			buildCmd := "npm run build"
			pnpmStep := ""
			pmName := "npm"

			switch proj.PackageManager {
			case "pnpm":
				installCmd = "pnpm install"
				buildCmd = "pnpm run build"
				pmName = "pnpm"
				pnpmStep = `      - name: Install pnpm
        uses: pnpm/action-setup@v4
        with:
          version: 9
`
			case "yarn":
				installCmd = "yarn install"
				buildCmd = "yarn run build"
				pmName = "yarn"
			case "npm-ci":
				installCmd = "npm ci"
				buildCmd = "npm run build"
				pmName = "npm"
			case "npm":
				installCmd = "npm install"
				buildCmd = "npm run build"
				pmName = "npm"
			}

			buildSteps = fmt.Sprintf(`      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

%s
      - name: Install dependencies
        run: %s`, pnpmStep, installCmd)

			if proj.HasBuildScript {
				buildSteps += fmt.Sprintf(`

      - name: Build
        run: %s`, buildCmd)
			}

			deploySteps = fmt.Sprintf(`      - name: Create Target Directory
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: mkdir -p ~/apps/%[1]s

      - name: Deploy via Rsync
        uses: burnett01/rsync-deployments@v8
        with:
          switches: -avzr --delete
          path: ./ --exclude node_modules --exclude .git
          remote_path: ~/apps/%[1]s
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
            export NVM_DIR="$HOME/.nvm"
            [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
            export PATH=$PATH:/usr/bin:/usr/local/bin:$HOME/.local/bin:$HOME/.npm-global/bin
            cd ~/apps/%[1]s
            PM_REPLACE install
            pm2 restart %[1]s || pm2 start "PM_REPLACE run start -- --port $PORT" --name %[1]s || pm2 start "PM_REPLACE run preview -- --port $PORT" --name %[1]s`, proj.AppName())

			deploySteps = strings.ReplaceAll(deploySteps, "PM_REPLACE", pmName)
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