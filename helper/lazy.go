package helper

import (
	"log"
	"sync"
)

type LoadMode string

const (
	Lazy  LoadMode = "lazy"
	Eager LoadMode = "eager"
)

type loaderOptions struct {
	mode LoadMode
}

type Option func(*loaderOptions)

// WithMode sets the loading mode for the loader
func WithMode(mode LoadMode) Option {
	return func(o *loaderOptions) {
		o.mode = mode
	}
}

type Loader[T any] struct {
	once   sync.Once
	fn     func() (T, error)
	value  T
	err    error
	loaded bool
	mu     sync.RWMutex
}

// NewLoader creates a generic loader that supports lazy and eager loading modes
//
// Features:
//   - Generic support: can load instances of any type (e.g., database connections, clients)
//   - Lazy mode: loads on first Get() call (default)
//   - Eager mode: loads immediately when Loader is created
//   - Thread-safe: uses sync.Once to ensure initialization happens only once
//
// Parameters:
//   - fn: initialization function that returns an instance and error
//   - opts: optional options to configure the loader
//
// Example usage:
//
//	// Lazy loading (default)
//	loader := NewLoader(func() (*Database, error) {
//	    db := &Database{}
//	    return db, db.Connect()
//	})
//
//	// Eager loading
//	loader := NewLoader(func() (*Database, error) {
//	    db := &Database{}
//	    return db, db.Connect()
//	}, WithMode(Eager))
//
//	// Get instance
//	db, err := loader.Get()
//
//	// Reload
//	err := loader.Reload()
func NewLoader[T any](fn func() (T, error), opts ...Option) *Loader[T] {
	options := &loaderOptions{
		mode: Lazy, // default mode
	}

	for _, opt := range opts {
		opt(options)
	}

	loader := &Loader[T]{
		fn: fn,
	}

	if options.mode == Eager {
		loader.load()
	}

	return loader
}

func (l *Loader[T]) Get() (T, error) {
	l.once.Do(func() {
		l.load()
	})

	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.value, l.err
}

// MustGet returns the value and panics if Get returns an error
func (l *Loader[T]) MustGet() T {
	value, err := l.Get()
	if err != nil {
		panic(err)
	}
	return value
}

func (l *Loader[T]) Reload() error {
	l.mu.Lock()
	l.once = sync.Once{}
	l.loaded = false
	l.mu.Unlock()

	var err error
	l.once.Do(func() {
		l.mu.Lock()
		l.value, l.err = l.fn()
		l.loaded = true
		err = l.err
		l.mu.Unlock()

		if l.err != nil {
			log.Printf("failed to load value: %v", l.err)
		} else {
			log.Printf("value loaded successfully")
		}
	})

	return err
}

func (l *Loader[T]) load() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loaded {
		return
	}

	l.value, l.err = l.fn()
	l.loaded = true

	if l.err != nil {
		log.Printf("failed to load value: %v", l.err)
	} else {
		log.Printf("value loaded successfully")
	}
}
