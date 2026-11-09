package service

import (
	"fmt"
	"ginchat/utils"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Upload 接收字段名为 file 的上传，默认存到本机。
// 成功时 Data 是 /asset/upload/文件名，页面把它当成图片或语音的地址。
// 要改走阿里云时，把下面这一行换成 UploadOSS，并在配置里填好密钥。
// @Tags     上传
// @Accept   multipart/form-data
// @Param    file formData file true "图片或语音"
// @Success  200 {object} utils.H
// @Router   /attach/upload [post]
func Upload(c *gin.Context) {
	UploadLocal(c)
}

// UploadLocal 把文件保存到 asset/upload。
// 目录不存在会先创建。文件名是「秒级时间 + 四位随机数 + 后缀」，不用上传者原来的路径，避免写出目录外面。
// 任何一步失败都立刻返回，避免前面已经写过失败响应之后还继续往下写成功响应。
func UploadLocal(c *gin.Context) {
	w := c.Writer
	src, head, err := c.Request.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	defer src.Close()

	if err = os.MkdirAll("./asset/upload", 0o755); err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Intn(10000), fileSuffix(head.Filename))
	dst, err := os.Create(filepath.Join("./asset/upload", fileName))
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	utils.RespOK(w, "/asset/upload/"+fileName, "发送图片成功")
}

// fileSuffix 从原始文件名里取出后缀，例如 .png、.mp3。
// 没有后缀、后缀超过 8 个字符，或里面夹了不是字母数字的符号，都退回 .png。
// 这样上传名里的斜杠和 .. 不会进到最终路径。
func fileSuffix(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) < 2 || len(ext) > 8 {
		return ".png"
	}
	for _, r := range ext[1:] {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return ".png"
		}
	}
	return ext
}

// UploadOSS 把文件传到阿里云对象存储。默认路由不调用它。
// Endpoint、AccessKeyId、Bucket 有一个为空，就返回「未配置阿里云 OSS」，进程继续听端口。
// 成功时 Data 是 https://桶名.Endpoint/文件名，页面可以直接当图片地址。
func UploadOSS(c *gin.Context) {
	w := c.Writer
	endpoint := viper.GetString("oss.Endpoint")
	if endpoint == "" || viper.GetString("oss.AccessKeyId") == "" || viper.GetString("oss.Bucket") == "" {
		utils.RespFail(w, "未配置阿里云 OSS")
		return
	}
	src, head, err := c.Request.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	defer src.Close()

	client, err := oss.New(endpoint, viper.GetString("oss.AccessKeyId"), viper.GetString("oss.AccessKeySecret"))
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	bucket, err := client.Bucket(viper.GetString("oss.Bucket"))
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Intn(10000), fileSuffix(head.Filename))
	if err = bucket.PutObject(fileName, src); err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	url := "https://" + viper.GetString("oss.Bucket") + "." + endpoint + "/" + fileName
	utils.RespOK(w, url, "发送图片成功")
}
