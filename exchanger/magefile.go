//+build mage

package main

import (
	"fmt"
	"os"
	"path"
	"sync"
	"unicode/utf8"

	"os/exec"

	"time"

	"bufio"

	"github.com/kr/pretty"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	CORE_PEER_ADDRESS      = getEnvDef("CORE_PEER_ADDRESS", "0.0.0.0:7052")
	CORE_CHAINCODE_ID_NAME = getEnvDef("CORE_CHAINCODE_ID_NAME", ChaincodeName+":0")
	DOCKER_CONTAINER_NAME  = getEnvDef("DOCKER_CONTAINER_NAME", ChaincodeName)
)

const (
	ProjectName   = "go-blockchain-yourself"
	ChaincodeName = "exchanger"

	NetworckSciptsDir = "devnetwork"
	DockerDevModeDir  = "chaincode-docker-devmode"
	CmdDir            = "cmd"
	BuildDir          = "build"
)

var (
	ProjectImportPath               = path.Join("github.com", "ninedraft", ProjectName)
	ChaincodeCmdSourceImportPath    = path.Join(ProjectImportPath, ChaincodeName, CmdDir, ChaincodeName)
	ChaincodeCmdSourcePath          = path.Join(ProjectPath, ChaincodeName, CmdDir, ChaincodeName)
	ProjectPath                     = path.Join(os.Getenv("GOPATH"), "src", ProjectImportPath)
	DevnetworkPAth                  = path.Join(ProjectPath, "devnetwork")
	ChaincodeDockerDevmodePath      = path.Join(DevnetworkPAth, "chaincode-docker-devmode")
	DevnetworkDockerComposeFilePath = path.Join(ChaincodeDockerDevmodePath, "docker-compose-simple.yaml")
	ChaincodeBinaryPath             = path.Join(ProjectPath, ChaincodeName, BuildDir, ChaincodeName)
)

func Build() error {
	var buildCmd = []string{"build"}
	if mg.Verbose() {
		buildCmd = append(buildCmd, "-v")
	}
	buildCmd = append(buildCmd, "-o", ChaincodeBinaryPath, ChaincodeCmdSourceImportPath)
	return sh.Run("go", buildCmd...)
}

func Run() error {
	mg.Deps(Build)
	var envs = Envs{
		"CORE_PEER_ADDRESS":      CORE_PEER_ADDRESS,
		"CORE_CHAINCODE_ID_NAME": CORE_CHAINCODE_ID_NAME,
	}
	if mg.Verbose() {
		pretty.Println(envs)
	}
	return sh.RunWith(envs, ChaincodeBinaryPath)
}

func Help() error {
	fmt.Println("Test hyperledger chaincode.\n" +
		"Run `mage build` to build binary file.\n" +
		"Run `mage run` to run chaincode in dev environment")
	fmt.Println("\n---\nEnvironment:")
	for name, value := range globalEnvs {
		fmt.Printf("\t%s:%q\n", name, value)
	}
	fmt.Println("---")
	return nil
}

func Container() error {
	mg.Deps(Build)
	return sh.Run("docker", "build", "-t", DOCKER_CONTAINER_NAME, ".")
}

func DevNet() error {
	mg.Deps(Container)
	fmt.Printf("creating log file\n")
	var logfile, err = os.Create("devnet.log")
	if err != nil {
		return err
	}
	fmt.Printf("setuping dev network\n")
	var compose = exec.Command("docker-compose", "-f", DevnetworkDockerComposeFilePath, "up")
	compose.Stdout = logfile
	compose.Stderr = logfile
	if err := compose.Start(); err != nil {
		return err
	}
	defer compose.Process.Kill()

	var cli = exec.Command("docker", "exec", "-it", "cli", "bash")
	stdout, err := cli.StdoutPipe()
	if err != nil {
		return err
	}
	stdin, err := cli.StdinPipe()
	if err != nil {
		return err
	}
	fmt.Printf("waiting network to start\n")
	time.Sleep(100 * time.Second)
	fmt.Printf("running cli\n")
	if err := cli.Start(); err != nil {
		return err
	}
	fmt.Printf("executing chaincode invoke\n")
	_, err = fmt.Fprintf(stdin, `peer chaincode invoke -n exchanger -c '{"Args":["get", "a"]}' -C myc\n`)
	if err != nil {
		return err
	}
	var scanner = bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	return scanner.Err()
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
