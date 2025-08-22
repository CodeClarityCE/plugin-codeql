package main

import (
	"context"
	"log"
	"os"
	"time"

	plugin "github.com/CodeClarityCE/plugin-codeql/src"
	output "github.com/CodeClarityCE/plugin-codeql/src/types"
	"github.com/CodeClarityCE/utility-types/boilerplates"
	types_amqp "github.com/CodeClarityCE/utility-types/amqp"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin_db "github.com/CodeClarityCE/utility-types/plugin_db"
)

// CodeQLAnalysisHandler implements the AnalysisHandler interface
type CodeQLAnalysisHandler struct{}

// StartAnalysis implements the AnalysisHandler interface
func (h *CodeQLAnalysisHandler) StartAnalysis(
	databases *boilerplates.PluginDatabases,
	dispatcherMessage types_amqp.DispatcherPluginMessage,
	config plugin_db.Plugin,
	analysisDoc codeclarity.Analysis,
) (map[string]any, codeclarity.AnalysisStatus, error) {
	return startAnalysis(databases, dispatcherMessage, config, analysisDoc)
}

// main is the entry point of the program.
func main() {
	pluginBase, err := boilerplates.CreatePluginBase()
	if err != nil {
		log.Fatalf("Failed to initialize plugin base: %v", err)
	}
	defer pluginBase.Close()

	// Start the plugin with our analysis handler
	handler := &CodeQLAnalysisHandler{}
	err = pluginBase.Listen(handler)
	if err != nil {
		log.Fatalf("Failed to start plugin: %v", err)
	}
}

func startAnalysis(databases *boilerplates.PluginDatabases, dispatcherMessage types_amqp.DispatcherPluginMessage, config plugin_db.Plugin, analysis_document codeclarity.Analysis) (map[string]any, codeclarity.AnalysisStatus, error) {
	// Get analysis config
	messageData := analysis_document.Config[config.Name].(map[string]any)

	// GET download path from ENV
	path := os.Getenv("DOWNLOAD_PATH")

	// Destination folder
	// destination := fmt.Sprintf("%s/%s/%s", path, organization, analysis.Commit)
	// Prepare the arguments for the plugin
	project := path + "/" + messageData["project"].(string)
	language := messageData["language"].(string)

	// Start the plugin
	out := plugin.Start(project, language, time.Now())

	result := codeclarity.Result{
		Result:     output.ConvertOutputToMap(out),
		AnalysisId: dispatcherMessage.AnalysisId,
		Plugin:     config.Name,
		CreatedOn:  time.Now(),
	}
	_, err := databases.Codeclarity.NewInsert().Model(&result).Exec(context.Background())
	if err != nil {
		panic(err)
	}

	// Prepare the result to store in step
	// In this case we only store the sbomKey
	// The other plugins will use this key to get the sbom
	res := make(map[string]any)
	res["codeQLKey"] = result.Id

	// The output is always a map[string]any
	return res, out.AnalysisInfo.Status, nil
}
