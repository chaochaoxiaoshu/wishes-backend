package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
	"wishes/config"

	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// StorageService 封装了对腾讯云对象存储的操作
type StorageService struct {
	client  *cos.Client
	config  *config.Config
	baseURL string
	bucket  string
}

// NewStorageService 创建腾讯云存储服务实例
func NewStorageService(config *config.Config) *StorageService {
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.COSBucketName, config.COSRegion))
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.COSSecretID,
			SecretKey: config.COSSecretKey,
		},
	})

	return &StorageService{
		client:  client,
		config:  config,
		baseURL: config.COSBaseURL,
		bucket:  config.COSBucketName,
	}
}

// UploadImage 上传图片到腾讯云COS
// 参数:
// - file: 要上传的文件
// - directory: 存储目录（例如：'images/avatar'）
// 返回:
// - fileURL: 上传成功后的文件访问URL
// - error: 错误信息
func (s *StorageService) UploadImage(file *multipart.FileHeader, directory string) (string, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %w", err)
	}
	defer src.Close()

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", uniqueID, ext)

	// 组合存储路径
	objectKey := filepath.Join(directory, fileName)

	// 设置上传选项
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: file.Header.Get("Content-Type"),
		},
	}

	// 执行上传
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = s.client.Object.Put(ctx, objectKey, src, opt)
	if err != nil {
		return "", fmt.Errorf("上传文件到COS失败: %w", err)
	}

	// 如果设置了自定义域名（通过baseURL），则使用自定义域名构建URL
	var fileURL string
	if s.baseURL != "" {
		fileURL = fmt.Sprintf("%s/%s", s.baseURL, objectKey)
	} else {
		fileURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s",
			s.config.COSBucketName,
			s.config.COSRegion,
			objectKey)
	}

	return fileURL, nil
}

// DeleteImage 从腾讯云COS删除图片
func (s *StorageService) DeleteImage(objectKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.client.Object.Delete(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}
