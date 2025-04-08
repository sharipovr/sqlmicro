package main

import (
	"bufio"
	"fmt"
	"github.com/sharipovr/sqlmicro/db"
	"os"
	"strings"
)

func main() {
	database, err := db.LoadFromFile()
	if err != nil {
		fmt.Println("Error loading database:", err)
		return
	}

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

	// Сохраняем базу перед выходом
	err = db.SaveToFile(database)
	if err != nil {
		fmt.Println("Error saving database:", err)
	} else {
		fmt.Println("Database saved to disk.")
	}
}
