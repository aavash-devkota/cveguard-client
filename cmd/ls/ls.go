package ls

import (
	"cveguard-client/types"
	"fmt"

	"github.com/clinaresl/table"
	_ "github.com/joho/godotenv/autoload"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List of all added projects",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Get project details such as `status` and `last_scanned_at` from server

		var projects []types.Project
		mapstructure.Decode(viper.Get("projects"), &projects)

		t, _ := table.NewTable("|c|c|c|")
		t.AddSingleRule()
		t.AddRow("Name", "UUID", "Project Path")
		t.AddSingleRule()

		for _, project := range projects {
			t.AddRow(" "+project.Name+" ", " "+project.UUID+" ", " "+project.Local_Path+" ") // Padding is added on each entries for easier text copy
			t.AddSingleRule()
		}
		fmt.Printf("%v\n", t)
	},
}
