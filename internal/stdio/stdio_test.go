package stdio

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadMessagesHandlesContentLengthFraming(t *testing.T) {
	var output bytes.Buffer
	input := strings.NewReader("Content-Length: 17\r\n\r\n{\"jsonrpc\":\"2.0\"}")

	err := ReadMessages(input, func(payload []byte) ([]byte, error) {
		if string(payload) != `{"jsonrpc":"2.0"}` {
			t.Fatalf("payload = %q", payload)
		}
		return []byte(`{"ok":true}`), nil
	}, &output)
	if err != nil {
		t.Fatalf("ReadMessages returned error: %v", err)
	}

	want := "Content-Length: 11\r\n\r\n{\"ok\":true}"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}
