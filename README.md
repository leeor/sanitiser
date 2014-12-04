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

Calling sanitise.Sanitiser() with an instance of the User struct will scrub (set to the zero value) the Password field in any context, the Hash field in the api context, and will always leave Name untouched.

When used on objects that are getting serialised together with the `omitempty` directive, the serialised representation will be completely clean of any sensitive data.

## TODO
* Figure out a mechanism for a less destructive process, i.e., create a
  sanitised copy, or allow restoring sanitised values
* Support list of contexts in call to `Sanitise()`
* Improve debug messages
