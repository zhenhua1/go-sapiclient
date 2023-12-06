package main

import (
	"fmt"
	"github.com/zhenhua1/go-sapiclient/sapiclient"
)

func main() {
	c, err := sapiclient.New()
	if err != nil {
		fmt.Println("创建实例失败" + err.Error())
		return
	}
	data := make(map[string]interface{})
	data["mobile"] = "18795487568"
	data["plat_identify"] = "inquiry"
	res, err := c.SetClientCfg("DIP", "CA5E9557164EB7E8CA74761ACDFFFBA2", "http://fa-serve.com").
		SetTimeOut(300).
		SetClientOptions(&sapiclient.ClientOptions{
			RetryCount:    1,
			RetryWaitTime: 1,
		}).SetRequestMethod("get").
		SetService("register").SetMethod("registerUser").DoRequest(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(c.RawResponseHeader)
	fmt.Println(c.RawResponseParams)
	fmt.Println(res)
}
