package main

import (
    "fmt"
)

func main() {
    var c Contract
    c, _ = c.Init("build/contracts/Test.json", "0x345ca3e014aaf5dca488057592ee47305d9b3e10", "http://localhost:9545")

    n, _ := c.RegisterEventListener("Event")
    c.Listen(n, fn)
}

func fn(data string) error {
    fmt.Println(data)

    return nil
}
