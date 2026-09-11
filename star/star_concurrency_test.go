package star

import (
	"sync"
	"testing"
)

func hipparcosSpica() *Star {
	return &Star{
		RAHours:       201.29835230 / 15.0,
		DecDeg:        -11.16124491,
		RAMasPerYear:  -42.50,
		DecMasPerYear: -31.73,
		ParallaxMas:   12.44,
		Epoch:         2448349.0625, // J1991.25
	}
}

// Many goroutines making the first calls on one shared Star must all get the
// serial result. Run under the race detector (make test-race) as well.
func TestStar_ConcurrentFirstUse(t *testing.T) {
	want := hipparcosSpica().PositionAU(j2000)

	for trial := 0; trial < 50; trial++ {
		s := hipparcosSpica() // fresh: nothing computed yet
		start := make(chan struct{})
		const goroutines = 16
		got := make([][3]float64, goroutines)
		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				<-start
				got[g] = s.PositionAU(j2000)
			}(g)
		}
		close(start)
		wg.Wait()
		for g, p := range got {
			if p != want {
				t.Fatalf("trial %d goroutine %d: PositionAU = %v, want %v", trial, g, p, want)
			}
		}
	}
}

// A Star is plain data: a copy, before or after use, gives the same results.
func TestStar_CopyGivesSameResult(t *testing.T) {
	s := hipparcosSpica()
	before := *s
	want := s.PositionAU(j2000 + 1000)
	after := *s
	if got := before.PositionAU(j2000 + 1000); got != want {
		t.Errorf("copy made before use: %v, want %v", got, want)
	}
	if got := after.PositionAU(j2000 + 1000); got != want {
		t.Errorf("copy made after use: %v, want %v", got, want)
	}
}

func BenchmarkStar_PositionAU(b *testing.B) {
	s := hipparcosSpica()
	var sink [3]float64
	for i := 0; i < b.N; i++ {
		sink = s.PositionAU(j2000 + float64(i))
	}
	_ = sink
}
