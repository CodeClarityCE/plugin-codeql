package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	plugin "github.com/CodeClarityCE/plugin-codeql/src"
	output "github.com/CodeClarityCE/plugin-codeql/src/types"
	amqp_helper "github.com/CodeClarityCE/utility-amqp-helper"
	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	types_amqp "github.com/CodeClarityCE/utility-types/amqp"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin_db "github.com/CodeClarityCE/utility-types/plugin_db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Define the arguments you want to pass to the callback function
type Arguments struct {
	codeclarity *bun.DB
	knowledge   *bun.DB
}

// main is the entry point of the program.
// It reads the configuration, initializes the necessary databases and graph,
// and starts listening on the queue.
func main() {
	config, err := readConfig()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	host := os.Getenv("PG_DB_HOST")
	if host == "" {
		log.Printf("PG_DB_HOST is not set")
		return
	}
	port := os.Getenv("PG_DB_PORT")
	if port == "" {
		log.Printf("PG_DB_PORT is not set")
		return
	}
	user := os.Getenv("PG_DB_USER")
	if user == "" {
		log.Printf("PG_DB_USER is not set")
		return
	}
	password := os.Getenv("PG_DB_PASSWORD")
	if password == "" {
		log.Printf("PG_DB_PASSWORD is not set")
		return
	}

	dsn_knowledge := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbhelper.Config.Database.Knowledge + "?sslmode=disable"
	sqldb_knowledge := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn_knowledge)))
	db_knowledge := bun.NewDB(sqldb_knowledge, pgdialect.New())
	defer db_knowledge.Close()

	dsn := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbhelper.Config.Database.Results + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db_codeclarity := bun.NewDB(sqldb, pgdialect.New())
	defer db_codeclarity.Close()

	args := Arguments{
		codeclarity: db_codeclarity,
		knowledge:   db_knowledge,
	}

	// Start listening on the queue
	amqp_helper.Listen("dispatcher_"+config.Name, callback, args, config)
}

func startAnalysis(args Arguments, dispatcherMessage types_amqp.DispatcherPluginMessage, config plugin_db.Plugin, analysis_document codeclarity.Analysis) (map[string]any, codeclarity.AnalysisStatus, error) {
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
	}
	_, err := args.codeclarity.NewInsert().Model(&result).Exec(context.Background())
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
