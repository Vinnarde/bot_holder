#!/bin/bash

echo "Сборка приложения Redirector..."

# Установка зависимостей
echo "Установка зависимостей..."
go mod download

# Сборка приложения
echo "Компиляция приложения..."
go build -o redirector main.go

echo "Сборка завершена. Исполняемый файл: redirector" 