package game

import (
	"errors"
	"time"
)

// SyncInput puts input in the queue and blocks until input can be processed.
// Returns false if input request was discarded
func (g *Game) SyncInput(input map[int]bool) error {
	for {
		select {
		case g.inpChan <- input:
			return nil
		case <-g.done:
			return errors.New("game over")
		}
	}
}

// Restart forcefully restarts the game
func (g *Game) Restart(gopherN int) {
	g.restart(gopherN)
}

type State struct {
	ID           int
	GophersState map[int]GopherState
	PipeBotY     float64
	PipeTopY     float64
}

type GopherState struct {
	PosYpercent float64
	SpeedY      float64
}

func (g *Game) CurState() (*State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mode == ModeGameOver {
		return nil, errors.New("game over")
	}
	return g.snapshotState(), nil
}

// NextState returns latest state of the game. Ensures new state for each request. Blocks until actual state is calculated.
// Returns error if there won't be any more new states
func (g *Game) NextState() (*State, error) {
	g.activeRequests.Add(1)
	defer g.activeRequests.Add(-1)
	select {
	case state := <-g.statesChan:
		return state, nil
	case <-g.done:
		return nil, errors.New("game over")
	}
}

func (g *Game) pushState() {
	if g.activeRequests.Load() == 0 {
		// no requests for this state
		// fmt.Println("active req is 0")
		if g.stepsPerUpdate > 1 {
			g.stepsPerUpdate--
		}
		return
	}
	state := g.snapshotState()
	// aviod locks anyway
	select {
	case g.statesChan <- state:
		// pushed state
	case <-time.After(time.Millisecond * 10):
		// fmt.Println("no requests for this state")
		// default:
		// 	fmt.Println("no req")

	}
}

func (g *Game) snapshotState() *State {
	state := &State{
		ID:           g.stepID,
		GophersState: make(map[int]GopherState, len(g.gophers)),
	}
	for _, gopher := range g.gophers {
		gopherState := GopherState{
			PosYpercent: float64(gopher.PosY()) / float64(g.windowH),
			SpeedY:      gopher.speedY / 100.0,
		}
		state.GophersState[gopher.ID] = gopherState
	}
	state.PipeBotY, state.PipeTopY = g.closestPipeYs()
	return state
}

func (g *Game) closestPipeYs() (float64, float64) {
	if len(g.pipesAhead) == 0 {
		return 0, 1
	}
	closest := g.pipesAhead[0]
	return float64(closest.PosTopY()) / float64(g.windowH), float64(closest.PosBotY()) / float64(g.windowH)
}
