package contract

import (
    . "github.com/ethereum/go-ethereum/accounts/abi"
    //"github.com/ethereum/go-ethereum/common"
    "fmt"
    "os"
    "encoding/json"
    "encoding/hex"
    "bytes"
    "net/http"
    "io/ioutil"
    "time"
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
    Result string `json:"result"`
}

type eventRes struct {
    Result []data `json:"result"`
}

type data struct {
    Data string `json:"data"`
}

type Contract struct {
    address string
    abi ABI
    json abiJSON
    url string
}

func (c Contract) Init(jsonPath string, address string, url string) (Contract, error) {
    x, _ := os.Open(jsonPath)

    dec := json.NewDecoder(x)
    var a abiJSON
    if err := dec.Decode(&a); err != nil {
        fmt.Printf("NO: %s\n", err)
        return c, err
    }

    c.json = a

    str, _ := json.Marshal(a.Contents)

    abi, err := JSON(bytes.NewReader(str))
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return c, err
    }

    c.abi = abi
    c.address = address
    c.url = url
    return c, nil
}

func (c Contract) Call(funcName string, args ...interface{}) ([]interface{}, error) {
    str, _ := c.sendFunc(funcName, "", "eth_call", args...)
    encb, _ := hex.DecodeString(str[2:])
    res, _ := c.abi.Methods[funcName].Outputs.UnpackValues(encb)

    return res, nil
}

func (c Contract) Transact(funcName string, from string, args ...interface{}) (string, error) {
    return c.sendFunc(funcName, from, "eth_sendTransaction", args...)
}

func (c Contract) RegisterEventListener(eventName string) (string, error) {
    var event prop
    var acc string

    for _, v := range c.json.Contents {
        if v.Name == eventName && v.Type == "event" {
            event = v
            break
        }
    }

    acc += event.Name + "("

    for _, v := range event.Inputs {
        acc += v.Type + ","
    }

    acc = acc[:len(acc) - 1] + ")"

    hashJSON := []byte(`{"jsonrpc":"2.0","method":"web3_sha3","params":["` + acc + `"],"id":1}`)
    hashTemp, _ := c.httpPost(hashJSON)
    hash := hashTemp.Result

    filterJSON := []byte(`{"jsonrpc":"2.0","method":"eth_newFilter","params":[{"topics":["` + hash + `"]}],"id":1}`)
    filterTemp, _ := c.httpPost(filterJSON)
    filterNum := filterTemp.Result

    return filterNum, nil
}

func (c Contract) Listen(eventNum string, eventName string, cb func([]interface{}) error) {
    checkJSON := []byte(`{"jsonrpc":"2.0","method":"eth_getFilterChanges","params":["` + eventNum + `"],"id":1}`)

    var r eventRes
    for {
        time.Sleep(time.Second * 2)

        resp := sendHttp(checkJSON, c.url)
        json.Unmarshal(resp, &r)

        for _, v := range r.Result {
            encb, _ := hex.DecodeString(v.Data[2:])
            res, _ := c.abi.Events[eventName].Inputs.UnpackValues(encb)
            cb(res)
        }
    }
}

func (c Contract) ListenOnce(eventNum string, eventName string, cb func([]interface{}) error) {
    checkJSON := []byte(`{"jsonrpc":"2.0","method":"eth_getFilterChanges","params":["` + eventNum + `"],"id":1}`)

    var r eventRes
    for {
        time.Sleep(time.Second * 2)

        resp := sendHttp(checkJSON, c.url)
        json.Unmarshal(resp, &r)

        for _, v := range r.Result {
            encb, _ := hex.DecodeString(v.Data[2:])
            res, _ := c.abi.Events[eventName].Inputs.UnpackValues(encb)
            cb(res)
            return
        }
    }
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

    var jsonStr []byte

    if (from == "") {
        jsonStr = []byte(`{"jsonrpc":"2.0","method": "` + rpcType + `", "params": [{"to": "` + c.address + `" , "data": "` + fmt.Sprintf("0x%x", out) + `", "gas": "0x2dc6c0"}], "id": 1}`)
    } else {
        jsonStr = []byte(`{"jsonrpc":"2.0","method": "` + rpcType + `", "params": [{"from": "` + from + `", "to": "` + c.address + `" , "data": "` + fmt.Sprintf("0x%x", out) + `", "gas": "0x2dc6c0"}], "id": 1}`)

    }

    resp, err := c.httpPost(jsonStr)

    if err != nil {
        return "", nil
    }

    return resp.Result, nil
}

func (c Contract) httpPost(jsonStr []byte) (res, error) {
    var r res

    body := sendHttp(jsonStr, c.url)

    json.Unmarshal(body, &r)
    return r, nil
}

func sendHttp(jsonStr []byte, url string) []byte {
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)

    if err != nil {
        return nil
    }

    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    return body
}
