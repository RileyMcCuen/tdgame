package td

import "tdgame/core"

type (
	Round struct {
		core.GameObjectNoop
		Cur, Delay, Round, Points int
		T                         *core.Ticker
		Enemies                   []Enemy
	}
)

var _ core.GameObject = (*Round)(nil)

func (r *Round) Process(ticks int, con core.Context) bool {
	r.T.Tick()
	return false
}

func (r *Round) Done() bool {
	return r.Cur == len(r.Enemies)
}

func (r *Round) Spawn() Enemy {
	if r.T.Done() {
		ret := r.Enemies[r.Cur]
		r.Cur++
		r.T = core.NewTicker(r.Delay)
		return ret
	}
	return nil
}
