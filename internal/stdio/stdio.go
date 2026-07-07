package stdio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ReadMessages(reader io.Reader, handle func([]byte) ([]byte, error), writer io.Writer) error {
	buffered := bufio.NewReader(reader)
	for {
		payload, framed, err := readMessage(buffered)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		response, err := handle(payload)
		if err != nil {
			return err
		}
		if framed {
			if _, err := fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n", len(response)); err != nil {
				return err
			}
			if _, err := writer.Write(response); err != nil {
				return err
			}
			continue
		}
		if _, err := writer.Write(response); err != nil {
			return err
		}
	}
}

func readMessage(reader *bufio.Reader) ([]byte, bool, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, false, err
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return readMessage(reader)
	}
	if !bytes.HasPrefix(bytes.ToLower(line), []byte("content-length:")) {
		return line, false, nil
	}

	parts := strings.SplitN(string(line), ":", 2)
	length, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, false, err
	}
	for {
		header, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, false, err
		}
		if len(bytes.TrimSpace(header)) == 0 {
			break
		}
	}
	payload := make([]byte, length)
	_, err = io.ReadFull(reader, payload)
	return payload, true, err
}
