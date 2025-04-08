package main

import (
	"bufio"
	"fmt"
	"github.com/sharipovr/sqlmicro/db"
	"os"
	"strings"
)

func main() {
	database := db.NewDatabase()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("MiniDB v0.1 — type SQL-like commands below:")
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if strings.ToLower(input) == "exit" {
			break
		}

		err := db.HandleCommand(database, input)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
