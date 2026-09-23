package oss

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPrepareStreamsToTemporaryFile(t *testing.T) {
	p, err := prepare(strings.NewReader("hello"), "../unsafe/HELLO.TXT", "")
	if err != nil {
		t.Fatal(err)
	}
	name := p.file.Name()
	defer p.close()
	if p.size != 5 || !strings.HasSuffix(p.hashName, ".txt") || p.mimeType != "text/plain" {
		t.Fatalf("预处理结果错误: %+v", p)
	}
	data := make([]byte, 5)
	if _, err := p.file.Read(data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("hello")) {
		t.Fatalf("临时文件内容错误: %q", data)
	}
	p.close()
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		t.Fatalf("临时文件未清理: %v", err)
	}
}
