package main

import (
	"DOC/app"
	"log"
)

func main() {
	newApp, err := app.NewApp("./")
	if err != nil {
		log.Fatalf("启动app失败 err : %v", err)
	}
	if err := newApp.Run(); err != nil {
		panic(err)
	}
}
