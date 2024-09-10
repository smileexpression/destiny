package controller

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/constant"
	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
	"smile.expression/destiny/pkg/http/api"
	"smile.expression/destiny/pkg/storage"
)

type StorageController struct {
	r             *gin.Engine
	db            *gorm.DB
	storageClient *storage.Client
}

func NewStorageController(r *gin.Engine, db *gorm.DB, storageClient *storage.Client) *StorageController {
	return &StorageController{
		r:             r,
		db:            db,
		storageClient: storageClient,
	}
}

func (s *StorageController) Register() {
	rg := s.r.Group("/api/v1/storage")

	rg.PUT("/upload", s.upload)
	rg.DELETE("/remove", s.remove)
}

func (s *StorageController) upload(c *gin.Context) {
	var (
		ctx0 = c.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	file, err := c.FormFile("file")
	if err != nil {
		log.WithError(err).Error("error getting file from form")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	content, err := file.Open()
	if err != nil {
		log.WithError(err).Error("error opening file")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	defer func(content multipart.File) {
		if err = content.Close(); err != nil {
			log.WithError(err).Error("error closing file")
		}
	}(content)

	resp, err := s.storageClient.PutObject(ctx0, "picture", file.Filename, content, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get(constant.ContentType),
	})
	if err != nil {
		log.WithError(err).Error("error uploading file")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"result": resp})
}

func HandleImage(c *gin.Context) {
	db := database.GetDB()

	// 从上下文中获取查询参数"id"
	id := c.Query("id")

	// 声明一个model.Image变量image，用于存储从数据库中检索到的图像
	var image model.Image

	// 使用db.First()从数据库中检索具有指定ID的图像记录
	err := db.First(&image, id).Error

	// 如果检索过程中出现错误，使用c.AbortWithStatusJSON()返回一个JSON响应，指示无法找到图像
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// 如果成功检索到图像记录，使用c.Data()将图像数据以HTTP响应的形式返回给客户端。此处假设所有图像均为JPEG格式的二进制数据，因此将MIME类型设置为"image/jpeg"
	c.Data(http.StatusOK, "image/jpeg", image.Blob)
}

func (s *StorageController) remove(c *gin.Context) {
	var (
		ctx0 = c.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	var req api.RemoveObjectRequest
	if err := c.BindJSON(&req); err != nil {
		log.WithError(err).Error("error parsing request")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	bucketName, objectName, err := parseURL(req.URL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if err = s.storageClient.RemoveObject(ctx0, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func parseURL(fileURL string) (string, string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", "", err
	}

	// 解析路径，例如 "/my-bucket/my-file.txt"
	path := strings.TrimPrefix(parsedURL.Path, "/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid URL format")
	}

	return parts[0], parts[1], nil
}
