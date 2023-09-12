package cmds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/authstar/spry/postgres"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema [actor] --output [output]",
	Short: "Output the generated schema from a model into a file",
	Args:  cobra.ExactArgs(1),
	Run:   generateSchema,
}

func GetActorSchema() cobra.Command {
	schemaCmd.Flags().StringP("output", "o", "", "Output path for the generated schema")
	return *schemaCmd
}

func generateSchema(cmd *cobra.Command, args []string) {
	var actorName = args[0]
	if actorName == "" {
		panic("actor name is required to generate a schema")
	}
	var fullPath, _ = cmd.Flags().GetString("output")
	var schema, err = postgres.PostgresGenerateActorSchema(actorName)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate schema: %e", err))
	}
	if fullPath != "" {
		fmt.Printf("Writing schema to %s\n", fullPath)
		pathOnly := filepath.Dir(fullPath)
		err := os.MkdirAll(pathOnly, os.ModePerm)
		if err != nil {
			panic(fmt.Sprintf("Could not create output directory: %e", err))
		} else {
			var file, err = os.Create(fmt.Sprintf("%s/%s.sql", fullPath, actorName))
			if err != nil {
				panic(fmt.Sprintf("Could not create output file: %e", err))
			} else {
				defer file.Close()
				_, err = file.WriteString(schema)
				if err != nil {
					panic(fmt.Sprintf("Could not write to output file: %e", err))
				}
			}
		}
	} else {
		fmt.Printf("%s\n", schema)
	}
}
