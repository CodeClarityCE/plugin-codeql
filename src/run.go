package codeclarity

import (
	"log"
	"os/exec"
)

// Entrypoint for the plugin
func Start() any {
	// Start the plugin
	log.Println("Starting plugin...")

	// Run the command
	cmd := exec.Command("codeql", "database", "create", "test", "--overwrite", "--language=javascript", "--source-root=/private/10242/10798/main")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed to create CodeQL database:", err)
		return err.Error()
	}

	log.Println("CodeQL database created successfully:", string(output))

	return "Hello, World!"
}
