package upload

import (
	"bytes"
	"os"
	"testing"

	"server_api/config"
	"server_api/internal/common/app"
)

func testApp(t *testing.T) *app.App {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("需要 MySQL、Redis 和本地存储配置；设置 RUN_INTEGRATION_TESTS=1 后运行")
	}
	cfg, err := config.Load("../../config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	application, err := app.Init(cfg)
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	return application
}

// 服务层二进制流上传：模拟获取到的文件流（如微信小程序二维码）直接转存
func TestSaveFileStream(t *testing.T) {
	application := testApp(t)

	stream := bytes.NewReader([]byte("hello file stream"))
	record, err := SaveFileStream(application, "api", "hello.txt", "text/plain", stream, nil)
	if err != nil {
		t.Fatalf("SaveFileStream 失败: %v", err)
	}
	if record.HashName == "" || record.RelativePath == "" || record.AbsolutePath == "" {
		t.Fatalf("记录字段不完整: %+v", record)
	}
	if record.Client != "api" || record.UploaderID != nil {
		t.Fatalf("来源/上传人字段错误: %+v", record)
	}
	t.Logf("落库成功: hash=%s rel=%s abs=%s", record.HashName, record.RelativePath, record.AbsolutePath)
}
