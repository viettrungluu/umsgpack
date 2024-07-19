# umsgpack changelog

## 1.1.0 - 2024-07-19

* Significant performance improvements for `UnmarshalBytes`: in `BenchmarkUnmarshalBytes`, there is
  a 35% reduction in time and 53% reduction in allocations. See
  [e2265f9](https://github.com/viettrungluu/umsgpack/commit/e2265f9465553aee070d5812c7cd4a7963958a59)
  (and [c75db84](https://github.com/viettrungluu/umsgpack/commit/c75db8479666cd0c849696b586fedd3cd02acaff)).
* `Unmarshal`/`UnmarshalBytes` now return `io.EOF` only if nothing could be read; if a message was
  partially read, they now return `io.ErrUnexpectedEOF`. Previously, these reflected the status of
  internal reads and the distinction was not useful. See
  [#4](https://github.com/viettrungluu/umsgpack/issues/4).

## 1.0.0 - 2024-07-17

* Initial release!
