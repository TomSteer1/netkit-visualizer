package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"os/exec"
	"bytes"
  "github.com/goccy/go-graphviz"
  "github.com/goccy/go-graphviz/cgraph"
)


var machines = make(map[string]Machine)
var networks = make(map[string]Network)

func main() {
	fmt.Fprintln(os.Stderr,"Netkit Visualizer")
	fmt.Fprintln(os.Stderr,"Version 0.1")
	fmt.Fprintln(os.Stderr,"")
	fmt.Fprintln(os.Stderr,"Checking for config file...")
	if testForFile("lab.conf") {
		mapMachines()
		menuOptions := []string{"List Networks", "Create Graph"}
		for option := range menuOptions {
			fmt.Fprintln(os.Stderr,option+1, menuOptions[option])
		}
		fmt.Fprintln(os.Stderr,"Please select an option:")
		var input string
		fmt.Scanln(&input)
		switch input {
			case "1":
				listNetworks()
			case "2":
				createGraph()
			default:
					fmt.Fprintln(os.Stderr,"Invalid option")
		}
	} else {
		fmt.Fprintln(os.Stderr,"Config file not found. Exiting.")
		os.Exit(1)
	}
	os.Exit(0)
}

type Machine struct {
	Name string
	Cards map[string]Card
	Node *cgraph.Node
}

type Card struct {
	Name string
	IP string
	Network Network
}

type Network struct {
	Name string
	Machines []Machine
	Node *cgraph.Node
}

func mapMachines() {
	fmt.Fprintln(os.Stderr,"Loading Config File...")
	configFile, err := os.Open("lab.conf")
	HandleError(err)
	defer configFile.Close()
	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 0 && text[0] != '#' {
			if text[0:3] != "LAB" {
				name := strings.Split(text, "[")[0]
				var machine Machine
				if _, ok := machines[name]; ok {
					machine = machines[name]
				} else {
					machine = Machine{Name: name, Cards: make(map[string]Card)}
				}
				cardID := strings.Split(strings.Split(text, "[")[1], "]")[0]
				networkID := strings.Split(text, "]=")[1]
				var network Network
				if _, ok := networks[networkID]; ok {
					network = networks[networkID]
				} else {
					network = Network{Name: networkID, Machines: []Machine{}}
				}
				card := Card{Name: cardID, Network: network}
				machine.Cards[networkID] = card
				network.Machines = append(network.Machines, machine)
				machines[name] = machine
				networks[networkID] = network
				getMachineInfo(name)
			}
		}
	}
}

func getMachineInfo(machineName string) {
	file , _ := os.Open(machineName + ".startup")	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "ip addr add") {
			cardID := strings.Split(strings.Split(strings.Split(text, "dev ")[1], " ")[0],"eth")[1]
			ip := strings.Split(strings.Split(text, "ip addr add ")[1], "/")[0]
			for networkID, card := range machines[machineName].Cards {
				if card.Name == cardID {
					card.IP = ip
					machines[machineName].Cards[networkID] = card
				}
			}
		}
	}
}

func listNetworks() {
	fmt.Println("Listing networks...")
	for _, network := range networks {
		fmt.Println(network.Name)
		for _, machine := range network.Machines {
			fmt.Println("\t" + machine.Name)
			fmt.Println("\t\t" + machine.Cards[network.Name].IP)
		}
	}
}

func createGraph() {
	fmt.Fprintln(os.Stderr,"Creating Graph...")
	g := graphviz.New()
	graph, _ := g.Graph()
	for key, machine := range machines {
//		g.AddVertex(machine.Name,graph.VertexAttribute("shape", "box"))
		m , _ := graph.CreateNode(machine.Name)
		machine.Node = m
		m.Set("shape", "box")
		machines[key] = machine
	}
	for key, network := range networks {
//		g.AddVertex(network.Name, graph.VertexAttribute("color","blue"))
		n, _ :=graph.CreateNode(network.Name)
		network.Node = n
		networks[key] = network
//		color := getNextColor()
		for _, machine := range network.Machines {
					graph.CreateEdge(machine.Cards[network.Name].IP, machines[machine.Name].Node, network.Node)
//				g.AddEdge(machine.Name, network.Name, graph.EdgeAttribute("color",color), graph.EdgeAttribute("label",machine.Cards[network.Name].IP))
		}
	}
	var buf bytes.Buffer
	g.Render(graph, "dot", &buf)
//	image , _ := g.RenderImage(graph)
	g.RenderFilename(graph, graphviz.PNG, "graph.png")

	exec.Command("feh", "graph.png").Run()
	exec.Command("rm", "graph.png").Run()

}

var colors = []string{"red", "blue", "green", "yellow", "orange", "purple", "pink", "brown", "grey", "black"}

func getNextColor() string {
	if len(colors) > 0 {
		color := colors[0]
		colors = colors[1:]
		return color
	}	else {
		return "black"
	}

}

func testForFile(fileName string) (bool) {
	_, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return true
}

func contains(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func HandleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr,err)
		os.Exit(1)
	}
}
