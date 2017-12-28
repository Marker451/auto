package main

import "net"
import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const MAX_LIMIT = 5000000
const MULTIPLE = 1000000

var socketConn net.Conn

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func NewRPCRequest(method string, para []interface{}) (request *RPCRequest) {
	return &RPCRequest{
		Jsonrpc: "2.0",
		ID:      uuid.NewV4().String(),
		Method:  method,
		Params:  para,
	}
}

type Transcation struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
	Gas   string `json:"gas"`
}
type Response struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  string `json:"result"`
}
type PostContent struct {
	FromAddress string        `json:"from_address"`
	ToAddress   string        `json:"to_address"`
	Pwd         string        `json:"pwd"`
	Monkeys     []*MonkeyConf `json:"monkeys"`
}
type MonkeyConf struct {
	ID    string `json:"id"`
	Limit string `json:"limit"`
	Mode  string `json:"mode"`
}

func fixMonkeyIDs(ids []string) {
	for index, str := range ids {
		//补齐
		length := len(str)
		for k := 0; k < 6-length; k++ {
			str += "0"
		}
		ids[index] = str
	}
}
func fixMonkeyID(id *string) {
	//补齐
	length := len(*id)
	for k := 0; k < 6-length; k++ {
		*id += "0"
	}
}

func FeedMonkeys(w http.ResponseWriter, req *http.Request) {
	fmt.Println("fffffffff")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(body))
	post := &PostContent{}
	err = json.Unmarshal(body, post)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(post)
	feedMonkeys(post.FromAddress, post.ToAddress, post.Pwd, post.Monkeys)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`{"ret":"success"}`))
	return

}

func feedMonkeys(fromAddress string, toAddress string, pwd string, monkeys []*MonkeyConf) {
	if fromAddress == "" || pwd == "" || toAddress == "" || len(monkeys) == 0 {
		return
	}
	for _, monkey := range monkeys {
		id := monkey.ID
		fixMonkeyID(&id)
		idInt, err := strconv.Atoi(id)
		if err != nil {
			panic(err)
		}
		tmpNum, err := strconv.ParseFloat(monkey.Limit, 64)
		if err != nil {
			panic(err)
		}
		max, result := findMaxCombineV2(idInt, int(tmpNum * MULTIPLE), monkey.Mode)
		if len(result) <= 0 {
			return
		}
		fmt.Printf("id为%s 的小猴最佳喂养量为%g 单次喂养%g 喂养次数%d \n", id, float64(max)/MULTIPLE, float64(result[0])/MULTIPLE, len(result))
		for _, num := range result {
			//num = 1000
			hexNum, err := convertToWeiHex(int64(num))
			if err != nil {
				fmt.Println(err)
				return
			}
			sendTranscation(fromAddress, toAddress, string(hexNum), pwd)
			resp, err := readResponse(socketConn)
			if err != nil {
				fmt.Println(err)
			} else {
				if resp.Result == "" {
					fmt.Printf("id:%s 喂养失败\n", id)
				} else {
					fmt.Printf("id:%s 喂养成功 交易hash为 %s \n ", id, resp.Result)
				}
			}
			getTranscationCount(fromAddress)
			readResponse(socketConn)
			time.Sleep(time.Millisecond * 200)

		}

	}

}

var dataDir = flag.String("dataDir", "~/Library/OTCWalletData", "geth dataDir  you cant get this by ps -ef | grep geth ")

func main() {

	flag.Parse()
	tmp := *dataDir
	tmp = tmp + "/geth.ipc"
	dataDir = &tmp

	http.HandleFunc("/feedmonkeys", FeedMonkeys)
	err := http.ListenAndServe(":65399", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func getTranscationCount(address string) (err error) {
	para := []interface{}{}
	para = append(para, address)
	para = append(para, "latest")
	request := NewRPCRequest("eth_getTransactionCount", para)

	raw, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(string(raw))
	_, err = socketConn.Write([]byte(raw))
	if err != nil {
		println(err)
	}
	return
}
func readResponse(c net.Conn) (resp *Response, err error) {
	//maybe something wrong
	buf := make([]byte, 512)
	nr, err := c.Read(buf)
	if err != nil {
		return
	}
	data := buf[0:nr]
	resp = &Response{}
	err = json.Unmarshal(data, resp)
	//fmt.Println(string(data))
	return
}

func sendTranscation(from string, to string, value string, pwd string) (err error) {
	if socketConn == nil {
		socketConn, err = net.Dial("unix", *dataDir)
		if err != nil {
			panic(err)
		}
	}
	transcation := Transcation{
		From:  from,
		To:    to,
		Value: value,
		Gas:   "0x0",
	}
	para := []interface{}{}
	para = append(para, transcation)
	para = append(para, pwd)
	request := NewRPCRequest("personal_sendTransaction", para)
	raw, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(raw))
	_, err = socketConn.Write([]byte(raw))
	if err != nil {
		println(err)
	}
	return

}

//将 num 转换成 以太坊单位 WEI 再转成对应16进制
// 如需转换 1  则需传入 MULTIPLE
// num= MULTIPLE (1WKC) -> MULTIPLE000000000000WEI -> 0xde0b6b3a7640000
func convertToWeiHex(num int64) (b []byte, err error) {
	x := big.NewInt(MULTIPLE) //10^12
	y := big.NewInt(num)
	x.Mul(x, y)
	result := math.HexOrDecimal256(*x)
	b, err = result.MarshalText()
	if err != nil {
		fmt.Println(err)
		return
	}
	return

}

func findMaxCombineV2(num int, limit int, mode string) (max int, result []int) {
	switch mode{
	case "min":
			times := int(limit / num)
			max = times * num
			for i := 0; i < times; i++ {
				result = append(result, num)
			}
			return
	case "max":
		firstNum := int(limit / MULTIPLE)
		lastNums := limit % MULTIPLE
		if lastNums >= num{
			max = firstNum * MULTIPLE + num
		}else{
			max = (firstNum -1)* MULTIPLE + num
		}
		result = append(result,max)
		return
	}
	return
}
