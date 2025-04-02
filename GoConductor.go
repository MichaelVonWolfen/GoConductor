package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
)

type ServiceConfig struct {
	MinSpawn  int `yaml:"min_spawn"`
	MaxSpawn  int `yaml:"max_spawn"`
	SpawnRule int `yaml:"spawn_rule"`
	KillRule  int `yaml:"kill_rule"`
}
type RunningService struct {
	ProcessId int
	Config    ServiceConfig
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func startService(serviceName string, serviceConfig ServiceConfig, runningServices *map[string][]RunningService) {
	fmt.Println(serviceName)
	e, err := os.Executable()
	//err = exec.Command("ls").Start()
	if err != nil {
		fmt.Println("Error starting the process:", err)
		return
	}
	cmd := exec.Command(fmt.Sprintf("%s/%s/%s", path.Dir(e), serviceName, serviceName))

	outputFile, err := os.Create(fmt.Sprintf("%s.output.txt", serviceName))
	if err != nil {
		fmt.Println("Error creating output file for service:", serviceName, err)
		return
	}
	cmd.Stdout = outputFile
	cmd.Stderr = outputFile

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting the process:", err)
		return
	}
	fmt.Println("Process started:", cmd.Process.Pid)
	time.Sleep(10 * time.Second)
	fmt.Println("Sending SIGTERM...")
	err = cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Println("Error sending SIGTERM:", err)
		return
	}
	// Wait for the process to exit gracefully
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Println("Process exited with error:", err)
		} else {
			fmt.Println("Process exited gracefully")
		}
	case <-time.After(5 * time.Second): // If it doesn't exit, force kill
		fmt.Println("Process did not exit in time, killing it...")
		err := cmd.Process.Kill()
		if err != nil {
			return
		}
	}

	fmt.Println("Process terminated")
}
func main() {
	// Define command-line flags with descriptions
	configPath := flag.String("configPath", "conductor.config.yaml", "Path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	check(err)
	var services map[string]ServiceConfig
	var runningServices map[string][]RunningService

	err = yaml.Unmarshal(data, &services)
	check(err)

	fmt.Println("%+v", services)

	/*
		TODO:
			- Initiate one process for each key with the configurations provided in the values
			- Monitor the initiated processes and create a new one or kill one on a need basis of rb
			(ratio between number of elements remained in the associated queue and the total number of processes running associated)
				- if rb > spawn_rule and max_spawn > current_nb_of_processes -> spawn another process
				- if rb < kill_rule and min_spawn < current_nb_of_processes -> send kill message to a process
				- if queue is empty -> kill all spawns
			- when application is shutdown kill all the services to ensure there are no orphans alive
	*/
	for k, v := range services {
		fmt.Printf("%s: %+v\n", k, v)
		startService(k, v, &runningServices)
	}

}
