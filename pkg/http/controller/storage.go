package controller

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/constant"
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

func (c *StorageController) Register() {
	rg := c.r.Group("/storage")

	rg.PUT("/upload", c.upload)
	rg.DELETE("/remove", c.remove)

	rg2 := c.r.Group("/image")
	rg2.POST("/upload", c.multiUpload)
	rg2.GET("/get", c.get)
	rg2.DELETE("/delete", c.delete)
}

func (c *StorageController) get(ctx *gin.Context) {
	// 从上下文中获取查询参数"id"
	uri := ctx.Query("id")

	ctx.JSON(http.StatusOK, gin.H{"id": c.storageClient.SetEndpoint(uri)})
}

func (c *StorageController) multiUpload(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	username := ctx.GetString("user")
	// TODO test
	username = "smile"
	// 使用c.MultipartForm()从上下文中检索多部分表单数据
	form, err := ctx.MultipartForm()
	if err != nil {
		log.WithError(err).Error("failed to parse multipart form")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 初始化一个空的图像ID切片，以跟踪成功上传的图像的ID
	var URLs []string

	// 循环遍历多部分表单数据中的所有文件头
	for _, fileHeaders := range form.File {
		for _, fileHeader := range fileHeaders {
			// 对于每个文件头，使用fileHeader.Open()打开文件
			var content multipart.File
			content, err = fileHeader.Open()
			if err != nil {
				log.WithError(err).Error("failed to open multipart file")
				continue
			}

			var resp *api.PutObjectResponse
			objectName := fmt.Sprintf("%s/%s/%s", username, time.Now().Format("2006-01-02"), fileHeader.Filename)
			resp, err = c.storageClient.PutObject(ctx0, "picture", objectName, content, fileHeader.Size, minio.PutObjectOptions{
				ContentType: fileHeader.Header.Get(constant.ContentType),
			})

			if err = content.Close(); err != nil {
				log.WithError(err).Error("failed to close multipart file")
				return
			}

			// 将成功上传的图像ID添加到imageIds切片中
			URLs = append(URLs, c.storageClient.GetObjectName(resp.URL))
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"imageIds": URLs})
}

func (c *StorageController) upload(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	file, err := ctx.FormFile("file")
	if err != nil {
		log.WithError(err).Error("error getting file from form")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	content, err := file.Open()
	if err != nil {
		log.WithError(err).Error("error opening file")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	defer func(content multipart.File) {
		if err = content.Close(); err != nil {
			log.WithError(err).Error("error closing file")
		}
	}(content)

	resp, err := c.storageClient.PutObject(ctx0, "picture", file.Filename, content, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get(constant.ContentType),
	})
	if err != nil {
		log.WithError(err).Error("error uploading file")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{"result": resp})
}

func (c *StorageController) delete(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	username := ctx.GetString("user")
	// TODO
	username = "smile"

	bucketName := "picture"
	objectName := ctx.Query("id")
	if objectName == "" || !strings.HasPrefix(objectName, username+"/") {
		log.Errorf("invalid object name %s", objectName)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid object name " + objectName})
		return
	}

	if err := c.storageClient.RemoveObject(ctx0, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		log.WithError(err).Error("error deleting file")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func (c *StorageController) remove(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	var req api.RemoveObjectRequest
	if err := ctx.BindJSON(&req); err != nil {
		log.WithError(err).Error("error parsing request")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	bucketName, objectName, err := c.storageClient.ParseURL(req.URL)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if err = c.storageClient.RemoveObject(ctx0, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
}
