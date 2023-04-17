package threadmanager

type Manager struct {
	c chan int
}

func NewManager(maxRoutines int) *Manager {
	m := &Manager{
		c: make(chan int, maxRoutines),
	}
	for i := 0; i < maxRoutines; i++ {
		m.c <- i
	}
	return m
}

func (m *Manager) Lock() int {
	return <-m.c
}

func (m *Manager) Unlock(n int) {
	m.c <- n
}
