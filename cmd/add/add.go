package add

import (
	"bytes"
	"cveguard-client/types"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a project",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("Could not get the current working directory:", err)
		}

		postToServer(args[0], filepath.Join(cwd, args[1]))
	},
}

func postToServer(projectUUID string, localPath string) {
	userOS := runtime.GOOS
	hostName, _ := os.Hostname()

	values := map[string]string{"project_uuid": projectUUID, "client_id": viper.GetString("client_id"), "os": userOS, "hostname": hostName}
	jsonValue, _ := json.Marshal(values)
	res, err := http.Post(
		os.Getenv("SERVER_URL")+"/api/project-client",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Fatalln("[ERROR] Could not add project:", err)
	}
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 && res.StatusCode != 201 {
		log.Fatalln("[ERROR] Could not add project:", string(body))
	}
	defer res.Body.Close()

	type Response struct {
		Message     string `json:"message"`
		ProjectName string `json:"project_name"`
	}
	var responseJson Response
	if err := json.Unmarshal(body, &responseJson); err != nil {
		log.Fatalln("[ERROR] Could not parse response from server")
	}

	// Update projects array in configuration
	var projects []types.Project
	mapstructure.Decode(viper.Get("projects"), &projects)
	isEdited := false
	for i, project := range projects {
		if project.UUID == projectUUID {
			projects[i].Name = responseJson.ProjectName
			projects[i].Local_Path = localPath
			isEdited = true
			break
		}
	}
	if !isEdited {
		projects = append(projects, types.Project{Name: responseJson.ProjectName, UUID: projectUUID, Local_Path: localPath})
	}
	viper.Set("projects", projects)
	err = viper.WriteConfig()
	if err != nil {
		log.Fatalln("[ERROR] Could not write config file:", err)
	}

	log.Println("[INFO]", responseJson.Message)
}
