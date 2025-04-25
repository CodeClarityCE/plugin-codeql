package codeclarity

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"time"

	outputGenerator "github.com/CodeClarityCE/plugin-codeql/src/outputGenerator"
	output "github.com/CodeClarityCE/plugin-codeql/src/types"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
)

// Entrypoint for the plugin
func Start(project_path string, start time.Time) output.Output {
	// Start the plugin
	log.Println("Starting plugin...")

	// In case language is not supported return an error
	if false {
		exceptionManager.AddError("", exceptionManager.UNSUPPORTED_LANGUAGE_REQUESTED, "", exceptionManager.UNSUPPORTED_LANGUAGE_REQUESTED)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}

	// Execute the command and capture output
	// Note: In a real implementation, you would use exec.Command to run the command
	// and handle errors appropriately. This is a placeholder for demonstration purposes.
	databasePath := project_path + "/../javascript-database"
	if _, err := os.Stat(databasePath); err == nil {
		log.Printf("Database exists at %s, deleting...", databasePath)
		cmd := exec.Command("rm", "-rf", databasePath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error deleting database: %v, output: %s", err, string(out))
			exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
			return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
		}
		log.Printf("Database deleted successfully. Output: %s", string(out))
	} else {
		log.Printf("Database does not exist at %s", databasePath)
	}
	cmd := exec.Command("codeql", "database", "create", "--language=javascript-typescript", "--source-root", project_path, project_path+"/../javascript-database")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error creating database: %v", err)
		exceptionManager.AddError("", exceptionManager.GENERIC_ERROR, "", exceptionManager.GENERIC_ERROR)
		return outputGenerator.FailureOutput(output.AnalysisInfo{}, start)
	}
	log.Println("Database created successfully.")

	cmd = exec.Command("codeql", "database", "analyze", project_path+"/../javascript-database", "--format=sarif-latest", "--output="+project_path+"/../out.sarif")
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

	// Read the SARIF file
	sarifData, err := os.ReadFile(sarifPath)
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
