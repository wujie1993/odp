package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/e"
)

type RESTClient struct {
	endpoint string
}

func (c RESTClient) Get() *Request {
	return &Request{
		endpoint: c.endpoint,
		method:   http.MethodGet,
	}
}

func (c RESTClient) Put() *Request {
	return &Request{
		endpoint: c.endpoint,
		method:   http.MethodPut,
	}
}

func (c RESTClient) Post() *Request {
	return &Request{
		endpoint: c.endpoint,
		method:   http.MethodPost,
	}
}

func (c RESTClient) Delete() *Request {
	return &Request{
		endpoint: c.endpoint,
		method:   http.MethodDelete,
	}
}

func NewRESTClient(endpoint string) RESTClient {
	return RESTClient{
		endpoint: endpoint,
	}
}

type Request struct {
	endpoint     string
	resource     string
	resourceName string
	namespace    string
	method       string
	params       map[string]string
	apiVersion   string
	body         []byte
}

// Version 设置请求资源的结构版本
func (r *Request) Version(apiVersion string) *Request {
	r.apiVersion = apiVersion
	return r
}

// Namespace 设置请求资源的命名空间
func (r *Request) Namespace(namespace string) *Request {
	r.namespace = namespace
	return r
}

// Resource 设置请求资源的类型
func (r *Request) Resource(resource string) *Request {
	r.resource = resource
	return r
}

// Namespace 设置请求资源的名称
func (r *Request) Name(name string) *Request {
	r.resourceName = name
	return r
}

// Data 设置请求资源的类型
func (r *Request) Data(data interface{}) *Request {
	body, _ := json.Marshal(data)
	r.body = body
	return r
}

// Params 设置请求查询参数
func (r *Request) Params(params map[string]string) *Request {
	r.params = params
	return r
}

// Do 执行请求
func (r *Request) Do(ctx context.Context) *Result {
	urlStr := r.endpoint + "/api"

	if r.apiVersion == "" {
		return &Result{
			err: e.Errorf("please specific api version"),
		}
	}
	urlStr += "/" + r.apiVersion

	if r.namespace != "" {
		urlStr += "/namespaces/" + r.namespace
	}

	if r.resource == "" {
		return &Result{
			err: e.Errorf("please specific resource"),
		}
	}
	urlStr += "/" + r.resource

	if r.resourceName != "" {
		urlStr += "/" + r.resourceName
	}

	requestUrl, _ := url.Parse(urlStr)

	query := requestUrl.Query()
	for key, value := range r.params {
		query.Set(key, value)
	}

	requestUrl.RawQuery = query.Encode()
	log.Debugf("[%s] %s %s", r.method, requestUrl.String(), string(r.body))

	req, err := http.NewRequestWithContext(ctx, r.method, requestUrl.String(), bytes.NewReader(r.body))
	if err != nil {
		return &Result{
			err: err,
		}
	}

	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return &Result{
			err: err,
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Result{
			err: err,
		}
	}
	log.Debugf("%s %s", resp.Status, string(data))

	if resp.StatusCode != http.StatusOK {
		return &Result{
			data: data,
			err:  e.Errorf(resp.Status),
		}
	}

	return &Result{
		data: data,
	}
}

type Result struct {
	err  error
	data []byte
	body ResultBody
}

type ResultBody struct {
	OpCode int
	OpDesc string
	Data   interface{}
}

// Into 将http请求返回结果写入receiver中，receiver必须是指针
func (r *Result) Into(receiver interface{}) error {
	if r.err == nil {
		r.body = ResultBody{
			Data: receiver,
		}
	}
	if err := json.Unmarshal(r.data, &r.body); err != nil {
		return err
	}

	if r.err != nil {
		if r.body.OpDesc != "" {
			return e.Errorf(r.body.OpDesc)
		} else {
			return r.err
		}
	}
	return nil
}
