# umsgpack

A tiny (micro), simple implementation of [MessagePack](https://msgpack.org/)
([specification](https://github.com/msgpack/msgpack/blob/master/spec.md)).

Unlike other Go implementations of MessagePack, it more closely adheres to MessagePack's weak type
system. This has advantages and disadvantages, as discussed [here](design.md); in short,
transformation from/to weak types (like `map[any]any`) is left to packages like
[mapstructure](https://github.com/go-viper/mapstructure).

## Links

* [Status](status.md)
* [Documentation](https://pkg.go.dev/github.com/viettrungluu/umsgpack)
* [Design](design.md)
