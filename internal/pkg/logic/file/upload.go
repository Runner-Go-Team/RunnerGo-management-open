package file

import (
	"encoding/base64"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"os"
)

func UploadBase64(ctx *gin.Context, req *rao.FileUploadBase64Req) (string, error) {
	// 解码base64字符串，得到图片的字节数据
	imageBytes, fileErr := base64.StdEncoding.DecodeString(req.FileString)
	if fileErr != nil {
		log.Logger.Error("Upload imageBytes err", proof.WithError(fileErr))
	}

	// 创建文件夹和文件，用于保存图片
	imagesDir := "images"
	basedir := fmt.Sprintf("%s%s", conf.Conf.Static, imagesDir)

	fileErr = os.MkdirAll(basedir, os.ModePerm) // 创建basedir文件夹名称，可以自定义
	if fileErr != nil {
		log.Logger.Error("Upload MkdirAll err", proof.WithError(fileErr))
	}

	// 创建文件夹和文件，用于保存图片
	dir := fmt.Sprintf("%s/%s", basedir, req.PathDir) // 文件夹名称，可以自定义
	fileErr = os.MkdirAll(dir, os.ModePerm)
	if fileErr != nil {
		log.Logger.Error("Upload MkdirAll err", proof.WithError(fileErr))
	}
	fileName := fmt.Sprintf("%s.png", req.FileName) // 文件名称，可以自定义
	filePath := dir + "/" + fileName
	file, fileErr := os.Create(filePath)
	if fileErr != nil {
		log.Logger.Error("Upload Create err", proof.WithError(fileErr))
	}
	defer file.Close()

	// 将图片的字节数据写入文件
	_, fileErr = file.Write(imageBytes)
	if fileErr != nil {
		log.Logger.Error("Upload file.Write err", proof.WithError(fileErr))
	}

	// 返回文件的本地路径
	log.Logger.Info("Upload info path", filePath)

	showDir := fmt.Sprintf("/static/%s/%s/", imagesDir, req.PathDir)

	return fmt.Sprintf("%s%s", showDir, fileName), nil
}
