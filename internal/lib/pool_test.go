package lib

import (
	"sync"
	"testing"
)

// TestStruct - тестовая структура с методом Reset().
type TestStruct struct {
	ID    int
	Name  string
	Tags  []string
	Flags map[string]bool
}

func (t *TestStruct) Reset() {
	t.ID = 0
	t.Name = ""
	t.Tags = t.Tags[:0]
	clear(t.Flags)
}

func TestPool_GetPut(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Tags:  make([]string, 0, 10),
			Flags: make(map[string]bool),
		}
	})

	obj1 := pool.Get()
	if obj1 == nil {
		t.Fatal("Expected non-nil object from pool")
	}

	obj1.ID = 42
	obj1.Name = "test"
	obj1.Tags = append(obj1.Tags, "tag1", "tag2")
	obj1.Flags["active"] = true

	pool.Put(obj1)

	obj2 := pool.Get()
	if obj2 == nil {
		t.Fatal("Expected non-nil object from pool")
	}

	if obj2.ID != 0 {
		t.Errorf("Expected ID to be 0 after reset, got %d", obj2.ID)
	}
	if obj2.Name != "" {
		t.Errorf("Expected Name to be empty after reset, got %s", obj2.Name)
	}
	if len(obj2.Tags) != 0 {
		t.Errorf("Expected Tags to be empty after reset, got length %d", len(obj2.Tags))
	}
	if len(obj2.Flags) != 0 {
		t.Errorf("Expected Flags to be empty after reset, got length %d", len(obj2.Flags))
	}

	if cap(obj2.Tags) != 10 {
		t.Errorf("Expected Tags capacity to be preserved at 10, got %d", cap(obj2.Tags))
	}
}

func TestPool_MultiplePutGet(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Tags:  make([]string, 0),
			Flags: make(map[string]bool),
		}
	})

	objects := make([]*TestStruct, 10)

	for i := 0; i < 10; i++ {
		obj := pool.Get()
		obj.ID = i
		obj.Name = "test"
		objects[i] = obj
	}

	for _, obj := range objects {
		pool.Put(obj)
	}

	for i := 0; i < 10; i++ {
		obj := pool.Get()
		if obj.ID != 0 {
			t.Errorf("Object %d: Expected ID to be 0, got %d", i, obj.ID)
		}
		if obj.Name != "" {
			t.Errorf("Object %d: Expected Name to be empty, got %s", i, obj.Name)
		}
	}
}

func TestPool_Concurrent(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Tags:  make([]string, 0, 5),
			Flags: make(map[string]bool),
		}
	})

	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				obj := pool.Get()

				if obj.ID != 0 || obj.Name != "" || len(obj.Tags) != 0 || len(obj.Flags) != 0 {
					t.Errorf("Goroutine %d, iteration %d: Object not properly reset", id, j)
				}

				obj.ID = id
				obj.Name = "test"
				obj.Tags = append(obj.Tags, "tag")
				obj.Flags["used"] = true

				pool.Put(obj)
			}
		}(i)
	}

	wg.Wait()
}

func TestPool_CapacityPreservation(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Tags:  make([]string, 0, 100),
			Flags: make(map[string]bool, 50),
		}
	})

	obj := pool.Get()

	if cap(obj.Tags) != 100 {
		t.Errorf("Expected initial Tags capacity 100, got %d", cap(obj.Tags))
	}

	for i := 0; i < 50; i++ {
		obj.Tags = append(obj.Tags, "tag")
	}

	pool.Put(obj)
	obj2 := pool.Get()

	if cap(obj2.Tags) != 100 {
		t.Errorf("Expected Tags capacity 100 after reset, got %d", cap(obj2.Tags))
	}

	if len(obj2.Tags) != 0 {
		t.Errorf("Expected Tags length 0 after reset, got %d", len(obj2.Tags))
	}
}
