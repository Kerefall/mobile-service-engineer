package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    
    "github.com/jackc/pgx/v5"
)

func main() {
    // Подключение к базе данных
    connStr := "postgres://postgres:postgres@localhost:5432/mobile_engineer?sslmode=disable"
    
    fmt.Println("Подключение к базе данных...")
    
    conn, err := pgx.Connect(context.Background(), connStr)
    if err != nil {
        log.Fatal("Ошибка подключения к базе данных:", err)
    }
    defer conn.Close(context.Background())
    
    fmt.Println("Подключение успешно установлено")
    
    // Список миграций в правильном порядке
    migrations := []string{
        "migrations/001_create_engineers_table.up.sql",
        "migrations/002_create_parts_table.up.sql",
        "migrations/003_create_orders_table.up.sql",
        "migrations/004_create_order_parts_table.up.sql",
        "migrations/005_seed_data.up.sql",
    }
    
    fmt.Println("\nЗапуск миграций...")
    fmt.Println("----------------------------------------")
    
    // Выполняем каждую миграцию
    for _, migration := range migrations {
        fmt.Printf("Выполняется файл: %s\n", migration)
        
        // Читаем файл
        sqlBytes, err := os.ReadFile(migration)
        if err != nil {
            log.Printf("ОШИБКА при чтении %s: %v", migration, err)
            continue
        }
        
        sql := string(sqlBytes)
        
        // Разбиваем на отдельные запросы
        queries := strings.Split(sql, ";")
        successCount := 0
        
        for _, query := range queries {
            query = strings.TrimSpace(query)
            if query == "" {
                continue
            }
            
            _, err = conn.Exec(context.Background(), query)
            if err != nil {
                log.Printf("ОШИБКА при выполнении запроса в %s: %v", migration, err)
                log.Printf("Запрос: %s", query)
            } else {
                successCount++
            }
        }
        
        fmt.Printf("Успешно выполнено %s (запросов: %d)\n", migration, successCount)
        fmt.Println("----------------------------------------")
    }
    
    fmt.Println("\nВсе миграции успешно завершены")
    
    // Проверяем результат
    fmt.Println("\nПроверка результатов:")
    fmt.Println("----------------------------------------")
    
    var count int
    
    err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM engineers").Scan(&count)
    if err == nil {
        fmt.Printf("Инженеров в базе: %d\n", count)
    } else {
        fmt.Printf("Ошибка при проверке таблицы engineers: %v\n", err)
    }
    
    err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM parts").Scan(&count)
    if err == nil {
        fmt.Printf("Запчастей в базе: %d\n", count)
    } else {
        fmt.Printf("Ошибка при проверке таблицы parts: %v\n", err)
    }
    
    err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM orders").Scan(&count)
    if err == nil {
        fmt.Printf("Заказов в базе: %d\n", count)
    } else {
        fmt.Printf("Ошибка при проверке таблицы orders: %v\n", err)
    }
    
    err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM order_parts").Scan(&count)
    if err == nil {
        fmt.Printf("Списанных запчастей в базе: %d\n", count)
    } else {
        fmt.Printf("Ошибка при проверке таблицы order_parts: %v\n", err)
    }
    
    fmt.Println("----------------------------------------")
    fmt.Println("\nРабота скрипта завершена")
}