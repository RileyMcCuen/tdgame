package core

type (
	Ticker struct{ max, cur int }
)

func (t *Ticker) Max() int {
	return t.max
}

func (t *Ticker) Tick() bool {
	t.cur++
	return t.Done()
}

func (t *Ticker) TickBy(amount int) bool {
	t.cur = MinInt(t.max, t.cur+amount)
	return t.Done()
}

func (t *Ticker) Ticks() int {
	return t.cur
}

func (t *Ticker) Done() bool {
	return t.cur == t.max
}

func (t *Ticker) Reset() {
	t.cur = 0
}

func NewTicker(cur int) *Ticker {
	return &Ticker{cur, 0}
}
