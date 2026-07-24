package rcon

import (
	"fmt"
	"time"

	gorcon "github.com/gorcon/rcon"
)

func ExecuteCommand(host string, port int, password, command string) (string, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := gorcon.Dial(address, password, gorcon.SetDialTimeout(10*time.Second), gorcon.SetDeadline(10*time.Second))
	if err != nil {
		return "", fmt.Errorf("rcon dial %s: %w", address, err)
	}
	defer conn.Close()

	response, err := conn.Execute(command)
	if err != nil {
		return "", fmt.Errorf("rcon execute: %w", err)
	}

	return response, nil
}
