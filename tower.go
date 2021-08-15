package main

type (
	Tower struct {
		Spawner func() Particle
	}
)
