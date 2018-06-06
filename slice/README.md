# slice
--
    import "github.com/hidez8891/golib/slice"

Package slice provides slice helper function.

## Usage

#### func  All

```go
func All(slice interface{}, pred func(i int) bool) bool
```
All tests whether all values in the slice pass pred().

#### func  Any

```go
func Any(slice interface{}, pred func(i int) bool) bool
```
Any tests whether at least one value in the slice pass pred().

#### func  Filter

```go
func Filter(slice interface{}, pred func(i int) bool) interface{}
```
Filter creates a new slice with all values that pass pred().

#### func  FindIndex

```go
func FindIndex(slice interface{}, value interface{}) int
```
FindIndex returns index of the value in the slice. Otherwise -1 is returned.

#### func  FindIndexIf

```go
func FindIndexIf(slice interface{}, cond func(i int) bool) int
```
FindIndexIf returns index of the value in the slice that satisfy cond().
Otherwise -1 is returned.

#### func  ForEach

```go
func ForEach(slice interface{}, callback func(i int))
```
ForEach executes callback() for each slice's value.

#### func  Includes

```go
func Includes(slice interface{}, value interface{}) bool
```
Includes determines whether the slice includes the value.

#### func  Map

```go
func Map(slice interface{}, conv interface{}) interface{}
```
Map creates a new slice with the results of calling conv() on every element in
the slice.

#### func  Reduce

```go
func Reduce(slice interface{}, accumulater interface{}) interface{}
```
Reduce applies a function against accumulater() and each element in the slice to
reduce it to a single value.
