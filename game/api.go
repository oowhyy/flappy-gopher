package game

func (g *Game) SyncInput(input map[int]bool) bool {
	for {
		select {
		case g.inpChan <- input:
			return true
		case <-g.done:
			return false
		}
	}
}
