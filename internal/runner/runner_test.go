package runner

import (
	"strings"
	"testing"
	"time"
)

func TestNew_DefaultTimeout(t *testing.T) {
	r := New(0)
	if r.timeout != 30*time.Minute {
		t.Errorf("expected default timeout 30m, got %v", r.timeout)
	}
}

func TestNew_CustomTimeout(t *testing.T) {
	r := New(5 * time.Second)
	if r.timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", r.timeout)
	}
}

func TestRun_SuccessfulCommand(t *testing.T) {
	r := New(10 * time.Second)
	res := r.Run("echo-job", "echo hello")

	if res.Err != nil {
		t.Fatalf("expected no error, got %v", res.Err)
	}
	if !strings.Contains(res.Output, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", res.Output)
	}
	if res.JobName != "echo-job" {
		t.Errorf("expected JobName 'echo-job', got %q", res.JobName)
	}
	if res.Duration <= 0 {
		t.Error("expected positive duration")
	}
	if res.Started.IsZero() {
		t.Error("expected non-zero start time")
	}
}

func TestRun_FailingCommand(t *testing.T) {
	r := New(10 * time.Second)
	res := r.Run("fail-job", "exit 1")

	if res.Err == nil {
		t.Fatal("expected error for failing command, got nil")
	}
}

func TestRun_Timeout(t *testing.T) {
	r := New(100 * time.Millisecond)
	res := r.Run("slow-job", "sleep 10")

	if res.Err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestRun_CombinedOutput(t *testing.T) {
	r := New(10 * time.Second)
	res := r.Run("stderr-job", "echo out && echo err >&2")

	if !strings.Contains(res.Output, "out") {
		t.Errorf("expected stdout in output, got %q", res.Output)
	}
	if !strings.Contains(res.Output, "err") {
		t.Errorf("expected stderr in output, got %q", res.Output)
	}
}
