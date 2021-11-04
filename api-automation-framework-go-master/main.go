
package main

import (
	"com.tester/cmd"
	app "com.tester/utils"
	"flag"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"os"
	"time"
)

var (
	configsPath  *string
	scenarioPath *string
	scenarios    []app.Scenario
	config app.Config
	sessionID string
	runAt time.Time
)

func init() {
	configsPath = flag.String("c", "", "config file")
	scenarioPath = flag.String("s", "", "scenarios directory/file")
	flag.Parse()
	if *configsPath == "" || *scenarioPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	scenarios = cmd.GetScenarios(scenarioPath)
	config = cmd.GetConfigs(configsPath)
	u := uuid.NewV4()
	runAt = time.Now()
	sessionID = runAt.Format("2006-01-02T15:04:05")+"_"+ u.String()
}

func main() {
	totalScenarios := len(scenarios)
	finalScenarios := make(chan app.Scenario, totalScenarios)
	app.Commander(totalScenarios,finalScenarios,sessionID,scenarios, config)
	fmt.Printf("Run %d in %s ", totalScenarios,time.Since(runAt) )
}
