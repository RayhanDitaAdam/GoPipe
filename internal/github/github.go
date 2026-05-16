package github

import (
	"fmt"
	"os/exec"
)

func SetSecret(name, value string) {
	cmd := exec.Command("gh", "secret", "set", name, "--body", value)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Gagal set secret %s: %v\n", name, err)
	} else {
		fmt.Printf("✅ Secret %s berhasil diset.\n", name)
	}
}
