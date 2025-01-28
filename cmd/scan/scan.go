package scan

import (
	"bytes"
	"cveguard-client/types"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a project for vulnerabilities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectUUID := args[0]

		var projects []types.Project
		mapstructure.Decode(viper.Get("projects"), &projects)

		for _, project := range projects {
			if project.UUID == projectUUID {
				log.Println("[INFO] Parsing the packages information...")
				packagesList := getPackagesList(project.Local_Path, "npm")
				postToServer(args[0], packagesList)

				os.Exit(0)
			}
		}

		log.Fatalln("[ERROR] Invalid project UUID")
	},
}

func getPackagesList(filePath string, ecosystem string) map[string][]string {
	var packagesList map[string][]string

	switch ecosystem {
	case "npm":
		packagesList = getPackagesListNPM(filePath)
	}

	return packagesList
}

func postToServer(projectUUID string, dependencies map[string][]string) {
	log.Println("[INFO] Scanning for vulnerabilities...")

	type DependencyDetails struct {
		Name         string   `json:"name"`
		Dependencies []string `json:"dependencies"`
	}

	// Convert dependencies map to string array
	dependenciesArr := make([]DependencyDetails, 0, len(dependencies))
	for name, deps := range dependencies {
		dependenciesArr = append(dependenciesArr, DependencyDetails{
			Name:         name,
			Dependencies: deps,
		})
	}

	type inputValues struct {
		ProjectUUID  string              `json:"project_uuid"`
		ClientID     string              `json:"client_id"`
		Dependencies []DependencyDetails `json:"dependencies"`
	}

	values := inputValues{ProjectUUID: projectUUID, ClientID: viper.GetString("client_id"), Dependencies: dependenciesArr}
	jsonValue, _ := json.Marshal(values)
	res, err := http.Post(
		os.Getenv("SERVER_URL")+"/api/project-scans",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Fatalln("[ERROR] Could not scan project:", err)
	}
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 201 {
		log.Fatalln("[ERROR] Could not scan project: Status code", res.StatusCode)
	}
	defer res.Body.Close()

	type Response struct {
		ScanURL string `json:"scan_url"`
	}
	var responseJson Response
	if err := json.Unmarshal(body, &responseJson); err != nil {
		log.Fatalln("[ERROR] Could not parse response from server")
	}

	log.Println("[INFO] Scan complete!")
	log.Println("[INFO] See scan details at:", responseJson.ScanURL)
}
