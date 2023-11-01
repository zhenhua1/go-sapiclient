package sapiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"path"
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
	appKey        string
	appSecret     string
	sapiServerUrl string //服务器地址
	sapiServerIp  string //指定服务ip
	service       string //指定服务
	method        string //指定服务方法
	clientOptions *ClientOptions
}

// ClientOptions
// @Description: 客户端配置信息
type ClientOptions struct {
	Timeout int               //超时时间
	Headers map[string]string //header参数
	Nonce   string            //随机字符串
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
	viperObject.SetConfigFile(path.Join(workDir, cfg))
	if err = viperObject.ReadInConfig(); err != nil {
		err = errors.New("配置文件读取失败: " + err.Error())
		return
	}
	fmt.Println(viperObject.Get("server.Debug"))
	appKey := viperObject.GetString("sapi.appKey")
	appSecret := viperObject.GetString("sapi.appSecret")
	serverUrl := viperObject.GetString("sapi.serverUrl")
	if serverUrl == "" {
		serverUrl = S_API_URL
	}
	return &sApiClient{
		appKey:        appKey,
		appSecret:     appSecret,
		sapiServerUrl: serverUrl,
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
func (c *sApiClient) DoRequest(body map[string]string) (responseData ResponseData, err error) {
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

	path := "sapi/" + c.service + string(os.PathSeparator) + c.method
	urlReq := c.sapiServerUrl + path
	headers := map[string]string{
		"Accept":      "text/plain;charset=utf-8",
		"Content-Typ": "application/x-www-form-urlencoded",
		"charset":     "utf-8",
	}
	headerOptions := make(map[string]string, 0)
	headerOptions = c.clientOptions.Headers
	headerOptions["client-version"] = VERSION_CLIENT
	if c.clientOptions.Timeout == 0 {
		headerOptions["time"] = time.Now().String()
	}
	if c.clientOptions.Nonce == "" {
		headerOptions["nonce"] = Alnum()
	}
	headerOptions["appkey"] = c.appKey
	headerOptions["sign"] = SEncryptSign(c.appKey, c.appSecret, path, headerOptions["nonce"], headerOptions["time"])
	for key, val := range headerOptions {
		headers[key] = val
	}
	client := resty.New()
	res, err := client.R().SetHeaders(headers).SetFormData(body).Post(urlReq)
	if err != nil {
		return
	}
	if res.IsError() {
		jsonErr, _ := json.Marshal(res.Error())
		err = errors.New(string(jsonErr))
		return
	}
	err = json.Unmarshal(res.Body(), &responseData)
	return
}

// SetClientOptions
//
//	@Description: 设置客户端配置参数
//	@receiver c
//	@Author zzh 2023-10-31 17:25:10
//	@param options
func (c *sApiClient) SetClientOptions(options *ClientOptions) *sApiClient {
	c.clientOptions = options
	return c
}

// SetClientHeaders
//
//	@Description: 设置header头
//	@receiver c
//	@Author zzh 2023-10-31 17:24:47
//	@param headers
func (c *sApiClient) SetClientHeaders(headers map[string]string) *sApiClient {
	c.clientOptions.Headers = headers
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
