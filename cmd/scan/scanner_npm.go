package scan

import (
	"encoding/json"
	"github.com/Masterminds/semver/v3"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ledongthuc/goterators"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

type PackageLockFile struct {
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	LockfileVersion  int                    `json:"lockfileVersion"`
	Requires         bool                   `json:"requires"`
	Packages         map[string]*Package    `json:"packages"`
	RootDependencies map[string]*Dependency `json:"dependencies"`
}

type Package struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Requires        map[string]string `json:"requires"`
	Optional        bool              `json:"optional"`
}

type Dependency struct {
	Version         string                       `json:"version"`
	Dependencies    map[string]*NestedDependency `json:"dependencies"`
	DevDependencies map[string]string            `json:"devDependencies"`
	Requires        map[string]string            `json:"requires"`
}

type NestedDependency struct {
	Version      string                       `json:"version"`
	Dependencies map[string]*NestedDependency `json:"dependencies"`
	Requires     map[string]string            `json:"requires"`
}

func getPackagesListNPM(dirToScan string) map[string][]string {
	// Verify that the path is valid
	if _, err := os.Stat(dirToScan); os.IsNotExist(err) {
		log.Fatalln("Please provide valid path to project")
	}

	// Check if the project is a valid javascript project
	packageLockPath := filepath.Join(dirToScan, "package-lock.json")
	if _, err := os.Stat(packageLockPath); os.IsNotExist(err) {
		log.Fatalln("Given project directory is not a valid javascript project or is not using npm")
	}

	// Read the contents of the package-lock.json file and get the dependency tree
	packageLockFile, err := os.Open(packageLockPath)
	if err != nil {
		log.Fatalln("Could not read the package-lock.json file in the project")
	}
	jsonParser := json.NewDecoder(packageLockFile)
	var fileContentObject PackageLockFile
	if err = jsonParser.Decode(&fileContentObject); err != nil {
		log.Fatalln("Error parsing package-lock.json file:", err.Error())
	}

	// Get plain list of packages

	rootPackageName := ""

	packagesList := make(map[string]string)
	for packageName, packageDetails := range fileContentObject.Packages {
		pkg := packageName
		if packageName == "" {
			rootPackageName = packageDetails.Name + "@@" + packageDetails.Version
			pkg = packageDetails.Name
		}

		if packageDetails.Optional {
			continue
		}

		pkgSplits := strings.Split(pkg, "node_modules/")
		pkg = pkgSplits[len(pkgSplits)-1]
		packagesList[pkg+"@@"+packageDetails.Version] = packageName
	}

	// Then get the dependencies tree starting from root

	type dependencyItem struct {
		id       string
		parentId *string
		label    string
	}

	var dependenciesItems []dependencyItem

	addedItems := mapset.NewSet[string]()

	var parseDependenciesItems func(*string, map[string]string) // Defined here first to recursively call it
	parseDependenciesItems = func(parentId *string, dependencies map[string]string) {
		for dependency, dependencySemVer := range dependencies {
			dependencySemVerSplits := strings.Split(dependencySemVer, "@") // Some SemVers can contain the package name Example: 'npm:string-width@^4.2.0'
			if len(dependencySemVerSplits) > 0 {
				dependencySemVer = dependencySemVerSplits[len(dependencySemVerSplits)-1]
			}

			for packageEntry := range packagesList {
				packageSplits := strings.Split(packageEntry, "@@")
				packageName := packageSplits[0]
				packageVersion := packageSplits[1]

				semverConstraint, _ := semver.NewConstraint(dependencySemVer)
				packageVersionInstance, _ := semver.NewVersion(packageVersion)
				isPackageVersionSatisfied, _ := semverConstraint.Validate(packageVersionInstance)
				if err != nil {
					log.Fatalln("Package version not satisfied:", err)
				}

				if packageName == dependency && isPackageVersionSatisfied {
					isAddedAlready := addedItems.Contains(packageEntry)
					if !isAddedAlready {
						addedItems.Append(packageEntry)
					}

					dependenciesItems = append(dependenciesItems, dependencyItem{
						id:       packageEntry,
						parentId: parentId,
						label:    packageEntry,
					})

					if !isAddedAlready {
						parseDependenciesItems(&packageEntry, fileContentObject.Packages[packagesList[packageEntry]].Dependencies)
					}
					break
				}
			}
		}
	}

	depsAndDevDep := map[string]string{}
	maps.Copy(depsAndDevDep, fileContentObject.Packages[""].Dependencies)
	maps.Copy(depsAndDevDep, fileContentObject.Packages[""].DevDependencies)
	parseDependenciesItems(nil, depsAndDevDep)

	// Get dependency flow of each individual packages

	dependencyFlows := map[string][]string{}
	for packageName := range packagesList {
		var dependencies []string
		dependenciesBacktracks := goterators.Filter(dependenciesItems, func(d dependencyItem) bool {
			return d.id == packageName
		})

		for _, dependenciesBacktrack := range dependenciesBacktracks {
			var backtrackEntry string

			if dependenciesBacktrack.parentId != nil {
				backtrackEntry = *dependenciesBacktrack.parentId
			}
			if backtrackEntry == "" {
				backtrackEntry = rootPackageName
			}

			dependencies = append(dependencies, backtrackEntry)
		}

		dependencyFlows[packageName] = dependencies
	}

	return dependencyFlows
}
