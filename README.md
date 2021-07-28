gstop - stop goroutines/tasks securely, recursively

```go
s1 := gstop.New()
s1.Defer(func() {
    fmt.Println("s1 stopped")
})

s2 := s1.NewChild()
s2.Defer(func() {
    fmt.Println("s2 stopped")
})

s3 := s2.NewChild()
s3.Defer(func() {
    fmt.Println("s3 stopped")
})

s1.Stop()

// s1 stopped
// s3 stopped
// s2 stopped
```