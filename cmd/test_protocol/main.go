package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "codex", "proto")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	fmt.Println("Starting app-server...")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Start failed: %v\n", err)
		return
	}
	defer cmd.Wait()

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			fmt.Printf("[recv] %s\n", scanner.Text())
		}
		fmt.Println("[recv] EOF")
	}()

	time.Sleep(2 * time.Second)

	userMsg := map[string]any{
		"id": "msg-1",
		"op": map[string]any{
			"type": "user_input",
			"items": []map[string]any{
				{"type": "text", "text": "Say hello in one word"},
			},
		},
	}
	data, _ := json.Marshal(userMsg)
	fmt.Printf("[send] %s\n", data)
	stdin.Write(append(data, '\n'))

	time.Sleep(5 * time.Second)
	fmt.Println("Done waiting for responses")
	stdin.Close()
}
