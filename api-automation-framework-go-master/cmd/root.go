package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/copier"

	app "com.tester/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	scenarios    []app.Scenario
	config       app.Config
	allScenarios []app.Scenario
)

//GetConfigs func
func GetConfigs(configsFile *string) app.Config {
	log.Info("Loading test configurations ..:)")
	flag.Parse()
	if *configsFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if !strings.HasSuffix(*configsFile, "yaml") {
		log.Println("provide a configuration yaml file.")
		os.Exit(1)
	}
	abs, err := filepath.Abs(*configsFile)
	if err != nil {
		log.Fatalln(err)
	}

	testData, err := ioutil.ReadFile(abs)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(testData, &config)
	if err != nil {
		log.Fatalln(err)
	}
	return config
}

//GetScenarios func
func GetScenarios(scenarioFolder *string) []app.Scenario {

	log.Info("Loading Test Scenarios ..:)")
	flag.Parse()
	if *scenarioFolder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	scenarioFolders, err := getWalking(*scenarioFolder)
	if err != nil {
		log.Fatal("Error opening folder ", err)
	}
	innerScenario := app.Scenario{}
	for _, file := range scenarioFolders {
		if strings.HasSuffix(file, "yaml") {
			abs, err := filepath.Abs(file)
			if err != nil {
				log.Fatalln(err)
			}
			data, err := ioutil.ReadFile(abs)
			if err != nil {
				log.Fatalln(err)
			}
			err = yaml.Unmarshal(data, &scenarios)
			if err != nil {
				log.Fatalln(err)
			}
			for i, scenario := range scenarios {
				scenario.ID = i + 1
				if scenario.Repricas > 0 {
					for j := 1; j < scenario.Repricas; j++ {
						copier.Copy(&innerScenario, scenario)
						allScenarios = append(allScenarios, innerScenario)
					}
				}
				allScenarios = append(allScenarios, scenario)
			}
		}
	}
	return allScenarios
}

func getWalking(root string) ([]string, error) {
	subDirToSkip := "skip"
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error Walking the path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			fmt.Printf("skipping directory without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
