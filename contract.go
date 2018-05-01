package contract

import (
    . "github.com/ethereum/go-ethereum/accounts/abi"
    //"github.com/ethereum/go-ethereum/common"
    "fmt"
    "os"
    "encoding/json"
    "bytes"
    "net/http"
    "io/ioutil"
)

type abiJSON struct {
    Contents []prop `json:"abi"`
}

type prop struct {
    Constant bool `json:"constant"`
    Inputs []par `json:"inputs"`
    Name string `json:"name"`
    Outputs []par `json:"outputs"`
    Type string `json:"type"`
}

type par struct {
    Name string `json:"name"`
    Type string `json:"type"`
}

type res struct {
    Result string `json:"result'`
}

type Contract struct {
    address string
    abi ABI
}

func (c Contract) Init(jsonPath string, address string) (Contract, error) {
    x, _ := os.Open(jsonPath)

    dec := json.NewDecoder(x)
    var a abiJSON
    if err := dec.Decode(&a); err != nil {
        fmt.Printf("NO: %s\n", err)
        return c, err
    }

    str, _ := json.Marshal(a.Contents)

    abi, err := JSON(bytes.NewReader(str))
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return c, err
    }

    c.abi = abi
    c.address = address
    return c, nil
}

func (c Contract) Call(funcName string) (string, error) {
    return c.sendFunc(funcName, "", "eth_call")
}

func (c Contract) Transact(funcName string, from string, args ...interface{}) (string, error) {
    return c.sendFunc(funcName, from, "eth_sendTransaction", args...)
}

func (c Contract) sendFunc(funcName string, from string, rpcType string, args ...interface{}) (string, error) {
    var out []byte
    var err error

    if len(args) > 0 {
        out, err = c.abi.Pack(funcName, args...)
    } else {
        out, err = c.abi.Pack(funcName)
    }

    if err != nil {
        fmt.Printf("Error 2: %s\n", err)
        return "", err
    }

    url := "http://localhost:9545"

    var jsonStr []byte

    if (from == "") {
        jsonStr = []byte(`{"jsonrpc":"2.0","method": "` + rpcType + `", "params": [{"to": "` + c.address + `" , "data": "` + fmt.Sprintf("0x%x", out) + `"}], "id": 1}`)
    } else {
        jsonStr = []byte(`{"jsonrpc":"2.0","method": "` + rpcType + `", "params": [{"from": "` + from + `", "to": "` + c.address + `" , "data": "` + fmt.Sprintf("0x%x", out) + `"}], "id": 1}`)

    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)

    if err != nil {
        return "", err
    }

    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)

    var r res
    json.Unmarshal(body, &r)
    return r.Result, nil
}
