# go-automapper

[![GoDoc](https://godoc.org/github.com/PeteProgrammer/go-automapper?status.svg)](https://godoc.org/github.com/PeteProgrammer/go-automapper)

Go-automapper can automatically map data between different types with
identically named fields in GO. Useful for initializing DTO types from build in
data.

## V2

The major change is the ability to specify how different types are mapped,
making this useful also for validation.

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

func StrToInt(dst, src any) error {
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
    return user, err
}

func init() {
    formsMapper.Register(
        reflect.TypeOf[int](),    // Destination type / domain object
        reflect.TypeOf[string](), // Source type / forms data
        StrToInt)
}
```

### Useful errors

The error returned from `Map()` decorates the conversion errors with information
about which field name/slice index with invalid data.

One use case is to render validation errors next to the invalid form fields;
though this scenario has yet to be battle tested
