package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/cronitor-local/internal/logger"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	l := logger.New(nil, logger.LevelInfo)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestInfo_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	l.Info("myjob", "job started")

	out := buf.String()
	if !strings.Contains(out, "[INFO]") {
		t.Errorf("expected [INFO] in output, got: %s", out)
	}
	if !strings.Contains(out, "job=myjob") {
		t.Errorf("expected job=myjob in output, got: %s", out)
	}
	if !strings.Contains(out, "job started") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestDebug_BelowLevel_Suppressed(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	l.Debug("", "verbose detail")

	if buf.Len() != 0 {
		t.Errorf("expected no output for DEBUG below INFO level, got: %s", buf.String())
	}
}

func TestError_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelDebug)
	l.Error("backupjob", "exit code 1")

	out := buf.String()
	if !strings.Contains(out, "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got: %s", out)
	}
}

func TestInfof_NoJob(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	l.Infof("loaded %d jobs", 3)

	out := buf.String()
	if !strings.Contains(out, "loaded 3 jobs") {
		t.Errorf("expected formatted message, got: %s", out)
	}
	if strings.Contains(out, "job=") {
		t.Errorf("expected no job= field when job is empty, got: %s", out)
	}
}

func TestWarn_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelWarn)
	l.Warn("syncjob", "slow execution")

	out := buf.String()
	if !strings.Contains(out, "[WARN]") {
		t.Errorf("expected [WARN] in output, got: %s", out)
	}
}
