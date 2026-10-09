package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dsn := "postgres://task_user:task_password@localhost:15432/task_manager?sslmode=disable"
	pgx, err := pgxpool.New(ctx, dsn)
	if err != nil {
		//лог
		fmt.Printf("error: %v", err)
		return
	}

	defer pgx.Close()

	err = pgx.Ping(ctx)
	if err != nil {
		//лог
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Println("Подключение успешно")
	//лог
}
