# go-automapper

[![GoDoc](https://godoc.org/github.com/PeteProgrammer/go-automapper?status.svg)](https://godoc.org/github.com/PeteProgrammer/go-automapper)

Go-automapper can automatically map data between different types with
identically named fields in GO. Useful for initializing DTO types from build in
data.

## "V2"

The major change is the ability for client code to provide custom conversions
between types where direct conversion isn't possible.

Install with

```sh
go get github.com/stroiman/go-automapper@next
```

> [!CAUTION]
> Be aware of inconsistencies in argument order!
>
> The original automapper had the exported function `Map(source, dest)`. The new
> `Mapper` type has `Mapper.Map(dest, source)`.
>
> The original function is left unchanged for compatibility reasons, but the new
> signature was chosen to align with other go functions, e.g. the buildin `copy`
> function.

> [!NOTE]
> As such, this version _isn't yet_ breaking backwards compatibility, but at the
> cost of inconsistency. I have not yet decided whether to create true v2
> release, or live with the inconsistencies.

### Motivating example

While the library was originally written for use cases where type conversion
should always be possible, I realised another scenario would fit well;
converting from unvalidated to validated input.

A specific example of this scenario is converting HTML form post data,
consisting entirely of strings, into a "business object".

```go
// FormData is all strings, this is what we receive from the browser
type FormData struct {
    FirstName string
    LastName  string
    Age       string 
}

type User struct {
    FirstName string
    LastName  string
    Age       int
}

func StrToInt(_ context.Context, dst, src any) error {
    a := src.(string)
    b := dst.(*int)
    var err error
    *b, err = strconv.Atoi(a)
    return err
}

var formsMapper automapper.Mapper

func ParseForm(f FormData) (User, error) {
    var user User
    err := formsMapper.Map(&user, f)
    // Or with a context.Context
    // err := formsMapper.MapContext(ctx, &user, f)
    return user, err
}

func init() {
    formsMapper.Register(
        reflect.TypeOf[int](),    // Destination type / domain object
        reflect.TypeOf[string](), // Source type / forms data
        StrToInt)
}
```

### Context support

`MapContext` accepts a `context.Context`, passing it to the relevant mapper
functions. The motivating use case is localization, when parsing data in
human-readable format into machine-readable format depends on the locale of the
current user.

### Useful errors

The error returned from `Map()` decorates the conversion errors with information
about which field name/slice index with invalid data. It will also include _all_
conversion errors, not just the first encountered.

This allows automatically showing relevant validation errors in HTML forms,
e.g., if form field names map directly to struct field names.

### Works for structs too

You can even add custom conversion functions between struct types; or between
primitive types and struct types. This will take precedence over default conversion.

But if you rely heavily on this feature, consider whether this library is the
right tool for the job! Maybe custom mapping code is the better trade-off? More
code but easier to reason about.
