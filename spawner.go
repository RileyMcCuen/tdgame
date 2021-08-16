package main

type (
	Spawner interface {
		Processor
		Start()
		Spawn() Enemy
	}
	GameSpawner struct {
		started bool
		cur     int
		rounds  []Spawner
	}
	RoundSpawner struct {
		cur     int
		enemies []Enemy
		*Ticker
	}
)

func NewGameSpawner(numRounds int, delay int, enemies []Enemy) Spawner {
	rounds := make([]Spawner, numRounds)
	for i := 0; i < numRounds; i++ {
		rounds[i] = NewRoundSpawner(delay, enemies)
	}
	return &GameSpawner{false, 0, rounds}
}

func (g *GameSpawner) Process(ticks int) {
	if g.started {
		g.rounds[g.cur].Process(ticks)
	}
}

func (g *GameSpawner) Done() bool {
	return !g.started || g.rounds[g.cur].Done()
}

func (g *GameSpawner) Spawn() Enemy {
	ret := g.rounds[g.cur].Spawn()
	if g.rounds[g.cur].Done() {
		g.started = false
		g.cur++
	}
	return ret
}

func (g *GameSpawner) Start() {
	g.started = true
}

func NewRoundSpawner(delay int, enemies []Enemy) Spawner {
	return &RoundSpawner{0, enemies, NewTicker(delay)}
}

func (r *RoundSpawner) Process(ticks int) {
	r.Tick()
}

func (r *RoundSpawner) Done() bool {
	return r.cur == len(r.enemies)
}

func (r *RoundSpawner) Ready() bool {
	return !r.Done() && r.Ticker.Done()
}

func (r *RoundSpawner) Spawn() Enemy {
	if r.Ready() {
		return r.enemies[r.cur]
	}
	return nil
}

func (r *RoundSpawner) Start() {}
