package main

import (
	"bufio"
	"flag"
	"fmt"
	ac "github.com/PeterHickman/ansi_colours"
	"github.com/PeterHickman/toolbox"
	"os"
	"path/filepath"
	"strings"
)

func usage(name string) {
	fmt.Println("dkr build|run|connect|stop|delete")
	fmt.Println()
	fmt.Printf("  build   - Builds the image [%s] from Dockerfile\n", name)
	fmt.Printf("  run     - Runs the container [%s] daemonised\n", name)
	fmt.Printf("  connect - Connect to the container [%s]\n", name)
	fmt.Printf("  stop    - Stops container [%s]\n", name)
	fmt.Printf("  delete  - Delete container [%s]\n", name)
	fmt.Println()
	fmt.Println("You can chain commands => dkr build run connect")

	os.Exit(1)
}

func dockerfile_name(dockerfile string) string {
	var name = ""
	dir, _ := os.Getwd()
	name = strings.ToLower(filepath.Base(dir))

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#NAME") {
			name = strings.Fields(text)[1]
		}
	}

	return name
}

func dockerfile_run(dockerfile string) string {
	var runtime []string

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#RUN") {
			runtime = append(runtime, strings.Join(strings.Fields(text)[1:], " "))
		}
	}

	return strings.Join(runtime, " ")
}

func dockerfile_expose(dockerfile string) []string {
	ports := []string{}

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "EXPOSE") {
			ports = append(ports, strings.Join(strings.Fields(text)[1:], " "))
		}
	}

	return ports
}

func dockerfile_volumes(dockerfile string) []string {
	volumes := []string{}

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#VOLUME") {
			volumes = append(volumes, strings.Join(strings.Fields(text)[1:], " "))
		}
	}

	return volumes
}

func dockerfile_env(dockerfile string) []string {
	envs := []string{}

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#ENV") {
			envs = append(envs, strings.Join(strings.Fields(text)[1:], " "))
		}
	}

	return envs
}

func dockerfile_build(dockerfile string) string {
	build := ""

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#BUILD") {
			build += " " + strings.Join(strings.Fields(text)[1:], " ")
		}
	}

	return build
}

func dockerfile_ignore(dockerfile string) {
	ignores := []string{}

	file, _ := os.Open(dockerfile)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#IGNORE") {
			ignores = append(ignores, strings.Fields(text)[1])
		}
	}

	file.Close()

	if len(ignores) == 0 {
		return
	}

	if toolbox.FileExists(".dockerignore") {
		file, _ := os.Open(".dockerignore")
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ignores = append(ignores, scanner.Text())
		}
	}

	file, _ = os.Create(".dockerignore")
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range ignores {
		fmt.Fprintln(w, line)
	}
}

func find(cmd, name string) bool {
	out, _ := toolbox.CommandOutput("docker container " + cmd)

	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if parts[1] == name {
			return true
		}
	}

	return false
}

func image_available(name string) bool {
	out, _ := toolbox.CommandOutput("docker image ls")

	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if parts[0] == name && parts[1] == "latest" {
			return true
		}
	}

	return false
}

func run_container(name, dockerfile string) {
	if image_available(name) {
		x := "docker container run -d --name " + name
		x += " " + dockerfile_run(dockerfile)

		for _, port := range dockerfile_expose(dockerfile) {
			if strings.Contains(port, ":") {
				x += " -p " + port
			} else {
				x += " -p " + port + ":" + port
			}
		}

		for _, volume := range dockerfile_volumes(dockerfile) {
			x += " -v " + volume
		}

		for _, env := range dockerfile_env(dockerfile) {
			x += " -e " + env
		}

		x += " " + name

		fmt.Println(ac.Bold("==> Running ") + ac.Blue(name))

		toolbox.Command(x)
	} else {
		fmt.Println(ac.Red("There is no image ") + ac.Blue(name+":latest") + ac.Red(" available. Did you build?"))
		os.Exit(1)
	}
}

func stop_container(name string) {
	if find("ps", name) {
		fmt.Println(ac.Bold("==> Stopping " + ac.Blue(name)))
		toolbox.Command("docker container stop " + name)
	} else {
		fmt.Println(ac.Bold("==> ") + ac.Blue(name) + ac.Bold(" is not running"))
	}
}

func build_container(name, dockerfile string) {
	dockerfile_ignore(dockerfile)

	fmt.Println(ac.Bold("==> Building ") + ac.Blue(name))

	x := "docker image build --file " + dockerfile + " -t " + name + " " + dockerfile_build(dockerfile) + " ."

	toolbox.Command(x)
}

func connect_container(name string) {
	fmt.Println(ac.Bold("==> Connecting to ") + ac.Blue(name))

	x := "docker container exec -it " + name + " /bin/bash"

	toolbox.Command(x)
}

func delete_container(name string) {
	if find("ls -a", name) {
		fmt.Println(ac.Bold("==> Deleting ") + ac.Blue(name))

		x := "docker container rm --force " + name

		toolbox.Command(x)
	} else {
		fmt.Println(ac.Bold("==> ") + ac.Blue(name) + ac.Bold(" is not there"))
	}
}

func main() {
	var override = flag.String("file", "", "Use this Dockerfile")
	var dockerfile = "Dockerfile"

	flag.Parse()

	if *override != "" {
		dockerfile = *override
	}

	if !toolbox.FileExists(dockerfile) {
		fmt.Printf(ac.Red("There is no docker file called %s here\n"), dockerfile)
		os.Exit(1)
	}

	name := dockerfile_name(dockerfile)

	if len(flag.Args()) == 0 {
		usage(name)
	}

	for _, cmd := range flag.Args() {
		switch strings.ToLower(cmd) {
		case "build":
			build_container(name, dockerfile)
		case "run":
			run_container(name, dockerfile)
		case "connect":
			connect_container(name)
		case "stop":
			stop_container(name)
		case "delete":
			delete_container(name)
		default:
			usage(name)
		}
	}
}
