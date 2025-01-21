package util

import (
	"os"
	"testing"
)

type testCase struct {
	args   []string
	output *FlagConfig
}

func TestParseSendFlags(t *testing.T) {
	_, _ = os.Create("test.txt")
	defer os.Remove("test.txt")

	cases := []testCase{
		{
			args:   []string{"lnkr", "send", "-file", "test.txt", "-addr", "192.168.1.1:6969"},
			output: &FlagConfig{Mode: "send", FilePath: "test.txt", Address: "192.168.1.1:6969", Host: "192.168.1.1", Port: "6969"},
		},
		{
			args:   []string{"lnkr", "send", "-file", "test.txt", "-host", "192.168.1.1", "-port", "6969"},
			output: &FlagConfig{Mode: "send", FilePath: "test.txt", Address: "192.168.1.1:6969", Host: "192.168.1.1", Port: "6969"},
		},
	}

	t.Run("send with the possible args", func(t *testing.T) {
		for _, test := range cases {
			out, err := ParseFlags(test.args)
			if err != nil {
				t.Errorf("%s", err.Error())
			}

			assertEqual(t, out.Mode, test.output.Mode, "mode")
			assertEqual(t, out.FilePath, test.output.FilePath, "file path")
			assertEqual(t, out.Address, test.output.Address, "address")
			assertEqual(t, out.Host, test.output.Host, "host")
			assertEqual(t, out.Port, test.output.Port, "port")
		}
	})
}

func TestParseReceiveFlags(t *testing.T) {
	cases := []testCase{
		{
			args:   []string{"lnkr", "receive", "-addr", "192.168.1.1:6969"},
			output: &FlagConfig{Mode: "receive", Address: "192.168.1.1:6969", Host: "192.168.1.1", Port: "6969"},
		},
		{
			args:   []string{"lnkr", "receive", "-host", "192.168.1.1", "-port", "6969"},
			output: &FlagConfig{Mode: "receive", Address: "192.168.1.1:6969", Host: "192.168.1.1", Port: "6969"},
		},
	}

	t.Run("receive with the possible args", func(t *testing.T) {
		for _, test := range cases {
			out, err := ParseFlags(test.args)
			if err != nil {
				t.Errorf("%s", err.Error())
			}

			assertEqual(t, out.Mode, test.output.Mode, "mode")
			assertEqual(t, out.Address, test.output.Address, "address")
			assertEqual(t, out.Host, test.output.Host, "host")
			assertEqual(t, out.Port, test.output.Port, "port")
		}
	})
}

func assertEqual(t *testing.T, got, want any, label string) {
	t.Helper()
	if got != want {
		t.Errorf("%v mismatch: got %v want %v", label, got, want)
	}
}
