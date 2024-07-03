# umsgpack

A tiny (micro), simple implementation of [MessagePack](https://msgpack.org/)
([specification](https://github.com/msgpack/msgpack/blob/master/spec.md)).

## Unmarshalling philosophy

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

## Current status

* ![umsgpack build and test status](https://github.com/viettrungluu/umsgpack/actions/workflows/go.yml/badge.svg)
* Decoding (unmarshalling) is supported.
* Possibly, it should also be able to decode maps to a target struct (type), instead of just to
  maps. (This would require a different interface, and require reflection, so would be done
  separately in any case.)
* Encoding (marshalling) is not yet supported.
