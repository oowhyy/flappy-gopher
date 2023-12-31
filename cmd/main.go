package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"gographics/game"
	"gographics/neapy"
	_ "image/png"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
	hook "github.com/robotn/gohook"
	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

const (
	windowWidth  = 640
	windowHeight = 480
)

func main() {
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Flappy Gopher")
	ebiten.SetTPS(60)
	game := game.NewGame(windowWidth, windowHeight, 20)

	go runExperiment(game)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}

	// time.Sleep(time.Second * 10)
}

func humanPlayer(game *game.Game) {
	hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {
		go func() {
			_ = game.SyncInput(map[int]bool{1: true, 2: true})
		}()
	})
	go func() {
		s := hook.Start()
		fmt.Println("waiting clicks")
		<-hook.Process(s)
		fmt.Println("done hook")
	}()
}

func runExperiment(g *game.Game) {
	contextPath := "./data/flappy.neat.yaml"
	genomePath := "./data/flappy_start.yaml"
	experimentName := "Flappy"
	flag.Parse()

	// Load NEAT options
	neatOptions, err := neat.ReadNeatOptionsFromFile(contextPath)
	neat.LogLevel = neat.LogLevelInfo
	if err != nil {
		log.Fatal("Failed to load NEAT options: ", err)
	}

	// Load Genome
	log.Printf("Loading start genome for %s experiment from file '%s'\n", experimentName, genomePath)
	reader, err := genetics.NewGenomeReaderFromFile(genomePath)
	if err != nil {
		log.Fatalf("Failed to open genome file, reason: '%s'", err)
	}
	startGenome, err := reader.Read()

	if err != nil {
		log.Fatalf("Failed to read start genome, reason: '%s'", err)
	}
	fmt.Println(startGenome)

	if err != nil {
		log.Fatal("Failed to create output directory: ", err)
	}

	// create experiment
	expt := experiment.Experiment{
		Id:       0,
		Trials:   make(experiment.Trials, neatOptions.NumRuns),
		RandSeed: 123,
	}
	var generationEvaluator experiment.GenerationEvaluator

	expt.MaxFitnessScore = 25000000.0 // as given by fitness function definition
	generationEvaluator = neapy.NewFlappyEvaluator(g)

	// prepare to execute
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	fmt.Println("ready to execute")
	// run experiment in the separate GO routine
	go func() {
		if err = expt.Execute(neat.NewContext(ctx, neatOptions), startGenome, generationEvaluator, nil); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	// register handler to wait for termination signals
	//
	go func(cancel context.CancelFunc) {
		fmt.Println("\nPress Ctrl+C to stop")

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-signals:
			// signal to stop test fixture
			cancel()
		case err = <-errChan:
			// stop waiting
		}
	}(cancel)

	// Wait for experiment completion
	//
	err = <-errChan
	if err != nil {
		// error during execution
		log.Fatalf("Experiment execution failed: %s", err)
	}

	// Print experiment results statistics
	//
	expt.PrintStatistics()

	fmt.Printf(">>> Start genome file:  %s\n", genomePath)
	fmt.Printf(">>> Configuration file: %s\n", contextPath)

	// Save experiment data in native format
	//
	// expResPath := fmt.Sprintf("%s/%s.dat", outDir, *experimentName)
	// if expResFile, err := os.Create(expResPath); err != nil {
	// 	log.Fatal("Failed to create file for experiment results", err)
	// } else if err = expt.Write(expResFile); err != nil {
	// 	log.Fatal("Failed to save experiment results", err)
	// }

	// Save experiment data in Numpy NPZ format if requested
	//
	// npzResPath := fmt.Sprintf("%s/%s.npz", outDir, *experimentName)
	// if npzResFile, err := os.Create(npzResPath); err != nil {
	// 	log.Fatalf("Failed to create file for experiment results: [%s], reason: %s", npzResPath, err)
	// } else if err = expt.WriteNPZ(npzResFile); err != nil {
	// 	log.Fatal("Failed to save experiment results as NPZ file", err)
	// }
}
