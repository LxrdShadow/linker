package util

import "testing"

func TestParseFlags(t *testing.T) {
	t.Run("error when there's no subcommands provided", func(t *testing.T) {
		args := []string{"lnkr"}
		want := "expected 'send' or 'receive' subcommands\n"

		_, err := ParseFlags(args)
		assertEqual(t, err.Error(), want, "error value")
	})

	t.Run("error when there's no addr or host+port on receive", func(t *testing.T) {
		args := []string{"lnkr", "receive"}
		want := "'receive' have to come with an address (host:port)\n"

		_, err := ParseFlags(args)
		assertEqual(t, err.Error(), want, "error value")
	})

	t.Run("error when we provide both addr and host or port on receive", func(t *testing.T) {
		args := []string{"lnkr", "receive", "-addr", "172.17.0.1:6969", "-host", "127.0.0.1"}
		want := "'receive' have to only come with 'addr' (host:port) or 'host' and 'port' \n"

		_, err := ParseFlags(args)
		assertEqual(t, err.Error(), want, "error value")
	})
}

func assertEqual(t *testing.T, got, want any, label string) {
	t.Helper()
	if got != want {
		t.Errorf("%v mismatch: got %v want %v", label, got, want)
	}
}
