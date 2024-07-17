# umsgpack design

## Unmarshalling

umsgpack unmarshals to weakly-typed data (e.g., MessagePack maps are unmarshalled as `map[any]any`).
This is especially useful if you frequently have to consume data originating from weakly-typed
languages. If communicating between strongly-typed languages (e.g., Go programs), this may not be
what you want to use (though in that case, perhaps you should be using Protocol Buffers or similar
instead, especially if you want strong typing and a clear schemas).

Conversion to strong (e.g., struct) types is considered to be a separate step, left to packages like
[mapstructure](https://github.com/go-viper/mapstructure). This allows greater flexibility and keeps
the unmarshaller simple. This is not the most efficient, but unmarshalling to strong types often
isn't very efficient anyway, since it frequently involves reflection and high complexity. (An
alternative is to generate code from some schema, but that loses flexibility and incurs a different
kind of complexity.)

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

That said, simple transformation of unmarshalled objects is supported:
* This is mainly to support unmarshalling extension types.
* For example, unresolved extension type -1 (the timestamp extension) is transformed to `time.Time`.
* This could also be used to emit an error if an unsupported extension type is encountered.
* It could also be used to validate the `string`s are valid UTF-8.
The difference is that the transformation is not aware of context: it only has the unmarshalled
object to work with (i.e., its type and data), whereas context is very much needed to convert to
structs.

## Marshalling

An object is marshalled in the following order (the process terminates when the object is
marshalled):
* First, the application transformer (if any) is applied. This may transform the object to some new
  (presumably marshallable) object.
  * Note that transformers can easily be composed.
* Next, unless disabled, the standard transformer is applied.
  * Currently, this just supports the standard timestamp extension type by transforming `time.Time`.
* Then the (possibly-transformed) object is marshalled, if it is of a supported type.
  * Supported types include basic types, along with arrays, slices, and maps (of supported types).
  * The unresolved extension type (consisting of the extension type and data) is also supported.
* If the (possibly-transformed) object is not of a supported type, then marshalling fails.
(The above is applied recursively for container objects when required.)

For example, a transformer can support extension types by transforming them to the unresolved
extension type (which can then be marshalled); this is how the standard transformer supports the
timestamp extension.

Another example is transforming structs to maps, which can then be marshalled. For maximum
flexibility and capability, one may want to use a package like
[mapstructure](https://github.com/go-viper/mapstructure) for such a transformer.
