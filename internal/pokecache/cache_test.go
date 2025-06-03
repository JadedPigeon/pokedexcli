package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		name string
		key  string
		val  []byte
	}{
		{
			name: "Basic key 1",
			key:  "https://example.com",
			val:  []byte("testdata"),
		},
		{
			name: "Basic key 2",
			key:  "https://example.com/path",
			val:  []byte("moretestdata"),
		},
		{
			name: "Empty value",
			key:  "https://empty.com",
			val:  []byte(""),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key %s", c.key)
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected value '%s', got '%s'", string(c.val), string(val))
			}
		})
	}

	t.Run("Get missing key", func(t *testing.T) {
		cache := NewCache(interval)
		_, ok := cache.Get("https://not-in-cache.com")
		if ok {
			t.Errorf("expected to not find key")
		}
	})

	t.Run("Mutating original slice shouldn't change cached value", func(t *testing.T) {
		cache := NewCache(interval)
		original := []byte("mutable")
		cache.Add("mutate-key", original)
		original[0] = 'X'
		val, ok := cache.Get("mutate-key")
		if !ok {
			t.Errorf("expected to find mutate-key")
			return
		}
		if string(val) == string(original) {
			t.Errorf("expected cached value to remain unchanged")
		}
	})
}

func TestReapLoop_Immediate(t *testing.T) {
	interval := 10 * time.Millisecond
	cache := NewCache(interval)
	cache.Add("key", []byte("data"))

	time.Sleep(15 * time.Millisecond)
	_, ok := cache.Get("key")
	if ok {
		t.Errorf("expected key to be reaped but it still exists")
	}
}

func TestReapLoop_Partial(t *testing.T) {
	interval := 50 * time.Millisecond
	cache := NewCache(interval)

	cache.Add("old", []byte("olddata"))
	time.Sleep(60 * time.Millisecond) // Let "old" expire

	cache.Add("new", []byte("newdata"))
	time.Sleep(10 * time.Millisecond) // "new" should still be alive

	if _, ok := cache.Get("old"); ok {
		t.Errorf("expected 'old' key to be reaped")
	}
	if _, ok := cache.Get("new"); !ok {
		t.Errorf("expected 'new' key to still exist")
	}
}
