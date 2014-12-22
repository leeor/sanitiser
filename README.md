# Sanitiser

This package provides contextual scrubbing of data based on tags.

## Usage

```go
type User struct {
  Name     string
  Password string `sanitise:"*"`
  Hash     []byte `sanitise:"api"`
}
```

Calling sanitiser.Sanitise() with an instance of the User struct will return a copy with the Password field scrubbed (set to the zero value) in any context, the Hash field in the api context, and will always leave Name untouched.

When used on objects that are getting serialised together with the `omitempty` directive, the serialised representation will be completely clean of any sensitive data.

## TODO
* Support list of contexts in call to `Sanitise()`
* Improve debug messages
