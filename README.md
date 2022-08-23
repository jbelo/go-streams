# go-streams

A simplistic encoding of streams in Go, for fun.

## Streams

A *Stream of T* is any value of a type that implements the Stream interface,

```
type Stream[T any] interface {
    Resolve(func (v T) error) (bool, Stream[T], error)
}
```

The `Resolve` method decomposes the stream into its constituent parts,
the *head* and the *remainder*. The head will be a value of type `T`
and the remainder will be itself a Stream of `T`.
Successive applications of `Resolve` on the remainder streams resolve
into a succession of values of type `T`, which intuitively form
“a stream of T’s.”

The `Resolve` method takes, as parameter, a function to be applied to
the head of the stream, and usually returns the remainder stream (as
the second component of the return tuple).
There isn’t a way to extract the head of the stream, nor the remainder
stream, other than applying `Resolve`, at least at the Stream level of
abstraction.

### End of Stream

The `Resolve` method also returns a boolean to indicate whether the
stream was in fact the *empty stream*, in which case there was no
head element to apply the parameter function to.
The empty stream will always remain empty, so once `Resolve` returns
true it is said that the *end of stream* has been reached.
That condition will remain so in every subsequent application of
`Resolve`.
Note that at the Stream level of abstraction there is no other way
to detect the end of stream condition, except if the Stream is the
value `nil`, which is empty by definition.

### Stream State

An evaluation of `Resolve` may not result in the evaluation of its
parameter function.
This is the case for the empty stream, as explained above, but that
is not the only case.
It is allowed to have a particular evaluation of `Resolve` to not
resolve the head of the stream.
Thus, there isn’t in general a one to one correspondence between
evaluations of `Resolve` and evaluations of its parameter function,
even for non-empty streams.
Nevertheless, for every resolution of the head, there will be an
evaluation of the parameter function applied to it.
In other words, no head of stream will be skipped.

Evaluations of `Resolve` may, as a consequence, not return the remainder
stream.
Instead, the same stream to which `Resolve` was applied to may be returned
(in general with its fields updated), or a new stream altogether may be
returned.
The returned stream should nevertheless capture the state reached in the
evaluation.

### Error Handling

If the stream resolution fails, then `Resolve` may return an error (on the
third component of the return tuple).
An error will usually lead to the end of stream.
The return tuple will hence have the end of stream boolean set to `true` in
case of an error.

### Examples

The following is an example definition of a type of stream that result from
dropping `n` heads of stream starting at a given base stream.
This is an example of a "stateful stream", which also illustrates the non
correspondence between evaluations of `Resolve` and evaluations of its
parameter function.

```
// A Dropper represents the stream that results from "dropping" the
// first n elements from a given stream.
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
```

Note that, for a `Dropper`, the first `n` evaluations to `Resolve` will
definitely not evaluate its parameter function `h`.
It could even take more than `n` evaluations, since the evaluation of
`s.base.Resolve(...)` may itself not evaluate its parameter function.
Note also that the stream keeps state in the `c` field, and the same
`s` stream is returned in every case.
