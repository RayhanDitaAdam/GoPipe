# Fundamental Concepts - GoPipe

## 1. Core Architecture
GoPipe is built as a modular CLI application in Go. Its logic is divided into specialized modules:

- **Detector**: Responsible for inspecting the filesystem to identify the programming language, framework, package manager, and application port.
- **Generator**: Uses the detected project information to construct a YAML workflow file for GitHub Actions.
- **Integrator**: Interfaces with `git` and `gh` (GitHub CLI) to automate repository setup and secret management.
- **Manager**: Handles local configuration (saved in `~/.gopipe.json`) to remember VPS details.

## 2. Detection Logic
GoPipe uses a heuristic-based detection system:
- **Language**: Checks for `package.json` (Node/Next) or `go.mod` (Go).
- **Framework**: Looks for `next.config.js` or specific directory structures (`pages`, `app`) to distinguish Next.js from generic Node.js.
- **Port**:
    1. Scans `.env`, `.env.local`, `.env.example` for `PORT=`.
    2. Scans `package.json` scripts for `-p` flags (common in Next/Vite).
    3. Scans `.go` files for `ListenAndServe` or `Run` calls with port patterns.
    4. Defaults to 3000 (Node/Next) or 8080 (Go).

## 3. CI/CD Workflow Strategy
The generated `deploy.yml` follows industry best practices:
- **Triggers**: Runs on every `push` to `main` or `master` branches.
- **Build Phase**: Compiles the application or installs dependencies in a clean Ubuntu environment.
- **Deploy Phase**:
    - **Go**: Uses `appleboy/scp-action` for binary transfer and `appleboy/ssh-action` for service restart.
    - **Node/Next**: Uses `burnett01/rsync-deployments` for efficient file synchronization and `appleboy/ssh-action` for PM2 management.

## 4. Security
- GoPipe **never** hardcodes sensitive information in the workflow file.
- It leverages **GitHub Secrets** for SSH details (`SSH_HOST`, `SSH_USER`, `SSH_KEY`).
- Local VPS configs are stored with restricted permissions (`0600`) in the user's home directory.

## 5. Extensibility
The codebase is designed to be easily extendable. To add support for a new language (e.g., Python):
1. Add a new case in `detectProject` (detector.go).
2. Define build/deploy steps in `buildWorkflow` (workflow.go).
