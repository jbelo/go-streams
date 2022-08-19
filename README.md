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
the head of the stream, and returns the remainder stream (as the second
component of the return tuple).
There isn’t a way to extract the head of the stream, nor the remainder
stream, other than applying the `Resolve` method, at least at the Stream
level of abstraction.

The `Resolve` method also returns a boolean to indicate whether the
stream was in fact the *empty stream*, in which case there was no
head element to apply the parameter function to.
The empty stream will always remain empty, so once `Resolve` returns
true it is said that the *end of stream* has been reached.
That condition will remain so in every subsequent application of
`Resolve`.
Note that at the Stream abstraction level there is no other way to
detect the end of stream condition, except if the Stream is the
value `nil`, which is empty by definition.

The description of error handling will come next.

## Error Handling

## Examples
