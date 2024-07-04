# umsgpack

A tiny (micro), simple implementation of [MessagePack](https://msgpack.org/)
([specification](https://github.com/msgpack/msgpack/blob/master/spec.md)).

## Unmarshalling design

umsgpack unmarshals to weakly-typed data (e.g., MessagePack maps are unmarshalled as `map[any]any`).
This is especially useful if you frequently have to consume data originating from weakly-typed
languages. If communicating between strongly-typed languages (e.g., Go programs), this may not be
what you want to use (though in that case, perhaps you should be using Protocol Buffers or similar
instead, especially if you want strong typing and a clear schemas).

Conversion to strong (e.g., struct) types is considered to be a separate step. This allows greater
flexibility and keeps the unmarshaller simple. This is not the most efficient, but unmarshalling to
strong types often isn't very efficient anyway, since it frequently involves reflection and high
complexity. (An alternative is to generate code from some schema, but that loses flexibility and
incurs a different kind of complexity.)

In my experience, many unmarshallers have trouble being useful for relatively simple schema such as:
```
{
  type: "SomeType",
  data: <type-dependent data>
}
```
I think it's valuable to be able to simply and separately (and "manually") convert `data` to a
strong type based on `type`. It's hard to design/build an unmarshaller that handles such
variadic-type situations simply, much less efficiently. (E.g., one complication is that in the input
stream the data for the `type` need not precede the data for `data`.)

## Marshalling design (WIP)

An object is marshalled in the following order (the process terminates when the object is
marshalled):
* First, if an object is a supported built-in type, then it marshalled as such.
* Next, the object is marshalled as an application extension type, if applicable.
* Then, the object is marshalled as a standard extension type, if applicable. (Doing application
  extensions before allows applications to override the built-in serialization of `time.Time`.)
* Finally, *transformers* are attempted, in order. On the first transformer that applies, all of the
  above is repeated on the transformed object (but transformers are not re-attempted, to avoid
  infinite loops).
  * If no transformer applies, then marshalling the object is not supported.
  * If a transformer applies, but the transformed object cannot be marshalled (without applying
    transformers), then marshalling the object is not supported.
(The above is applied recursively for container objects when required.)

A transformer is a function that either returns a new, transformed object or indicates that it does
not applies.

TODO: examples (arrays, maps, structs)

It might be more logical to apply transformers first, and not have to repeat the marshalling process
(minus transformers). However, I expect applying transformers to possibly be slow, since it often
involves reflection.

## Current status

* ![umsgpack build and test status](https://github.com/viettrungluu/umsgpack/actions/workflows/go.yml/badge.svg)
* Decoding (unmarshalling) is supported.
  * Possibly, it should also be able to decode maps to a target struct (type), instead of just to
    maps. (This would require a different interface, and require reflection, so would be done
    separately in any case.)
* Basic encoding (marshalling) is supported.
  * Extensions (including timestamps) are not yet supported.
  * Nor are other custom serializations (e.g., to serialize some value as a different value).
  * Ergonomic encoding of arrays (other than `[]any`), maps (other than `map[any]any`), and structs
    is not yet supported.
  * Testing of failures/errors is very incomplete.
