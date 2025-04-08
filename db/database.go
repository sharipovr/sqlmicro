package db

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type DataType string

const (
	INT  DataType = "INT"
	TEXT DataType = "TEXT"
	BOOL DataType = "BOOL"
)
const saveFile = "db.gob"

type Column struct {
	Name string
	Type DataType
}

type Table struct {
	Name    string
	Columns []Column
	Rows    [][]any
}

type Database struct {
	Tables map[string]*Table
}

func NewDatabase() *Database {
	return &Database{
		Tables: make(map[string]*Table),
	}
}

// Обработка команды
func HandleCommand(db *Database, input string) error {
	inputUpper := strings.ToUpper(input)
	switch {
	case strings.HasPrefix(inputUpper, "CREATE TABLE"):
		return handleCreateTable(db, input)
	case strings.HasPrefix(inputUpper, "INSERT INTO"):
		return handleInsertInto(db, input)
	case strings.HasPrefix(inputUpper, "SELECT"):
		return handleSelect(db, input)
	case strings.HasPrefix(inputUpper, "DELETE FROM"):
		return handleDeleteFrom(db, input)
	default:
		return errors.New("Unsupported command")
	}
}

// Пример: CREATE TABLE users (id INT, name TEXT, active BOOL);
func handleCreateTable(db *Database, input string) error {
	re := regexp.MustCompile(`(?i)^CREATE TABLE (\w+)\s*\((.+)\);?$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) != 3 {
		return errors.New("Invalid CREATE TABLE syntax")
	}

	tableName := matches[1]
	columnDefs := strings.Split(matches[2], ",")
	var columns []Column

	for _, def := range columnDefs {
		parts := strings.Fields(strings.TrimSpace(def))
		if len(parts) != 2 {
			return fmt.Errorf("Invalid column definition: %s", def)
		}
		colName := parts[0]
		colType := DataType(strings.ToUpper(parts[1]))

		if colType != INT && colType != TEXT && colType != BOOL {
			return fmt.Errorf("Unsupported type: %s", colType)
		}

		columns = append(columns, Column{Name: colName, Type: colType})
	}

	if _, exists := db.Tables[tableName]; exists {
		return errors.New("Table already exists")
	}

	db.Tables[tableName] = &Table{
		Name:    tableName,
		Columns: columns,
		Rows:    make([][]any, 0),
	}

	fmt.Printf("Table '%s' created successfully.\n", tableName)
	return nil
}

func handleInsertInto(db *Database, input string) error {
	re := regexp.MustCompile(`(?i)^INSERT INTO (\w+)\s+VALUES\s*\((.+)\);?$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) != 3 {
		return errors.New("Invalid INSERT INTO syntax")
	}

	tableName := matches[1]
	rawValues := matches[2]

	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("Table '%s' does not exist", tableName)
	}

	// Разбить по запятым, учитывая строки в кавычках
	values, err := parseCSVValues(rawValues)
	if err != nil {
		return err
	}

	if len(values) != len(table.Columns) {
		return fmt.Errorf("Expected %d values, got %d", len(table.Columns), len(values))
	}

	row := make([]any, len(values))
	for i, valStr := range values {
		colType := table.Columns[i].Type
		valStr = strings.TrimSpace(valStr)

		switch colType {
		case INT:
			num, err := strconv.Atoi(valStr)
			if err != nil {
				return fmt.Errorf("Invalid INT: %s", valStr)
			}
			row[i] = num
		case TEXT:
			valStr = strings.Trim(valStr, "'\"") // Убираем кавычки
			row[i] = valStr
		case BOOL:
			valStrLower := strings.ToLower(valStr)
			if valStrLower == "true" {
				row[i] = true
			} else if valStrLower == "false" {
				row[i] = false
			} else {
				return fmt.Errorf("Invalid BOOL: %s", valStr)
			}
		}
	}

	table.Rows = append(table.Rows, row)
	fmt.Println("Row inserted successfully.")
	return nil
}

// Простая разбивка значений с поддержкой кавычек
func parseCSVValues(input string) ([]string, error) {
	var values []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(input); i++ {
		c := input[i]
		switch c {
		case '\'', '"':
			if inQuotes {
				if c == quoteChar {
					inQuotes = false
				} else {
					current.WriteByte(c)
				}
			} else {
				inQuotes = true
				quoteChar = c
			}
		case ',':
			if inQuotes {
				current.WriteByte(c)
			} else {
				values = append(values, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		values = append(values, current.String())
	}

	return values, nil
}

// Пример: SELECT * FROM users;
func handleSelect(db *Database, input string) error {
	re := regexp.MustCompile(`(?i)^SELECT \* FROM (\w+);?$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) != 2 {
		return errors.New("Invalid SELECT syntax")
	}

	tableName := matches[1]
	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("Table '%s' does not exist", tableName)
	}

	// Заголовки
	for _, col := range table.Columns {
		fmt.Printf("%s\t", col.Name)
	}
	fmt.Println()

	// Данные
	for _, row := range table.Rows {
		for _, val := range row {
			fmt.Printf("%v\t", val)
		}
		fmt.Println()
	}

	if len(table.Rows) == 0 {
		fmt.Println("(no rows)")
	}

	return nil
}

// Пример: DELETE FROM users WHERE id = 1;
func handleDeleteFrom(db *Database, input string) error {
	re := regexp.MustCompile(`(?i)^DELETE FROM (\w+)\s+WHERE\s+(\w+)\s*=\s*(.+);?$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) != 4 {
		return errors.New("Invalid DELETE syntax")
	}

	tableName := matches[1]
	columnName := matches[2]
	valueStr := strings.TrimSpace(matches[3])

	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("Table '%s' does not exist", tableName)
	}

	// Найти индекс колонки
	colIndex := -1
	colType := DataType("")
	for i, col := range table.Columns {
		if col.Name == columnName {
			colIndex = i
			colType = col.Type
			break
		}
	}

	if colIndex == -1 {
		return fmt.Errorf("Column '%s' does not exist", columnName)
	}

	// Преобразовать значение
	var parsedValue interface{}
	switch colType {
	case INT:
		v, err := strconv.Atoi(valueStr)
		if err != nil {
			return fmt.Errorf("Invalid INT value: %s", valueStr)
		}
		parsedValue = v
	case TEXT:
		parsedValue = strings.Trim(valueStr, "'\"")
	case BOOL:
		lv := strings.ToLower(valueStr)
		if lv == "true" {
			parsedValue = true
		} else if lv == "false" {
			parsedValue = false
		} else {
			return fmt.Errorf("Invalid BOOL value: %s", valueStr)
		}
	}

	// Удаляем строки
	var newRows [][]interface{}
	deleted := 0

	for _, row := range table.Rows {
		if row[colIndex] == parsedValue {
			deleted++
			continue
		}
		newRows = append(newRows, row)
	}

	table.Rows = newRows
	fmt.Printf("Deleted %d row(s).\n", deleted)
	return nil
}

func SaveToFile(db *Database) error {
	file, err := os.Create(saveFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(db)
}

func LoadFromFile() (*Database, error) {
	file, err := os.Open(saveFile)
	if err != nil {
		// Если файла нет — вернём пустую БД
		return NewDatabase(), nil
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var db Database
	err = decoder.Decode(&db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}
