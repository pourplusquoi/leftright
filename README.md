# leftright

A concurrent map that is optimized for scenarios where reads are more frequent than writes.

## Example

```golang
import "github.com/pourplusquoi/leftright"

reader, writer := leftright.New()
reader2 := reader.Clone()

go func() {
    writer.Insert("a", 1)
    writer.Insert("b", 2)
    writer.Insert("c", 3)
    writer.Publish()

    writer.Remove("a")
    writer.Publish()
}()

go func() {
    a, exists := reader.Get("a")
    fmt.Println(a, exists)

    b, exists := reader.Get("b")
    fmt.Println(b, exists)

    c, exists := reader.Get("c")
    fmt.Println(c, exists)
}()

go func() {
    a, exists := reader2.Get("a")
    fmt.Println(a, exists)

    b, exists := reader2.Get("b")
    fmt.Println(b, exists)

    c, exists := reader2.Get("c")
    fmt.Println(c, exists)
}()
```

## Explanation

Basically, there are two copies of the same map in the concurrency model, where the writer writes to the one map and readers read from the other map. Only after the writer explicitly publishes changes can readers see the modifications until then.

The benefit of "left-right" concurrency model is that reads are entirely lock-free and can be very fast, and that the writer can decide when to publish changes to readers. However, the cost is that the model consumes extra memory, and that the model only supports single writer, and that the writer has to do extra work while writing.

To summarize, this concurrency model is suitable for use cases where reads are more frequent than writes and strong consistency is not required.
