package neat

import (
	"context"

	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
)

const fitnessThreshold = 15.5

type flappyEvaluator struct {
}

func NewFlappyEvaluator(outputPath string) experiment.GenerationEvaluator {
	return &flappyEvaluator{}
}

// GenerationEvaluate This method evaluates one epoch for given population and prints results into output directory if any.
func (e *flappyEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {

	return nil
}

// orgEvaluate evaluates fitness of the provided organism
func (e *flappyEvaluator) orgEvaluate(organism *genetics.Organism) (bool, error) {
	return false, nil
}
