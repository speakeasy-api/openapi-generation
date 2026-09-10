//go:build !js || !wasm

package format

import (
	"sync"

	"github.com/tetratelabs/wazero/experimental"
)

// pooledAllocator reuses WASM linear memory buffers across module
// instantiations. For _start-based WASI modules (clangfmt, rubyfmt) that
// create a new instance per file, this eliminates tens of GB of short-lived
// allocations that would otherwise pressure the GC.
type pooledAllocator struct {
	mu   sync.Mutex
	bufs [][]byte
}

func (a *pooledAllocator) Allocate(capBytes, _ uint64) experimental.LinearMemory {
	a.mu.Lock()
	var buf []byte
	for i, b := range a.bufs {
		if uint64(len(b)) >= capBytes {
			buf = b
			a.bufs[i] = a.bufs[len(a.bufs)-1]
			a.bufs[len(a.bufs)-1] = nil
			a.bufs = a.bufs[:len(a.bufs)-1]
			break
		}
	}
	a.mu.Unlock()

	if buf == nil {
		// Start at the requested capacity (initial memory pages), not the
		// declared max (often 4 GB). The buffer will grow on demand via
		// Reallocate if the module calls memory.grow, and subsequent
		// instantiations reuse the grown buffer from the pool.
		buf = make([]byte, 0, capBytes)
	}

	return &pooledMemory{buf: buf[:0], alloc: a}
}

// pooledMemory implements experimental.LinearMemory backed by a reusable
// byte slice. On Free the buffer is returned to the parent pooledAllocator.
type pooledMemory struct {
	buf   []byte
	alloc *pooledAllocator
}

func (m *pooledMemory) Reallocate(size uint64) []byte {
	s := int(size)
	if s <= cap(m.buf) {
		prev := len(m.buf)
		m.buf = m.buf[:s]
		if s > prev {
			clear(m.buf[prev:s])
		}
		return m.buf
	}
	// Capacity exceeded — allocate a larger buffer with doubling growth
	// to amortize repeated memory.grow calls (important for rubyfmt which
	// grows memory incrementally during execution).
	newCap := 2 * cap(m.buf)
	if newCap < s {
		newCap = s
	}
	newBuf := make([]byte, s, newCap)
	copy(newBuf, m.buf)
	m.buf = newBuf
	return m.buf
}

func (m *pooledMemory) Free() {
	if m.buf != nil && m.alloc != nil {
		m.alloc.mu.Lock()
		m.alloc.bufs = append(m.alloc.bufs, m.buf[:cap(m.buf)])
		m.alloc.mu.Unlock()
		m.buf = nil
	}
}
