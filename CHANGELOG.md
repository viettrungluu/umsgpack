# umsgpack changelog

## 1.1.0 - 2024-07-19

* [#2](https://github.com/viettrungluu/umsgpack/issues/2) Significant performance improvements for
  `UnmarshalBytes`: in `BenchmarkUnmarshalBytes`, there is a 35% reduction in time and 53% reduction
  in allocations.
* [#4](https://github.com/viettrungluu/umsgpack/issues/4) `Unmarshal`/`UnmarshalBytes` now return
  `io.EOF` only if nothing could be read; if a message was partially read, they now return
  `io.ErrUnexpectedEOF`. Previously, these reflected the status of internal reads and the
  distinction was not useful.

## 1.0.0 - 2024-07-17

* Initial release!
