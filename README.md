# sqlmicro
SQL MIcro is my own database server system, written in GO

Пример структуры проекта
bash
Copy
Edit
mini-db/
├── main.go
├── server/          # TCP-сервер
├── parser/          # Парсер SQL/DSL
├── executor/        # Выполнение запросов
├── storage/         # Хранение таблиц/данных
├── model/           # Типы: Table, Row, Column
├── tests/
└── README.md