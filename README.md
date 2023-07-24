# teomutex

Golang package teomutex is Teonet Cloud Mutex baset on Google Cloud Storage.
It can be used to serialize computations anywhere on the global internet.

[![GoDoc](https://godoc.org/github.com/teonet-go/teomutex?status.svg)](https://godoc.org/github.com/teonet-go/teomutex/)
[![Go Report Card](https://goreportcard.com/badge/github.com/teonet-go/teomutex)](https://goreportcard.com/report/github.com/teonet-go/teomutex)

## How to install it

The reference implementation in this repo is written in Go. To use teomutex
in a Go program, install the code using this command: `go get -u github.com/marcacohen/teomutex`.

## How to use it

- Create Google Cloud Storage bucket in which lock objects will be stored.
    Use next command to create backet: `gsutil mb gs:mutex`. By default
    the teomutex uses the "mutex" backet name. To use another backet name
    set it in second parameter of the `teomutex.NewMutex` function.

- In your application import the `github.com/teonet/teomutex` package,
    and create new mutex:

```go
    // Creates new Teonet Mutex object.
    m, err := teomutex.NewMutex("test/lock/some_object")
    if err != nil {
        // Process error
        return
    }
    defer m.Close()
```

- Use the `m.Lock` and `m.Unlock` functions to lock and unlock:

```go
    // Lock mutex
    err = m.Lock()
    if err != nil {
        // Process error
        return
    }

    // Do somthing in this protected area

    // Unlock mutex
    err = m.Unlock()
    if err != nil {
        // Process error
        return
    }
```

You can find complete packets documentation at: <https://pkg.go.dev/github.com/teonet-go/teomutex>

## Licence

[BSD](LICENSE)
