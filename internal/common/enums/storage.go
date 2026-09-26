package enums

import "slices"

// 文件存储渠道类型
const (
	StorageChannelLocal  = "local"  // 本地存储
	StorageChannelAliOSS = "alioss" // 阿里云OSS
	StorageChannelCOS    = "cos"    // 腾讯云COS
	StorageChannelQiniu  = "qiniu"  // 七牛云Kodo
	StorageChannelMinio  = "minio"  // MinIO/S3兼容
)

// ValidStorageChannels 文件存储渠道的合法取值，新增渠道时在此登记。
var ValidStorageChannels = []string{
	StorageChannelLocal,
	StorageChannelAliOSS,
	StorageChannelCOS,
	StorageChannelQiniu,
	StorageChannelMinio,
}

// IsValidStorageChannel 判断存储渠道是否为已定义的合法枚举值。
func IsValidStorageChannel(channel string) bool {
	return slices.Contains(ValidStorageChannels, channel)
}

// 存储默认渠道标志
const (
	StorageDefaultYes = 1
	StorageDefaultNo  = 0
)
