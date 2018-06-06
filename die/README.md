# die
--
    import "github.com/hidez8891/golib/die"

Package die provides error handling.

## Usage

#### func  If

```go
func If(cond bool, msg interface{})
```
If raise msg if cond is true.

#### func  IfErr

```go
func IfErr(err error)
```
IfErr raise err if err is not nil.

#### func  Revival

```go
func Revival(f func(interface{}))
```
Revival rescue raise message.

#### func  RevivalDetail

```go
func RevivalDetail(f func(string, int, interface{}))
```
RevivalDetail rescue raise message and get file/line infomation.

#### func  RevivalErr

```go
func RevivalErr(f func(error))
```
RevivalErr rescue raise error.
