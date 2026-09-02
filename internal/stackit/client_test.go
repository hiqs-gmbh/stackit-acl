package stackit

import (
	"errors"
	"testing"
)

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("some error"), false},
		{"not authenticated", &commandError{args: []string{"project", "list"}, stderr: "Error: you are not authenticated.\n", err: errors.New("exit status 1")}, true},
		{"unauthorized", &commandError{args: []string{"project", "list"}, stderr: "Error: Request failed with status code 401 Unauthorized", err: errors.New("exit status 1")}, true},
		{"status 401", &commandError{args: []string{"project", "list"}, stderr: "Error: Request failed with status code 401", err: errors.New("exit status 1")}, true},
		{"other status", &commandError{args: []string{"project", "list"}, stderr: "Error: Request failed with status code 500", err: errors.New("exit status 1")}, false},
		{"empty stderr", &commandError{args: []string{"project", "list"}, err: errors.New("exit status 1")}, false},
		{"401 only in args", &commandError{args: []string{"redis", "instance", "describe", "abc401def"}, stderr: "Error: Request failed with status code 500", err: errors.New("exit status 1")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandErrorMessage(t *testing.T) {
	err := &commandError{args: []string{"project", "list"}, stderr: "Error: you are not authenticated.", err: errors.New("exit status 1")}
	want := "stackit project list: exit status 1\nError: you are not authenticated."
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
