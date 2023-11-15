package neapy

import (
	"context"
	"fmt"
	"gographics/game"

	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

const fitnessThreshold = 250000.0

type flappyEvaluator struct {
	gm *game.Game
}

func NewFlappyEvaluator(game *game.Game) experiment.GenerationEvaluator {
	return &flappyEvaluator{gm: game}
}

// GenerationEvaluate This method evaluates one epoch for given population and prints results into output directory if any.
func (e *flappyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	defer epoch.FillPopulationStatistics(pop)
	e.gm.Restart(len(pop.Organisms))
	options, ok := neat.FromContext(ctx)
	if !ok {
		return neat.ErrNEATOptionsNotFound
	}
	// while game is not over
	timeAlive := 0.0
	lastState := -1
	for {
		state, err := e.gm.NextState()
		if err != nil {
			break
		}
		if state.ID == lastState {
			fmt.Println(state.ID)
			continue
		}
		lastState = state.ID
		timeAlive += 0.1
		actions := make(map[int]bool, len(pop.Organisms))
		for i, agent := range pop.Organisms {
			gState, ok := state.GophersState[i]
			if !ok {
				// skip losers
				agent.Fitness = -10
				continue
			}
			pheno, err := agent.Phenotype()
			if err != nil {
				return err
			}
			depth, err := pheno.MaxActivationDepth()
			if err != nil {
				return err
			}
			// feed forward
			out := 0.0
			x := []float64{gState.PosYpercent, gState.SpeedY, state.PipeBotY, state.PipeTopY}
			if err := pheno.LoadSensors(x); err != nil {
				return err
			}
			if _, err := pheno.ForwardSteps(depth); err != nil {
				return err
			}
			out = pheno.ReadOutputs()[0]
			if _, err := pheno.Flush(); err != nil {
				return err
			}
			// perform actual action in game to progress through states
			if out > 0.5 {
				actions[i] = true
			}
			// calc fitness
			agent.Fitness = float64(timeAlive * timeAlive)
			// are we there yet?
			if agent.Fitness > fitnessThreshold {
				epoch.Solved = true
				agent.IsWinner = true
				epoch.Solved = true
				epoch.WinnerNodes = len(agent.Genotype.Nodes)
				epoch.WinnerGenes = agent.Genotype.Extrons()
				epoch.WinnerEvals = options.PopSize*epoch.Id + agent.Genotype.Id
				epoch.Champion = agent
				neat.InfoLog(fmt.Sprintf(">>>> Output activations: %e\n", out))
				return nil
			} else {
				agent.IsWinner = false
			}
		}
		err = e.gm.SyncInput(actions)
		if err != nil {
			break
			// return fmt.Errorf("sync input send failed: %w", err)
		}
	}
	return nil
}
