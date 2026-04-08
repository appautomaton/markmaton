package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/appautomaton/markmaton/internal/engine"
	"github.com/appautomaton/markmaton/internal/model"
)

func main() {
	request, err := readRequest(os.Stdin)
	if err != nil {
		fail(err)
	}

	response, err := engine.Process(request)
	if err != nil {
		fail(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		fail(err)
	}
}

func readRequest(r io.Reader) (model.Request, error) {
	var request model.Request

	payload, err := io.ReadAll(r)
	if err != nil {
		return request, fmt.Errorf("read request: %w", err)
	}
	if len(payload) == 0 {
		return request, errors.New("request body is empty")
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		return request, fmt.Errorf("decode request: %w", err)
	}

	request.ApplyDefaults()
	if err := request.Validate(); err != nil {
		return request, err
	}

	return request, nil
}

func fail(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
