package watcher

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type OssClient struct {
	OssURL     string
	httpClient *http.Client
}

func (c *OssClient) UploadFile(fileName string, fileContent []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建一个 form 文件部分
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}

	// 将 fileContent 写入 form 文件部分
	_, err = io.Copy(part, bytes.NewReader(fileContent))
	if err != nil {
		return "", err
	}
	// 必须在 io.Copy 之后关闭 multipart writer 以正确写入 multipart 的结尾
	err = writer.Close()
	if err != nil {
		return "", err
	}
	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", c.OssURL+"/upload", body)
	if err != nil {
		return "", err
	}

	// 设置请求的 Content-Type 为 multipart 的 Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 简化处理，直接返回响应状态
	return resp.Status, nil
}

func (c *OssClient) SayHello() (string, error) {
	// 构建请求 URL
	reqURL := c.OssURL + "/hello"

	// 发起 GET 请求
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应体内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 将响应体转换为字符串并返回
	return string(body), nil
}

func (c *OssClient) FileExists(filename string) (bool, error) {
	resp, err := c.httpClient.Get(c.OssURL + "/exists/" + filename)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result map[string]bool
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, err
	}

	exists, ok := result["exists"]
	if !ok {
		return false, nil // 或者错误处理
	}
	return exists, nil
}

func NewOssClient(OssURL string) *OssClient {
	return &OssClient{
		OssURL:     OssURL,
		httpClient: &http.Client{},
	}
}
