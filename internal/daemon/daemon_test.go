package daemon

import (
	"os"
	"testing"
)

func TestRequested(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"api --daemon", []string{"api", "--daemon"}, true},
		{"api -d", []string{"api", "-d"}, true},
		{"api --daemon=true", []string{"api", "--daemon=true"}, true},
		{"api 不带 daemon", []string{"api"}, false},
		{"--daemon 在 api 之前", []string{"--daemon", "api"}, true},
		{"-c 配置与 api", []string{"-c", "conf/app.ini", "api", "-d"}, true},
		{"其他子命令的 --daemon", []string{"export", "--daemon"}, false},
		{"显式关闭", []string{"api", "--daemon=false"}, false},
	}
	for _, c := range cases {
		if got := Requested(c.args); got != c.want {
			t.Errorf("%s: Requested(%v) = %v, want %v", c.name, c.args, got, c.want)
		}
	}
}

func TestIsChild(t *testing.T) {
	os.Unsetenv(envKey)
	if IsChild() {
		t.Fatal("IsChild() = true, want false when env is empty")
	}

	t.Setenv(envKey, "1")
	if !IsChild() {
		t.Fatal("IsChild() = false, want true when env is 1")
	}

	t.Setenv(envKey, "0")
	if IsChild() {
		t.Fatal("IsChild() = true, want false when env is 0")
	}
}
