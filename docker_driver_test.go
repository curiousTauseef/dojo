package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)


func setTestEnv() {
	os.Unsetenv("DISPLAY")
}

func TestDockerDriver_ConstructDockerRunCmd_Interactive(t *testing.T){
	type mytestStruct struct {
		shellInteractive bool
		userInteractiveConfig string
		expOutput string
	}
	interactiveOutput := "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro " +
		"-v /tmp/some-env-file-multiline:/etc/dojo.d/variables/00-multiline-vars.sh " +
		"-v /tmp/some-env-file-bash-functions:/etc/dojo.d/variables/01-bash-functions.sh " +
		"--env-file=/tmp/some-env-file -ti --name=name1 img:1.2.3"
	notInteractiveOutput := "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro " +
		"-v /tmp/some-env-file-multiline:/etc/dojo.d/variables/00-multiline-vars.sh " +
		"-v /tmp/some-env-file-bash-functions:/etc/dojo.d/variables/01-bash-functions.sh " +
		"--env-file=/tmp/some-env-file --name=name1 img:1.2.3"
	mytests := []mytestStruct{
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "true",
			expOutput: interactiveOutput},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "false",
			expOutput: notInteractiveOutput},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "",
			expOutput: interactiveOutput},

		mytestStruct{ shellInteractive: false, userInteractiveConfig: "true",
			expOutput: interactiveOutput},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "false",
			expOutput: notInteractiveOutput},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "",
			expOutput: notInteractiveOutput},
	}
	setTestEnv()
	logger := NewLogger("debug")
	for _,v := range mytests {
		config := getTestConfig()
		config.Interactive = v.userInteractiveConfig
		var ss ShellServiceInterface
		if v.shellInteractive {
			ss = NewMockedShellServiceInteractive(logger)
		} else {
			ss = NewMockedShellServiceNotInteractive(logger)
		}
		d := NewDockerDriver(ss, NewMockedFileService(logger), logger)
		cmd := d.ConstructDockerRunCmd(config, "/tmp/some-env-file",
			"/tmp/some-env-file-multiline", "/tmp/some-env-file-bash-functions",
			"name1")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("shellInteractive: %v, userConfig: %v", v.shellInteractive, v.userInteractiveConfig))
	}
}

func TestDockerDriver_ConstructDockerRunCmd_Command(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	baseOutput := "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro " +
		"-v /tmp/some-env-file-multiline:/etc/dojo.d/variables/00-multiline-vars.sh " +
		"-v /tmp/some-env-file-bash-functions:/etc/dojo.d/variables/01-bash-functions.sh " +
		"--env-file=/tmp/some-env-file --name=name1 img:1.2.3"
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "",
			expOutput: baseOutput},
		mytestStruct{ userCommandConfig: "bash",
			expOutput: baseOutput + " bash"},
		mytestStruct{ userCommandConfig: "bash -c \"echo hello\"",
			expOutput: baseOutput + " bash -c \"echo hello\""},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		logger := NewLogger("debug")
		d := NewDockerDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
		cmd := d.ConstructDockerRunCmd(config, "/tmp/some-env-file",
			"/tmp/some-env-file-multiline", "/tmp/some-env-file-bash-functions",
			"name1")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func TestDockerDriver_ConstructDockerRunCmd_DisplayEnvVar(t *testing.T){
	type mytestStruct struct {
		displaySet bool
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ displaySet: true,
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro " +
			"-v /tmp/some-env-file-multiline:/etc/dojo.d/variables/00-multiline-vars.sh " +
			"-v /tmp/some-env-file-bash-functions:/etc/dojo.d/variables/01-bash-functions.sh " +
			"--env-file=/tmp/some-env-file -v /tmp/.X11-unix:/tmp/.X11-unix --name=name1 img:1.2.3"},
		mytestStruct{ displaySet: false,
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro " +
			"-v /tmp/some-env-file-multiline:/etc/dojo.d/variables/00-multiline-vars.sh " +
			"-v /tmp/some-env-file-bash-functions:/etc/dojo.d/variables/01-bash-functions.sh " +
			"--env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		if v.displaySet {
			os.Setenv("DISPLAY","123")
		} else {
			setTestEnv()
		}
		logger := NewLogger("debug")
		d := NewDockerDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
		cmd := d.ConstructDockerRunCmd(config, "/tmp/some-env-file",
			"/tmp/some-env-file-multiline", "/tmp/some-env-file-bash-functions",
			"name1")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("displaySet: %v", v.displaySet))
	}
}

func TestDockerDriver_HandleRun_Unit(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	d := NewDockerDriver(NewMockedShellServiceNotInteractive(logger), fs, logger)
	config := getTestConfig()
	config.RunCommand = ""
	envService := NewMockedEnvService()
	envService.AddVariable(`MULTI_LINE=one
two
three`)
	es := d.HandleRun(config, "testrunid", envService)
	assert.Equal(t, 0, es)
	assert.False(t, fileExists("/tmp/dojo-environment-testrunid"))
	assert.False(t, fileExists("/tmp/dojo-environment-multiline-testrunid"))
	assert.False(t, fileExists("/tmp/dojo-environment-bash-functions-testrunid"))
	assert.Equal(t, 3, len(fs.FilesWrittenTo))
	assert.Equal(t, "ABC=123\n", fs.FilesWrittenTo["/tmp/dojo-environment-testrunid"])
	assert.Equal(t, "export MULTI_LINE=$(echo b25lCnR3bwp0aHJlZQ== | base64 -d)\n", fs.FilesWrittenTo["/tmp/dojo-environment-multiline-testrunid"])
	assert.Equal(t, 3, len(fs.FilesRemovals))
	assert.Equal(t, "/tmp/dojo-environment-testrunid", fs.FilesRemovals[1])
	assert.Equal(t, "/tmp/dojo-environment-multiline-testrunid", fs.FilesRemovals[2])
	assert.Equal(t, "/tmp/dojo-environment-bash-functions-testrunid", fs.FilesRemovals[0])
}

func fileExists(filePath string) bool {
	_, err := os.Lstat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(fmt.Sprintf("error when running os.Lstat(%q): %s", filePath, err))
	}
	return true
}

func TestDockerDriver_HandleRun_RealFileService(t *testing.T) {
	logger := NewLogger("debug")
	d := NewDockerDriver(NewMockedShellServiceNotInteractive(logger), NewFileService(logger), logger)
	config := getTestConfig()
	config.WorkDirOuter = "/tmp"
	config.RunCommand = ""
	runID := "testrunid"
	es := d.HandleRun(config, runID, NewEnvService())
	assert.Equal(t, 0, es)
	es = d.CleanAfterRun(config, runID)
	assert.Equal(t, 0, es)
	assert.False(t, fileExists("/tmp/dojo-environment-testrunid"))
}

func TestDockerDriver_HandleRun_RealEnvService(t *testing.T) {
	logger := NewLogger("debug")
	d := NewDockerDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	config := getTestConfig()
	config.RunCommand = ""
	runID := "testrunid"
	es := d.HandleRun(config, runID, NewEnvService())
	assert.Equal(t, 0, es)
	es = d.CleanAfterRun(config, runID)
	assert.Equal(t, 0, es)
	assert.False(t, fileExists("/tmp/dojo-environment-testrunid"))
}

func TestDockerDriver_HandlePull_Unit(t *testing.T) {
	logger := NewLogger("debug")
	d := NewDockerDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	config := getTestConfig()
	es := d.HandlePull(config)
	assert.Equal(t, 0, es)
}