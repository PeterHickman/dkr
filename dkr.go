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

var dockerfile string

func usage() {
	fmt.Println("dkr [--file otherdocker.yml] build|run|connect|stop|delete")
	fmt.Println()
	fmt.Println("The optional --file flag will allow you to override the Dockerfile")
	fmt.Println()
	fmt.Println("  build   - Builds the image from Dockerfile")
	fmt.Println("  run     - Runs the container daemonised")
	fmt.Println("  connect - Connect to the container")
	fmt.Println("  stop    - Stops container")
	fmt.Println("  delete  - Delete container")
	fmt.Println()
	fmt.Println("You can chain commands => dkr build run connect")
	fmt.Println()
	fmt.Println("The dockerfile can contain various extra tags that will")
	fmt.Println("allow dkr to run things for you")
	fmt.Println()
	fmt.Println("  #RUN    -- Anything after this will be passed to the run command")
	fmt.Println("  #BUILD  -- Anything after this will be passed to the build command")
	fmt.Println("  #IGNORE -- Anything after this will be added to the .dockerignore file")
	fmt.Println("  #NAME   -- By default the container name is the same as the directory")
	fmt.Println("             but this tag will allow you to set a name")

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
		parts := strings.Fields(scanner.Text())
		if parts[0] == "#NAME" {
			name = strings.ToLower(parts[1])
		}
	}

	return name
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

func expose_tag(dockerfile string) string {
	ports := ""

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if parts[0] == "EXPOSE" {
			for _, port := range parts[1:] {
				ports += " -p " + port
				if !strings.Contains(port, ":") {
					ports += ":" + port
				}
			}
		}
	}

	return ports
}

func run_tag(dockerfile string) string {
	runtime := ""

	file, _ := os.Open(dockerfile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if parts[0] == "#RUN" {
			runtime += " " + strings.Join(parts[1:], " ")
		}
	}

	return runtime
}

func build_tag(dockerfile string) string {
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
		x := "docker container run -d --name " + name + " " + run_tag(dockerfile) + expose_tag(dockerfile) + " " + name

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

	x := "docker image build --file " + dockerfile + " -t " + name + " " + build_tag(dockerfile) + " ."

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

func init() {
	var d = flag.String("file", "Dockerfile", "Use this Dockerfile")
	flag.Parse()

	dockerfile = *d

	if !toolbox.FileExists(dockerfile) {
		fmt.Printf(ac.Red("There is no docker file called %s here\n\n"), dockerfile)
		usage()
	}
}

func main() {
	name := dockerfile_name(dockerfile)

	if len(flag.Args()) == 0 {
		usage()
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
			usage()
		}
	}
}
