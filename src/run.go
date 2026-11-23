package codeclarity

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	outputGenerator "github.com/CodeClarityCE/plugin-codeql/src/outputGenerator"
	output "github.com/CodeClarityCE/plugin-codeql/src/types"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
)

// Entrypoint for the plugin
func Start(project_path string, language string, start time.Time) output.Output {
	// Start the plugin
	log.Println("Starting plugin...")

	// In case language is not supported return an error
	if false {
		exceptionManager.AddError("", exceptionManager.UNSUPPORTED_LANGUAGE_REQUESTED, "", exceptionManager.UNSUPPORTED_LANGUAGE_REQUESTED)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}

	switch language {
	case "js", "ts", "javascript", "typescript", "javascript-typescript":
		language = "javascript-typescript"
	case "go", "golang":
		language = "go"
	case "python", "py":
		language = "python"
	}

	// Execute the command and capture output
	// Note: In a real implementation, you would use exec.Command to run the command
	// and handle errors appropriately. This is a placeholder for demonstration purposes.
	databasePath := project_path + "/../project-database"
	if _, err := os.Stat(databasePath); err == nil {
		log.Printf("Database exists at %s, deleting...", databasePath)
		if err := os.RemoveAll(databasePath); err != nil {
			log.Printf("Error deleting database: %v", err)
			exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
			return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
		}
		log.Printf("Database deleted successfully.")
	} else {
		log.Printf("Database does not exist at %s", databasePath)
	}
	cmd := exec.Command("codeql", "database", "create", "--language="+language, "--source-root", project_path, project_path+"/../project-database") //nolint:gosec // G204: fixed binary on the project paths
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error creating database: %v", err)
		exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}
	log.Println("Database created successfully.")

	cmd = exec.Command("codeql", "database", "analyze", project_path+"/../project-database", "--format=sarif-latest", "--output="+project_path+"/../out.sarif") //nolint:gosec // G204: fixed binary on the project paths
	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error analyzing database: %v", err)
		exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}
	log.Println("Database analyzed successfully.")

	// Parse the SARIF output
	sarifPath := project_path + "/../out.sarif"
	var codeQL output.CodeQL

	// Read the SARIF file from the directory codeql wrote it to.
	sarifData, err := readInDir(filepath.Dir(sarifPath), filepath.Base(sarifPath))
	if err != nil {
		log.Printf("Error reading SARIF file: %v", err)
		exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}
	if err := json.Unmarshal(sarifData, &codeQL); err != nil {
		log.Printf("Error unmarshalling SARIF file: %v", err)
		exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}

	// TODO perform analysis and fill this object
	// You can adapt its type your needs
	data := map[string]output.WorkspaceInfo{
		".": {
			Results: codeQL.Runs[0].Results,
		},
	}

	// Generate license stats
	analysisStats := outputGenerator.GenerateAnalysisStats(data)

	// Return the analysis results
	return outputGenerator.SuccessOutput(data, analysisStats, output.AnalysisInfo{}, start)
}

// readInDir reads name without leaving dir.
func readInDir(dir, name string) ([]byte, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.ReadFile(name)
}
