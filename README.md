# umsgpack

A tiny (micro), simple implementation of [MessagePack](https://msgpack.org/)
([specification](https://github.com/msgpack/msgpack/blob/master/spec.md)).

Unlike other Go implementations of MessagePack, it more closely adheres to MessagePack's weak type
system. This has advantages and disadvantages, as discussed [here](design.md); in short,
transformation between weak types (like `map[any]any`) and strong types (structs) is left to
packages like [mapstructure](https://github.com/go-viper/mapstructure).

## Links

* [Status](status.md)
* [Changelog](CHANGELOG.md)
* [Documentation](https://pkg.go.dev/github.com/viettrungluu/umsgpack#section-documentation)
* [Design](design.md)
