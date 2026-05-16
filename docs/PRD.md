# PRD - GoPipe

## 1. Overview
GoPipe is a CLI tool designed to automate CI/CD pipeline generation for modern web applications (Node.js, Next.js, and Go). It bridges the gap between local development and automated deployment by generating GitHub Actions workflows and assisting with GitHub/VPS integration.

## 2. Problem Statement
Many developers find setting up CI/CD pipelines (GitHub Actions, SSH secrets, VPS deployment scripts) complex and time-consuming. There is a need for a "zero-config" or "low-config" tool that can detect project settings and set up the entire pipeline automatically.

## 3. Target Audience
- Full-stack developers using Node.js, Next.js, or Go.
- DevOps beginners who want to automate their deployments to a VPS.
- Developers looking for a quick way to bootstrap CI/CD for new projects.

## 4. Key Features
- **Project Detection**: Automatically identifies if a project is Node.js, Next.js, or Go.
- **Port Discovery**: Scans source files and environment variables to detect the application port.
- **Workflow Generation**: Generates optimized `.github/workflows/deploy.yml` based on the detected environment.
- **GitHub Integration**: Connects local projects to GitHub repositories (new or existing) and pushes code.
- **Secret Management**: Automatically uploads SSH secrets (Host, User, Key) to GitHub via the GitHub CLI.
- **VPS Deployment**: Sets up deployment scripts for SCP/Rsync and service management (Systemd for Go, PM2 for Node/Next).

## 5. Technical Requirements
- Written in Go for portability and performance.
- Dependencies: `git`, `gh` (GitHub CLI).
- Supports:
    - Node.js (npm, yarn, pnpm)
    - Next.js (with static export support)
    - Go (main.go entry point)

## 6. Success Metrics
- Reduction in time taken to set up a working CI/CD pipeline from minutes/hours to seconds.
- High accuracy in project and port detection.
- Seamless "one-command" experience for developers.
