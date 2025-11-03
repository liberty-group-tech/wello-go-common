package helper

import (
	"errors"
	"sync"
	"testing"
)

func TestNewLoader_Lazy(t *testing.T) {
	callCount := 0
	loader := NewLoader(func() (string, error) {
		callCount++
		return "lazy-value", nil
	})

	if callCount != 0 {
		t.Errorf("expected callCount to be 0, got %d", callCount)
	}

	value, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected callCount to be 1, got %d", callCount)
	}

	if value != "lazy-value" {
		t.Errorf("expected value to be 'lazy-value', got '%s'", value)
	}

	// Second call should not trigger reload
	value2, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected callCount to still be 1, got %d", callCount)
	}

	if value2 != "lazy-value" {
		t.Errorf("expected value2 to be 'lazy-value', got '%s'", value2)
	}
}

func TestNewLoader_Eager(t *testing.T) {
	callCount := 0
	loader := NewLoader(func() (string, error) {
		callCount++
		return "eager-value", nil
	}, WithMode(Eager))

	if callCount != 1 {
		t.Errorf("expected callCount to be 1 (eager load), got %d", callCount)
	}

	value, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected callCount to still be 1, got %d", callCount)
	}

	if value != "eager-value" {
		t.Errorf("expected value to be 'eager-value', got '%s'", value)
	}
}

func TestLoader_Get(t *testing.T) {
	loader := NewLoader(func() (int, error) {
		return 42, nil
	})

	value, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 42 {
		t.Errorf("expected value to be 42, got %d", value)
	}

	// Multiple calls should return same value
	for i := 0; i < 5; i++ {
		value2, err := loader.Get()
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
		if value2 != 42 {
			t.Errorf("expected value to be 42 on call %d, got %d", i, value2)
		}
	}
}

func TestLoader_Get_WithError(t *testing.T) {
	expectedErr := errors.New("test error")
	loader := NewLoader(func() (string, error) {
		return "", expectedErr
	})

	value, err := loader.Get()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if value != "" {
		t.Errorf("expected empty value on error, got '%s'", value)
	}

	// Error should be cached
	value2, err2 := loader.Get()
	if err2 != expectedErr {
		t.Errorf("expected cached error '%v', got '%v'", expectedErr, err2)
	}
	if value2 != "" {
		t.Errorf("expected empty value, got '%s'", value2)
	}
}

func TestLoader_Reload(t *testing.T) {
	callCount := 0
	loader := NewLoader(func() (int, error) {
		callCount++
		return callCount * 10, nil
	})

	// First call
	value1, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value1 != 10 {
		t.Errorf("expected value1 to be 10, got %d", value1)
	}
	if callCount != 1 {
		t.Errorf("expected callCount to be 1, got %d", callCount)
	}

	// Second call should return cached value
	value2, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value2 != 10 {
		t.Errorf("expected value2 to be 10, got %d", value2)
	}
	if callCount != 1 {
		t.Errorf("expected callCount to still be 1, got %d", callCount)
	}

	// Reload
	err = loader.Reload()
	if err != nil {
		t.Fatalf("unexpected error on reload: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected callCount to be 2 after reload, got %d", callCount)
	}

	// Get after reload should return new value
	value3, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value3 != 20 {
		t.Errorf("expected value3 to be 20, got %d", value3)
	}
}

func TestLoader_Reload_WithError(t *testing.T) {
	callCount := 0
	testErr := errors.New("test error")

	loader := NewLoader(func() (string, error) {
		callCount++
		if callCount == 1 {
			return "success", nil
		}
		return "", testErr
	})

	// First successful load
	value1, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value1 != "success" {
		t.Errorf("expected value1 to be 'success', got '%s'", value1)
	}

	// Reload with error
	err = loader.Reload()
	if err != testErr {
		t.Errorf("expected error '%v', got '%v'", testErr, err)
	}

	// Get after failed reload
	value2, err := loader.Get()
	if err != testErr {
		t.Errorf("expected error '%v', got '%v'", testErr, err)
	}
	if value2 != "" {
		t.Errorf("expected empty value on error, got '%s'", value2)
	}
}

func TestLoader_ConcurrentAccess(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	loader := NewLoader(func() (int, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		return 100, nil
	})

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			value, err := loader.Get()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if value != 100 {
				t.Errorf("expected value to be 100, got %d", value)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	actualCallCount := callCount
	mu.Unlock()

	if actualCallCount != 1 {
		t.Errorf("expected function to be called once, got %d times", actualCallCount)
	}
}

func TestLoader_WithMode_Default(t *testing.T) {
	callCount := 0
	loader := NewLoader(func() (string, error) {
		callCount++
		return "default", nil
	})

	// Default should be lazy
	if callCount != 0 {
		t.Errorf("expected callCount to be 0 (lazy default), got %d", callCount)
	}

	_, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected callCount to be 1 after Get, got %d", callCount)
	}
}

func TestLoader_StructType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	loader := NewLoader(func() (Person, error) {
		return Person{Name: "Alice", Age: 30}, nil
	})

	person, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if person.Name != "Alice" {
		t.Errorf("expected Name to be 'Alice', got '%s'", person.Name)
	}

	if person.Age != 30 {
		t.Errorf("expected Age to be 30, got %d", person.Age)
	}
}

func TestLoader_PointerType(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	loader := NewLoader(func() (*Config, error) {
		return &Config{Host: "localhost", Port: 8080}, nil
	})

	config, err := loader.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config to not be nil")
	}

	if config.Host != "localhost" {
		t.Errorf("expected Host to be 'localhost', got '%s'", config.Host)
	}

	if config.Port != 8080 {
		t.Errorf("expected Port to be 8080, got %d", config.Port)
	}
}

func TestLoader_MustGet(t *testing.T) {
	loader := NewLoader(func() (string, error) {
		return "success-value", nil
	})

	value := loader.MustGet()
	if value != "success-value" {
		t.Errorf("expected value to be 'success-value', got '%s'", value)
	}
}

func TestLoader_MustGet_WithError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, but none occurred")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("expected panic to be an error, got %T", r)
		}
		if err.Error() != "test error" {
			t.Errorf("expected panic error to be 'test error', got '%v'", err)
		}
	}()

	loader := NewLoader(func() (string, error) {
		return "", errors.New("test error")
	})

	_ = loader.MustGet()
	t.Fatal("should have panicked")
}

func TestWithMode(t *testing.T) {
	opt := WithMode(Eager)
	if opt == nil {
		t.Fatal("expected WithMode to return a non-nil option")
	}

	options := &loaderOptions{mode: Lazy}
	opt(options)

	if options.mode != Eager {
		t.Errorf("expected mode to be Eager, got %s", options.mode)
	}
}
