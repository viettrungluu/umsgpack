# umsgpack

A tiny (micro), simple implementation of MessagePack.

## Current status

* Basic decoding (unmarshalling) is supported.
* I still have to write more tests.
* Possibly, it should also be able to decode maps to a target struct (type), instead of just to
  maps. (This would require a different interface, and require reflection, so would be done
  separately in any case.)
* Encoding (marshalling) is not yet supported.
