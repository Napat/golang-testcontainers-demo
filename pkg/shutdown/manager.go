package shutdown

import (
	"context"
	"sync"
	"time"
)

type Manager struct {
	wg       sync.WaitGroup
	timeout  time.Duration
	handlers []func(context.Context) error
}

func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		timeout: timeout,
	}
}

func (m *Manager) AddHandler(handler func(context.Context) error) {
	m.handlers = append(m.handlers, handler)
}

func (m *Manager) Shutdown(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	m.wg.Add(len(m.handlers))
	errCh := make(chan error, len(m.handlers))

	for _, handler := range m.handlers {
		go func(h func(context.Context) error) {
			defer m.wg.Done()
			if err := h(timeoutCtx); err != nil {
				errCh <- err
			}
		}(handler)
	}

	// Wait for all handlers to complete or timeout
	doneCh := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	case err := <-errCh:
		return err
	case <-doneCh:
		return nil
	}
}
