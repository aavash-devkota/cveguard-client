package remove

import (
	"cveguard-client/types"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a project (Only from client connection)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		postToServer(args[0])
	},
}

func postToServer(projectUUID string) {
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodDelete, os.Getenv("SERVER_URL")+"/api/project-client/"+viper.GetString("client_id")+"?project_uuid="+projectUUID, nil)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln("[ERROR] Could not remove project:", err)
	}
	if res.StatusCode == 404 {
		log.Fatalln("[ERROR] Project not found for this client")
	}

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		log.Fatalln("[ERROR] Could not remove project:", string(body))
	}
	defer res.Body.Close()

	type Response struct {
		Message string `json:"message"`
	}
	var responseJson Response
	if err := json.Unmarshal(body, &responseJson); err != nil {
		log.Fatalln("[ERROR] Could not parse response from server")
	}

	// Update projects array in configuration
	var projects []types.Project
	mapstructure.Decode(viper.Get("projects"), &projects)
	for i, project := range projects {
		if project.UUID == projectUUID {
			projects = append(projects[:i], projects[i+1:]...)
			break
		}
	}
	viper.Set("projects", projects)
	err = viper.WriteConfig()
	if err != nil {
		log.Fatalln("[ERROR] Could not write config file:", err)
	}

	log.Println("[INFO]", responseJson.Message)
}
