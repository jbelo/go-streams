package streams

import (
	"bufio"
	"fmt"
	"golang.org/x/exp/constraints"
	"os"
)

// A Stream of `T`s is either the empty stream or an element of type T,
// followed by a stream of `T`s. In other words, conceptually, a Stream
// of `T`s is a, possibly infinite, sequence of elements of type T.
//
// Streams must support a `Resolve` operation, which decomposes the stream
// into its constituent parts, the *head* element and the *remainder* stream.
// For this, the operation is parameterized by a function `h` that is to be
// applied to the head element of the stream.
// `Resolve` results in a boolean indicating an *end of stream* condition.
// When true, the stream is in fact the empty stream and there is actually
// no head element to apply the function `h` to. Otherwise, if false, then
// either there is a known head element, to which the function `h` was
// applied to, or not. If not, then there isn't yet a known head element,
// and the call to `Resolve` must have at least made progress in determining
// the end of stream condition and the next head element, if any. The call
// may thus be a noop from the caller's point of view.
// `Resolve` also returns the stream object to be used for the next call,
// and possibly an error.
//
// For each type that implements the `Stream` interface, both `nil` and the
// zero value should be valid representations of the empty stream.
//
// If there is an error, an end of stream condition is signaled, along with
// the error. Notice that an error may be *internal* or *external*, meaning,
// it may originate from code in the stream library, or code in the stream
// processing, respectively. These situations should be distinguishable on
// the client.
type Stream[T any] interface {
	Resolve(func(v T) error) (bool, Stream[T], error)
}

// A Mapper represents the stream that results from applying a given function
// `f` to each element of a given base stream. The base stream has elements
// of type `T`, and the Mapper has elements of type `U`. The `Resolve` operation
// decomposes this latter stream of `U`s.
type Mapper[T, U any] struct {
	base Stream[T]
	f    func(T) (U, error)
}

func (s *Mapper[T, U]) Resolve(h func(U) error) (bool, Stream[U], error) {
	if s == nil || s.base == nil {
		return true, nil, nil
	}

	eos, nxs, err := s.base.Resolve(func(v T) error {
		u, e := s.f(v)
		if e != nil {
			return e
		}

		e = h(u)

		return e
	})

	s.base = nxs

	if err != nil {
		return true, s, err
	}

	return eos, s, nil
}

func Map[T, U any](s Stream[T], f func(T) (U, error)) Stream[U] {
	return &Mapper[T, U]{base: s, f: f}
}

// A Dropper represents the stream that results from "dropping" the
// first few elements from a given stream.
type Dropper[T any] struct {
	base Stream[T]
	n, c int
}

func (s *Dropper[T]) Resolve(h func(u T) error) (bool, Stream[T], error) {
	if s == nil || s.base == nil {
		return true, nil, nil
	}

	eos, nxs, err := s.base.Resolve(func(v T) error {
		if s.c < s.n {
			s.c++

			return nil
		}

		s.c++

		return h(v)
	})

	s.base = nxs

	if err != nil {
		return true, s, err
	}

	return eos, s, nil
}

func Drop[T any](s Stream[T], n int) Stream[T] {
	return &Dropper[T]{base: s, n: n}
}

type Truncater[T any] struct {
	base Stream[T]
	hold []T
	n, i int
}

func Truncate[T any](s Stream[T], n int) Stream[T] {
	return &Truncater[T]{base: s, hold: make([]T, 0, n), n: n}
}

func (s *Truncater[T]) Resolve(h func(v T) error) (bool, Stream[T], error) {
	eos, nxs, err := s.base.Resolve(func(v T) error {
		if len(s.hold) < cap(s.hold) {
			s.hold = append(s.hold, v)

			return nil
		}
		head := s.hold[s.i]
		s.hold[s.i] = v
		s.i++
		if s.i == len(s.hold) {
			s.i = 0
		}

		err := h(head)

		return err
	})

	s.base = nxs

	if err != nil {
		return true, s, err
	}

	return eos, s, nil
}

type Differ[T constraints.Integer] struct {
	base Stream[T]
	hold []T
	n, i int
}

func DiffN[T constraints.Integer](s Stream[T], n int) Stream[T] {
	return &Differ[T]{base: s, n: n}
}

func Diff[T constraints.Integer](s Stream[T]) Stream[T] {
	return DiffN(s, 1)
}

func (s *Differ[T]) Resolve(h func(v T) error) (bool, Stream[T], error) {
	eos, nxs, err := s.base.Resolve(func(v T) error {
		if len(s.hold) < s.n {
			s.hold = append(s.hold, v)

			return nil
		}

		head := v - s.hold[s.i]
		s.hold[s.i] = v
		s.i++
		if s.i == len(s.hold) {
			s.i = 0
		}

		err := h(head)

		return err
	})

	s.base = nxs

	if err != nil {
		return true, s, err
	}

	return eos, s, nil
}

type Filterer[T any] struct {
	base Stream[T]
	f    func(v T) bool
}

func Filter[T any](s Stream[T], f func(v T) bool) Stream[T] {
	return &Filterer[T]{base: s, f: f}
}

func (s *Filterer[T]) Resolve(h func(v T) error) (bool, Stream[T], error) {
	if s == nil || s.base == nil {
		return true, s, nil
	}

	eos, nxs, err := s.base.Resolve(func(v T) error {
		if !s.f(v) {
			return nil
		}

		err := h(v)

		return err
	})

	s.base = nxs

	if err != nil {
		return true, s, err
	}

	return eos, s, nil
}

type Windower[T any] struct {
	base    Stream[T]
	hold    []T
	i, n, f int
}

func Windowed[T any](s Stream[T], n int, f int) Stream[Stream[T]] {
	return &Windower[T]{base: s, hold: make([]T, 0, f*n), n: n, f: f}
}

func (s *Windower[T]) Resolve(h func(v Stream[T]) error) (bool, Stream[Stream[T]], error) {
	if s == nil || s.base == nil {
		return true, s, nil
	}

	eos, nxs, err := s.base.Resolve(func(v T) error {
		s.hold = append(s.hold, v)
		s.i++

		if s.i < s.n {
			// No complete window yet
			return nil
		}

		headSlice := s.hold[s.i-s.n : s.i]
		headStream := NewFromSlice(headSlice)
		err := h(headStream)

		if s.i == s.f*s.n {
			s.i = s.n - 1
			s.hold = make([]T, s.i, s.f*s.n)
			copy(s.hold, headSlice[1:])
		}

		return err
	})

	s.base = nxs

	if err != nil {
		return true, s, err
	}

	return eos, s, nil
}

type StreamOfFileInts struct {
	filename string
}

func NewStreamOfFileInts(filename string) Stream[int] {
	return &StreamOfFileInts{filename: filename}
}

func (s *StreamOfFileInts) Resolve(h func(v int) error) (bool, Stream[int], error) {
	in, err := os.Open(s.filename)
	if err != nil {
		return false, s, err
	}

	var v int
	_, err = fmt.Fscanf(in, "%d", &v)
	if err != nil {
		return true, s, nil
	}

	err = h(v)
	if err != nil {
		return true, s, err
	}

	return false, &StreamOfFileIntsOpen{in: in}, nil
}

type StreamOfFileIntsOpen struct {
	in *os.File
}

func (s *StreamOfFileIntsOpen) Resolve(h func(v int) error) (bool, Stream[int], error) {
	var v int
	_, err := fmt.Fscanf(s.in, "%d", &v)
	if err != nil {
		return true, s, nil
	}

	err = h(v)
	if err != nil {
		return true, s, err
	}

	return false, s, nil
}

type StreamOfFileLines struct {
	filename string
}

func NewStreamOfFileLines(filename string) Stream[string] {
	return &StreamOfFileLines{filename: filename}
}

func (s *StreamOfFileLines) Resolve(h func(v string) error) (bool, Stream[string], error) {
	file, err := os.Open(s.filename)
	if err != nil {
		return false, s, err
	}

	in := bufio.NewScanner(file)

	eoi := in.Scan()
	if eoi {
		return true, s, in.Err()
	}

	line := in.Text()

	err = h(line)
	if err != nil {
		return true, s, err
	}

	return false, &StreamOfFileLinesOpen{in: in}, nil
}

type StreamOfFileLinesOpen struct {
	in *bufio.Scanner
}

func (s *StreamOfFileLinesOpen) Resolve(h func(v string) error) (bool, Stream[string], error) {
	eoi := s.in.Scan()
	if eoi {
		return true, s, s.in.Err()
	}

	line := s.in.Text()

	err := h(line)
	if err != nil {
		return true, s, err
	}

	return false, s, nil
}

type StreamFromSlice[T any] struct {
	elems []T
	next  int
}

func NewFromSlice[T any](elems []T) Stream[T] {
	return &StreamFromSlice[T]{elems: elems, next: 0}
}

func (s *StreamFromSlice[T]) Resolve(h func(v T) error) (bool, Stream[T], error) {
	if len(s.elems) <= s.next {
		return true, s, nil
	}

	head := s.elems[s.next]
	s.next++

	err := h(head)
	if err != nil {
		return true, s, err
	}

	return false, s, nil
}

func Collect[T any](s Stream[T]) ([]T, error) {
	var collection []T
	for {
		eos, nxs, err := s.Resolve(func(v T) error {
			collection = append(collection, v)

			return nil
		})
		s = nxs
		if eos || err != nil {
			return collection, err
		}
	}
}

func Accumulate[T any](s Stream[T], r T, f func(a, b T) T) (T, error) {
	for {
		eos, nxs, err := s.Resolve(func(v T) error {
			r = f(r, v)
			return nil
		})
		s = nxs
		if eos || err != nil {
			return r, err
		}
	}
}

func Count[T any](s Stream[T]) (int, error) {
	r := 0
	for {
		eos, nxs, err := s.Resolve(func(v T) error {
			r++
			return nil
		})
		s = nxs
		if eos || err != nil {
			return r, err
		}
	}
}
