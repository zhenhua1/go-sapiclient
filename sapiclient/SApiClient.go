package sapiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	//RPC服务中心地址 可以使用配置，也可以直接固定线上服务中心
	S_API_URL = "http://sapi.totrial.com/"
	//客户端版本，更新用
	VERSION_CLIENT = "v1.0.0.20230920"
	//配置文件路径地址
	CFG_PATH = "manifest/config/config.toml"
)

type sApiClient struct {
	appKey            string
	appSecret         string
	sapiServerUrl     string      //服务器地址
	sapiServerIp      string      //指定服务ip
	requestMethod     string      //指定请求方法 http的情况下默认是post请求
	service           string      //指定服务
	method            string      //指定服务方法
	RawResponseHeader http.Header //响应头
	RawResponseParams string      //响应参数
	RawStatusCode     int         //响应状态码
	ClientOptions     *ClientOptions
}

// ClientOptions
// @Description: 客户端配置信息
type ClientOptions struct {
	Timeout       int               //超时时间
	Headers       map[string]string //header参数
	Nonce         string            //随机字符串
	RetryCount    int               //重试次数
	RetryWaitTime int               //重试等待时间 秒
}

// ResponseData
// @Description: 响应结果数据
type ResponseData struct {
	Code int
	Msg  string
	Data interface{}
}

// New
//
//	@Description: 创建一个SApiClient client方法
//	@Author zzh 2023-10-31 17:57:20
//	@param ctx
//	@return *sApiClient
func New(cfgPath ...string) (c *sApiClient, err error) {
	// 获取项目目录
	workDir, _ := os.Getwd()
	viperObject := viper.New()
	cfg := CFG_PATH
	if len(cfgPath) > 0 {
		cfg = cfgPath[0]
	}
	appKey, appSecret, serverUrl := "", "", ""
	filePath := path.Join(workDir, cfg)
	_, err = os.Stat(filePath)
	if err == nil && !os.IsNotExist(err) {
		viperObject.SetConfigFile(filePath)
		if err = viperObject.ReadInConfig(); err != nil {
			err = errors.New("配置文件读取失败: " + err.Error())
			return
		}
		if viperObject.IsSet("sapi.appKey") {
			appKey = viperObject.GetString("sapi.appKey")
		}
		if viperObject.IsSet("sapi.appSecret") {
			appSecret = viperObject.GetString("sapi.appSecret")
		}
		if viperObject.IsSet("sapi.serverUrl") {
			serverUrl = viperObject.GetString("sapi.serverUrl")
		}
	}
	if serverUrl == "" {
		serverUrl = S_API_URL
	}
	return &sApiClient{
		appKey:        appKey,
		appSecret:     appSecret,
		sapiServerUrl: serverUrl,
		ClientOptions: &ClientOptions{},
	}, nil
}

// DoRequest
//
//	@Description: 发起请求
//	@receiver c
//	@Author zzh 2023-10-31 17:56:34
//	@param body
//	@return responseData
//	@return err
func (c *sApiClient) DoRequest(body map[string]interface{}) (responseData *ResponseData, err error) {
	if c.appKey == "" || c.appSecret == "" {
		err = errors.New("appKey或者appSecret不能为空")
		return
	}
	if c.service == "" {
		err = errors.New("service不能为空")
		return
	}
	if c.method == "" {
		err = errors.New("method不能为空")
		return
	}
	if c.sapiServerIp != "" {
		urlParse, _ := url.Parse(c.sapiServerUrl)
		c.sapiServerUrl = urlParse.Scheme + "://" + c.sapiServerIp + urlParse.Path
	}

	pathUrl := "sapi/" + c.service + "/" + c.method
	c.sapiServerUrl = strings.TrimRight(c.sapiServerUrl, "/") + "/"
	urlReq := c.sapiServerUrl + pathUrl
	headers := map[string]string{
		"Accept":      "text/plain;charset=utf-8",
		"Content-Typ": "application/x-www-form-urlencoded",
		"charset":     "utf-8",
	}
	headerOptions := make(map[string]string, 0)
	if c.ClientOptions.Headers != nil {
		headerOptions = c.ClientOptions.Headers
	}
	headerOptions["client-version"] = VERSION_CLIENT
	headerOptions["time"] = strconv.Itoa(int(time.Now().Unix()))
	if c.ClientOptions.Nonce == "" {
		headerOptions["nonce"] = Alnum()
	}
	headerOptions["appkey"] = c.appKey
	headerOptions["sign"] = SEncryptSign(c.appKey, c.appSecret, pathUrl, headerOptions["nonce"], headerOptions["time"])
	for key, val := range headerOptions {
		headers[key] = val
	}
	c.ClientOptions.Headers = headers
	client := resty.New()
	if c.ClientOptions.Timeout != 0 {
		client = client.SetTimeout(time.Duration(c.ClientOptions.Timeout) * time.Second)
	}
	if c.ClientOptions.RetryCount != 0 {
		client = client.SetRetryCount(c.ClientOptions.RetryCount)
		if c.ClientOptions.RetryCount != 0 {
			client = client.SetRetryWaitTime(time.Duration(c.ClientOptions.RetryCount))
		} else {
			client = client.SetRetryWaitTime(1 * time.Second)
		}
	}
	clientReq := client.R().SetHeaders(headers)
	res := &resty.Response{}
	//目前只支持get和post请求并且get的参数在url中post的参数在body中
	if c.requestMethod != "" && strings.ToLower(c.requestMethod) == "get" {
		for k, v := range body {
			clientReq.SetQueryParam(k, fmt.Sprintf("%v", v))
		}
		res, err = clientReq.Get(urlReq)
	} else {
		res, err = clientReq.SetBody(body).Post(urlReq)
	}
	if err != nil {
		err = errors.New(err.Error())
		return
	}
	if res.IsError() {
		jsonErr, _ := json.Marshal(res.Error())
		err = errors.New(string(jsonErr))
		return
	}
	c.RawResponseHeader = res.RawResponse.Header
	c.RawResponseParams = res.String()
	c.RawStatusCode = res.StatusCode()
	err = json.Unmarshal(res.Body(), &responseData)
	return
}

// SetClientCfg
//
//	@Description: 设置客户端key、secret、domain
//	@receiver c
//	@Author zzh 2023-11-03 12:10:21
//	@param appKey
//	@param appSecret
//	@param serverUrl
func (c *sApiClient) SetClientCfg(appKey, appSecret, serverUrl string) *sApiClient {
	c.appKey = appKey
	c.appSecret = appSecret
	c.sapiServerUrl = serverUrl
	return c
}

// SetClientOptions
//
//	@Description: 设置客户端配置参数
//	@receiver c
//	@Author zzh 2023-10-31 17:25:10
//	@param options
func (c *sApiClient) SetClientOptions(options *ClientOptions) *sApiClient {
	if options != nil {
		if c.ClientOptions.Timeout != 0 {
			options.Timeout = c.ClientOptions.Timeout
		}
		c.ClientOptions = options
	}
	return c
}

// SetClientHeaders
//
//	@Description: 设置header头
//	@receiver c
//	@Author zzh 2023-10-31 17:24:47
//	@param headers
func (c *sApiClient) SetClientHeaders(headers map[string]string) *sApiClient {
	if headers != nil {
		c.ClientOptions.Headers = headers
	}
	return c
}

// SetSapiServerIp
//
//	@Description: 指定服务ip
//	@receiver c
//	@Author zzh 2023-10-31 16:17:01
//	@param sapiServerIp
func (c *sApiClient) SetSapiServerIp(sapiServerIp string) *sApiClient {
	c.sapiServerIp = sapiServerIp
	return c
}

// SetRequestMethod
//
//	@Description: 指定HTTP请求方法
//	@receiver c
//	@Author zzh 2023-12-06 15:50:21
//	@param requestMethod
//	@return *sApiClient
func (c *sApiClient) SetRequestMethod(requestMethod string) *sApiClient {
	if requestMethod == "" {
		c.requestMethod = "POST"
	} else {
		c.requestMethod = requestMethod
	}
	return c
}

// SetService
//
//	@Description: 指定服务
//	@receiver c
//	@Author zzh 2023-10-31 16:13:33
//	@param service
func (c *sApiClient) SetService(service string) *sApiClient {
	c.service = service
	return c
}

// SetMethod
//
//	@Description: 指定服务方法
//	@receiver c
//	@Author zzh 2023-10-31 16:13:40
//	@param method
func (c *sApiClient) SetMethod(method string) *sApiClient {
	c.method = method
	return c
}

// SetTimeOut
//
//	@Description: 设置超时时间 秒
//	@receiver c
//	@Author zzh 2023-11-03 14:45:45
//	@param timeOut
//	@return *sApiClient
func (c *sApiClient) SetTimeOut(timeOut int) *sApiClient {
	c.ClientOptions.Timeout = timeOut
	return c
}
