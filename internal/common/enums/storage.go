package enums

// 文件存储渠道类型
const (
	StorageChannelLocal  = "local"  // 本地存储
	StorageChannelAliOSS = "alioss" // 阿里云OSS
	StorageChannelCOS    = "cos"    // 腾讯云COS
	StorageChannelQiniu  = "qiniu"  // 七牛云Kodo
	StorageChannelMinio  = "minio"  // MinIO/S3兼容
)

// 存储默认渠道标志
const (
	StorageDefaultYes = 1
	StorageDefaultNo  = 0
)
