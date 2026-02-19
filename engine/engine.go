package engine

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/GiorgosMarga/satellites/node"
)

const (
	MaxCommDistance = 3000 // km
)

type Engine struct {
	Nodes map[int]*node.Node
}

func New() *Engine {
	return &Engine{
		Nodes: make(map[int]*node.Node),
	}
}

func (eng *Engine) reset() {
	for _, n := range eng.Nodes {
		n.Reset()
	}
}

func (eng *Engine) readSnapshot(filepath string) error {
	eng.reset()
	f, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	cleanedLines := bytes.ReplaceAll(f, []byte{'\r'}, []byte{})
	lines := bytes.Split(cleanedLines, []byte{'\n'})

	satId := 1
	for _, l := range lines {
		if len(l) == 0 {
			continue
		}
		line := string(l)
		splittedLine := strings.Split(line, " ")
		if len(splittedLine) != 3 {
			return fmt.Errorf("expected 3 values received %d\n", len(splittedLine))
		}

		var (
			X, Y, Z float64
			err     error
		)

		if X, err = strconv.ParseFloat(splittedLine[0], 64); err != nil {
			return err
		}
		if Y, err = strconv.ParseFloat(splittedLine[1], 64); err != nil {
			return err
		}
		if Z, err = strconv.ParseFloat(splittedLine[2], 64); err != nil {
			return err
		}
		pos := []float64{X, Y, Z}
		// newNode := node.New(satId, pos)

		var newNode *node.Node
		if _, ok := eng.Nodes[satId]; ok {
			newNode = eng.Nodes[satId]
			newNode.Update(pos)
		} else {
			newNode = node.New(satId, pos)
		}
		satId++
		for _, existingNode := range eng.Nodes {
			if existingNode.ID < newNode.ID {
				dist := node.GetEuclidianDistance(existingNode.Position, newNode.Position)
				if dist < MaxCommDistance {
					existingNode.AddPeer(newNode)
					newNode.AddPeer(existingNode)
				}
			}
		}
		eng.Nodes[newNode.ID] = newNode
	}

	return nil
}

func (eng *Engine) Start(filepath string) error {
	dir, err := os.ReadDir(filepath)
	if err != nil {
		return err
	}

	log.Printf("Reading folder: %s\n", filepath)

	for idx, f := range dir {
		if idx == 10 {
			break
		}
		fileName := path.Join(filepath, f.Name())
		fmt.Println("Running", fileName)
		if err := eng.readSnapshot(fileName); err != nil {
			return err
		}

		// wg := &sync.WaitGroup{}
		for _, n := range eng.Nodes {
			// wg.Go(n.Start)
			go n.Start()
		}
		<-time.After(10 * time.Second)

		for _, n := range eng.Nodes {
			n.Stop()
		}
		if err := eng.logResults(fileName); err != nil {
			fmt.Println(err)
		}
		// return nil
	}
	return nil
}

func (eng *Engine) logResults(filename string) error {

	path := path.Join("engLogs", "clusters", path.Base(filename))

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, n := range eng.Nodes {
		for _, peer := range n.Neighbors {
			if n.ID < peer.ID {
				fmt.Fprintf(f, "%d-%d\n", n.ID, peer.ID)
			}
		}
	}
	fmt.Fprintf(f, "\n\n")
	clusters := make(map[int][]int, 0)

	for _, n := range eng.Nodes {
		fmt.Fprintf(f, "%d->%d\n", n.ID, n.State.ClusterId)
		clusters[n.State.ClusterId] = append(clusters[n.State.ClusterId], n.ID)
	}

	for clusterId, cluster := range clusters {
		fmt.Println(clusterId, cluster)
		if !slices.Contains(cluster, clusterId-1) || !slices.Contains(cluster, clusterId+1) {
			fmt.Printf("%d doesnt contain next and prev\n", clusterId)
		}
	}
	return nil
}
