package streams

import (
	"fmt"
	"reflect"
	"testing"
)

func TestShouldCollectEmptyToEmpty(t *testing.T) {
	s := NewFromSlice([]int{})
	c, _ := Collect(s)

	if len(c) != 0 {
		t.Error(`Didn't collect empty`)
	}
}

func TestShouldCollect(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	c, _ := Collect(s)

	if !reflect.DeepEqual([]int{3, 1, 4}, c) {
		t.Error(`Didn't collect`)
	}
}

func TestShouldMapEmpty(t *testing.T) {
	s := NewFromSlice([]int{})
	s = Map(s, func(v int) (int, error) {
		return v + 1, nil
	})

	c, _ := Collect(s)

	if len(c) != 0 {
		t.Error(`Didn't Map Empty`)
	}
}

func TestShouldMap(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Map(s, func(v int) (int, error) {
		return v + 1, nil
	})

	c, _ := Collect(s)

	if !reflect.DeepEqual(c, []int{4, 2, 5}) {
		t.Error(`Didn't Map`)
	}
}

func TestShouldMapOnNilAsEmptyStream(t *testing.T) {
	s := (*Mapper[int, int])(nil)

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't Map on nil`)
	}
}

func TestShouldMapOnZeroValueAsEmptyStream(t *testing.T) {
	s := &Mapper[int, int]{}

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't Map on zero value`)
	}
}

func TestMapShouldEosOnError(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Map(s, func(v int) (int, error) {
		if v%2 == 0 {
			return 0, fmt.Errorf("error")
		}
		return v, nil
	})

	c, _ := Collect(s)

	if !reflect.DeepEqual(c, []int{3, 1}) {
		t.Error(`Didn't Map to eos on error`)
	}
}

func TestMapShouldErrorOnError(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Map(s, func(v int) (int, error) {
		if v%2 == 0 {
			return 0, fmt.Errorf("error")
		}
		return v, nil
	})

	_, err := Collect(s)

	if err == nil {
		t.Error(`Didn't Map error on error`)
	}
}

func TestShouldFlatMap(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4, 1})
	ss := Windowed(s, 2, 2)
	sl := FlatMap(ss, func(v int) (int, error) {
		return v, nil
	})

	c, _ := Collect(sl)

	if !reflect.DeepEqual(c, []int{3, 1, 1, 4, 4, 1}) {
		t.Error(`Didn't FlatMap`)
	}
}

func TestShouldFlatMapOnNilAsEmptyStream(t *testing.T) {
	s := (*FlatMapper[int, int])(nil)

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't FlatMap on nil`)
	}
}

func TestShouldFlatMapOnZeroValueAsEmptyStream(t *testing.T) {
	s := &FlatMapper[int, int]{}

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't FlatMap on zero value`)
	}
}

func TestShouldDrop(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Drop(s, 2)

	c, _ := Collect(s)

	if !reflect.DeepEqual(c, []int{4}) {
		t.Error(`Didn't Drop`)
	}
}

func TestShouldDropAll(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Drop(s, 3)

	c, _ := Collect(s)

	if len(c) != 0 {
		t.Error(`Didn't Drop all`)
	}
}

func TestShouldDropOnNilAsEmptyStream(t *testing.T) {
	s := (*Dropper[int])(nil)

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't Drop on nil`)
	}
}

func TestShouldDropOnZeroValueAsEmptyStream(t *testing.T) {
	s := &Dropper[int]{}

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't Map on zero value`)
	}
}

func TestShouldTruncate(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Truncate(s, 2)

	c, _ := Collect(s)

	if !reflect.DeepEqual(c, []int{3}) {
		t.Error(`Didn't Truncate`)
	}
}

func TestShouldTruncateAll(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = Truncate(s, 3)

	c, _ := Collect(s)

	if len(c) != 0 {
		t.Error(`Didn't Truncate all`)
	}
}

func TestShouldDiff(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4, 1})
	s = DiffN(s, 2)

	c, _ := Collect(s)

	if !reflect.DeepEqual(c, []int{1, 0}) {
		t.Error(`Didn't Diff`)
	}
}

func TestShouldDiffAll(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4})
	s = DiffN(s, 3)

	c, _ := Collect(s)

	if len(c) != 0 {
		t.Error(`Didn't Diff all`)
	}
}

func TestShouldAccumulate(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4, 1})

	a, _ := Accumulate(s, 0, func(a, b int) int { return a + b })

	if a != 9 {
		t.Error(`Didn't Accumulate`)
	}
}

func TestShouldAccumulateOnEmpty(t *testing.T) {
	s := NewFromSlice([]int{})

	a, _ := Accumulate(s, 1, func(a, b int) int { return a + b })

	if a != 1 {
		t.Error(`Didn't Accumulate on empty`)
	}
}

func TestShouldCount(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4, 1})

	c, _ := Count(s)

	if c != 4 {
		t.Error(`Didn't Count`)
	}
}

func TestShouldCountOnEmpty(t *testing.T) {
	s := NewFromSlice([]int{})

	c, _ := Count(s)

	if c != 0 {
		t.Error(`Didn't Count on empty`)
	}
}

func TestShouldFilter(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4, 1})
	s = Filter(s, func(v int) bool { return v%2 != 0 })

	c, _ := Collect(s)

	if !reflect.DeepEqual(c, []int{3, 1, 1}) {
		t.Error(`Didn't Filter`)
	}
}

func TestShouldFilterOnEmpty(t *testing.T) {
	s := NewFromSlice([]int{})
	s = Filter(s, func(v int) bool { return v%2 != 0 })

	c, _ := Collect(s)

	if len(c) != 0 {
		t.Error(`Didn't Filter on empty`)
	}
}

func TestShouldFilterOnNil(t *testing.T) {
	s := (*Filterer[int])(nil)

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't Filter on nil`)
	}
}

func TestShouldFilterOnZeroValue(t *testing.T) {
	s := Filter(nil, func(v int) bool { return false })

	eos, _, _ := s.Resolve(func(v int) error { return nil })

	if !eos {
		t.Error(`Didn't Filter on zero value`)
	}
}

func TestShouldWindow(t *testing.T) {
	s := NewFromSlice([]int{3, 1, 4, 1})
	ss := Windowed(s, 2, 2)
	sl := Map(ss, func(v Stream[int]) ([]int, error) {
		return Collect(v)
	})

	c, _ := Collect(sl)

	if !reflect.DeepEqual(c, [][]int{{3, 1}, {1, 4}, {4, 1}}) {
		t.Error(`Didn't Window`)
	}
}

func TestShouldWindowOnEmpty(t *testing.T) {
	s := NewFromSlice([]int{})
	ss := Windowed(s, 2, 1)
	sl := Map(ss, func(v Stream[int]) ([]int, error) {
		return Collect(v)
	})

	c, _ := Collect(sl)

	if len(c) != 0 {
		t.Error(`Didn't Window on empty`)
	}
}

func TestShouldWindowOnNil(t *testing.T) {
	s := (*Windower[int])(nil)

	eos, _, _ := s.Resolve(func(v Stream[int]) error { return nil })

	if !eos {
		t.Error(`Didn't Window on nil`)
	}
}

func TestShouldWindowOnZeroValue(t *testing.T) {
	s := Windowed[int](nil, 1, 1)

	eos, _, _ := s.Resolve(func(v Stream[int]) error { return nil })

	if !eos {
		t.Error(`Didn't Window on zero value`)
	}
}
