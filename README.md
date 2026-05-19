<div align="center">
  <img src="./Logo.png" alt="GoPipe Logo" width="200"/>
</div>

# GoPipe

GoPipe is a CLI tool to generate GitHub Actions CI/CD workflows automatically No more YAML headaches, let GoPipe handle it

---

## Features

- Auto Project Detection: Detect Node.js, Next.js, and Go automatically
- Smart Port Discovery: Find application port from .env or source code
- GitHub Integration: Create new repository or connect to existing one
- Secret Automation: Upload SSH secrets to GitHub automatically using gh CLI
- VPS Deployment: Setup deployment using SCP/Rsync and auto-restart service (PM2/Systemd)

---

## Workflow Diagram

```mermaid
graph TD
    A[Local Code] -->|gopipe start| B{Detect Project}
    B -->|Node/Next| C[Generate Node Workflow]
    B -->|Go| D[Generate Go Workflow]
    C --> E[Upload to GitHub]
    D --> E
    E -->|Push to Main| F[GitHub Actions]
    F -->|Build| G[Build Artifacts]
    G -->|Deploy| H[VPS Server]
    H -->|Restart| I[App Live]
```

---

## Installation

```bash
# Clone repository
git clone https://github.com/RayhanDitaAdam/GoPipe.git

# Enter folder
cd GoPipe

# Install
bash install.sh
```

---

## Usage

Run this command in your project root:

```bash
gopipe start
```

Follow the on-screen instructions

---

## Project Structure

```text
.
├── cmd/
│   └── gopipe/
│       └── main.go       # Entry point
├── internal/
│   ├── commands/         # CLI logic & routing
│   ├── config/           # Config management
│   ├── detector/         # Project & Port detection
│   ├── github/           # GitHub CLI integration
│   └── workflow/         # YAML Generator
├── PRD.md                # Product Requirements
├── FUNDAMENTAL.md        # Technical fundamentals
└── README.md             # This file
```

---

## Documentation

- [PRD.md](./docs/PRD.md) - Product vision and features
- [FUNDAMENTAL.md](./docs/FUNDAMENTAL.md) - Technical architecture and logic

---

## Contributing

Want to add features or fix bugs? Feel free to open a Pull Request

---

Made by [RayhanDitaAdam](https://github.com/RayhanDitaAdam)
