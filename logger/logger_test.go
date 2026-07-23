package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestDefaultLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := &DefaultLogger{
		level:  LevelWarn,
		logger: log.New(&buf, "", 0),
	}

	l.Debug("debug msg")
	l.Info("info msg")
	l.Warn("warn msg")
	l.Error("error msg")

	output := buf.String()
	if strings.Contains(output, "DEBUG") {
		t.Error("DEBUG 消息不应输出")
	}
	if strings.Contains(output, "INFO") {
		t.Error("INFO 消息不应输出")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("WARN 消息应输出")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("ERROR 消息应输出")
	}
}

func TestDefaultLogger_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	l := &DefaultLogger{
		level:  LevelDebug,
		logger: log.New(&buf, "", 0),
	}
	l.Debug("test %s", "debug")
	if !strings.Contains(buf.String(), "test debug") {
		t.Error("DEBUG 级别应输出 debug 消息")
	}
}

func TestDefaultLogger_SetLevel(t *testing.T) {
	l := NewDefaultLogger()
	l.SetLevel(LevelError)
	if l.level != LevelError {
		t.Error("SetLevel 未生效")
	}
}

func TestNopLogger(t *testing.T) {
	l := &NopLogger{}
	// 不应 panic
	l.Debug("test")
	l.Info("test")
	l.Warn("test")
	l.Error("test")
	l.SetLevel(LevelDebug)
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %s, want %s", tt.level, got, tt.want)
		}
	}
}

func TestGlobalLogger(t *testing.T) {
	original := Default()
	defer SetDefault(original)

	nop := &NopLogger{}
	SetDefault(nop)
	if Default() != nop {
		t.Error("SetDefault 未生效")
	}
	// 全局便捷方法不应 panic
	Debugf("test")
	Infof("test")
	Warnf("test")
	Errorf("test")
}

// TestGlobalLogger_ConcurrentAccess 验证 SetDefault / Default 并发调用不 panic、不 data race。
// 用 -race 标志跑时若仍有竞争，race detector 会报告。
func TestGlobalLogger_ConcurrentAccess(t *testing.T) {
	original := Default()
	defer SetDefault(original)

	const goroutines = 20
	done := make(chan struct{})

	// 并发写
	for i := 0; i < goroutines; i++ {
		go func() {
			SetDefault(&NopLogger{})
			done <- struct{}{}
		}()
	}
	// 并发读
	for i := 0; i < goroutines; i++ {
		go func() {
			_ = Default()
			done <- struct{}{}
		}()
	}
	for i := 0; i < goroutines*2; i++ {
		<-done
	}
}
