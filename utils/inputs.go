package pkg

import (
	"io"
	"sync"
)

type StdinPool struct {
	mu       sync.RWMutex
	stdinMap map[string]*io.WriteCloser
}

func NewStdinConnection() StdinPool {
	return StdinPool{
		mu:       sync.RWMutex{},
		stdinMap: make(map[string]*io.WriteCloser),
	}
}

func (p *StdinPool) Set(key string, value *io.WriteCloser) {
	p.mu.Lock()
	p.stdinMap[key] = value
	p.mu.Unlock()
}

func (p *StdinPool) Get(key string) *io.WriteCloser{
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stdinMap[key]
}

func (p *StdinPool) Clear() {
	p.mu.Lock()
	p.stdinMap = make(map[string]*io.WriteCloser)
	p.mu.Unlock()
}
