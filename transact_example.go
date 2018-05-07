package main

import (
    "fmt"
    "github.com/dkaps125/go-contract"
)

func main() {
    var c contract.Contract
    c, _ = c.Init("build/contracts/Test.json", "0x345ca3e014aaf5dca488057592ee47305d9b3e10", "http://localhost:9545")

    s, _ := c.Call("getNum")
    fmt.Printf("%s\n", s)
    s, _ = c.Transact("setNum", "0x627306090abab3a6e1400e9345bc60c78a8bef57", uint8(123))
    fmt.Printf("%s\n", s)
    s, _ = c.Call("getNum")
    fmt.Printf("%s\n", s)

}

