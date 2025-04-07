package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"
	"time"
)

var TIME_BEFORE_KILL = 5 * time.Second

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

type SafeRunningServices struct {
	mu sync.Mutex
	rs map[string][]RunningService
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func addRunningService(serviceName string, service RunningService, runningServices *SafeRunningServices) {
	runningServices.mu.Lock()
	defer runningServices.mu.Unlock()
	runningServices.rs[serviceName] = append(runningServices.rs[serviceName], service)
}
func startService(serviceName string, serviceConfig ServiceConfig, runningServices *SafeRunningServices) {
	fmt.Printf("Service Name: %+v", serviceName)

	//Get path to current executed file
	e, err := os.Executable()
	if err != nil {
		fmt.Println("Error starting the process:", err)
		return
	}
	//Create the command to start the spawn in path.Dir/serviceName/serviceName
	cmd := exec.Command(fmt.Sprintf("%s/%s/%s", path.Dir(e), serviceName, serviceName))
	cmd.Args = append(cmd.Args, fmt.Sprintf("-name=%s", serviceName))

	writer := io.MultiWriter(os.Stdout)
	log.New(writer, "service: "+serviceName+": ", log.LstdFlags)
	cmd.Stdout = writer
	cmd.Stderr = writer
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting the process:", err)
		return
	}
	addRunningService(serviceName, RunningService{
		ProcessId: cmd.Process.Pid,
		Config:    serviceConfig,
	}, &*runningServices)

	fmt.Println("Process started:", cmd.Process.Pid)
}
func stopService(service RunningService, serviceName string) {
	proc, err := os.FindProcess(service.ProcessId)
	check(err)
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Println("Error sending SIGTERM:", err)
		return
	}
	//Wait for the process to exit gracefully
	done := make(chan error, 1)

	//Starts a go routine to stop the process
	go func() {
		_, err := proc.Wait()
		done <- err
	}()

	// Await for either the channel done to get an error ( with value or nil) or 5 seconds before it kills the process by force
	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("Process %d(%s) exited with error: \n%s", proc.Pid, serviceName, err)
		} else {
			fmt.Printf("Process %d(%s) exited gracefully\n", proc.Pid, serviceName)
		}
	case <-time.After(TIME_BEFORE_KILL):
		fmt.Printf("Process %d (%s) did not exit in time, killing it...\n", proc.Pid, serviceName)
		err := proc.Kill()
		if err != nil {
			fmt.Printf("Failed to forcefully kill process %d (%s): %v\n", service.ProcessId, serviceName, err)
		} else {
			fmt.Printf("Process %d (%s) killed forcefully\n", service.ProcessId, serviceName)
		}

	}
}
func main() {
	// Define command-line flags with descriptions
	configPath := flag.String("configPath", "conductor.config.yaml", "Path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	check(err)
	var services map[string]ServiceConfig
	var runningServices SafeRunningServices
	runningServices.rs = make(map[string][]RunningService)

	err = yaml.Unmarshal(data, &services)
	check(err)

	fmt.Println("%+v", services)

	/*
		TODO:
			-DONE: Initiate one process for each key with the configurations provided in the values
			- Monitor the initiated processes and create a new one or kill one on a need basis of rb
			(ratio between number of elements remained in the associated queue and the total number of processes running associated)
				- if rb > spawn_rule and max_spawn > current_nb_of_processes -> spawn another process
				- if rb < kill_rule and min_spawn < current_nb_of_processes -> send kill message to a process
				- if queue is empty -> kill all spawns
			- when application is shutdown kill all the services to ensure there are no orphans alive
	*/
	for k, v := range services {
		fmt.Printf("%s: %+v\n", k, v)
		fmt.Printf("\033[32m%+v\033[0m\n", runningServices.rs)
		startService(k, v, &runningServices)
	}
	time.Sleep(10 * time.Second)

	//runningServices.mu.Lock()
	//defer runningServices.mu.Unlock()
	for serviceName, v := range runningServices.rs {
		for _, service := range v {
			stopService(service, serviceName)
		}
		fmt.Printf("\u001B[32mKilled all processes for service: %s\n\u001B[0m", serviceName)
	}
}
