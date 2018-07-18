//+build mage

package main

import (
	"fmt"
	"os"
	"path"
	"sync"
	"unicode/utf8"

	"github.com/kr/pretty"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	CORE_PEER_ADDRESS      = getEnvDef("CORE_PEER_ADDRESS", "0.0.0.0:7052")
	CORE_CHAINCODE_ID_NAME = getEnvDef("CORE_CHAINCODE_ID_NAME", ProjectName+":0")
	DOCKER_CONTAINER_NAME  = getEnvDef("DOCKER_CONTAINER_NAME", ProjectName)
)

const (
	ProjectName = "exchanger"

	CmdPath   = "cmd"
	BuildPath = "build"
)

var (
	PWD = path.Join(os.Getenv("GOPATH"), "github.com", "ninedraft", ProjectName)
)

func Build() error {
	var buildTarget = path.Join(PWD, BuildPath, ProjectName)
	var buildSource = "./" + path.Join(CmdPath, ProjectName)
	var buildCmd = []string{"build"}
	if mg.Verbose() {
		buildCmd = append(buildCmd, "-v")
	}
	buildCmd = append(buildCmd, "-o", buildTarget, buildSource)
	return sh.Run("go", buildCmd...)
}

func Run() error {
	mg.Deps(Build, DeveloperNetwork)
	var envs = Envs{
		"CORE_PEER_ADDRESS":      CORE_PEER_ADDRESS,
		"CORE_CHAINCODE_ID_NAME": CORE_CHAINCODE_ID_NAME,
	}
	if mg.Verbose() {
		pretty.Println(envs)
	}
	var runPath = path.Join(PWD, BuildPath, ProjectName)
	return sh.RunWith(envs, runPath)
}

func Help() error {
	fmt.Println("Test hyperledger chaincode.\n" +
		"Run `mage build` to build binary file.\n" +
		"Run `mage run` to run chaincode in dev environment\n")
	fmt.Println("\n---\nEnvironment:")
	for name, value := range globalEnvs {
		fmt.Printf("\t%s:%q\n", name, value)
	}
	fmt.Println("---")
	return nil
}

func DeveloperNetwork() error {
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type Envs map[string]string

var globalEnvs = Envs{}
var envMutex sync.Mutex

func getEnvDef(name, def string) string {
	var value, ok = os.LookupEnv(name)
	if !ok {
		value = def
	}

	envMutex.Lock()
	defer envMutex.Unlock()
	globalEnvs[name] = value

	return value
}

func maxStringWidth(strs []string) int {
	var width = 0
	for _, str := range strs {
		var strWidth = utf8.RuneCountInString(str)
		if strWidth > width {
			width = strWidth
		}
	}
	return width
}
