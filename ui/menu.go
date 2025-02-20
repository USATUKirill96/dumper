package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dumper/config"
)

// clearScreen очищает терминал
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// SelectEnvironment позволяет пользователю выбрать окружение
func SelectEnvironment(cfg *config.Config) (*config.Environment, error) {
	clearScreen()
	fmt.Println("\nДоступные окружения:")
	for i, env := range cfg.Environments {
		fmt.Printf("%d. %s\n", i+1, env.Name)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nВыберите номер окружения: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения ввода: %w", err)
		}

		var num int
		if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &num); err != nil {
			clearScreen()
			fmt.Println("Пожалуйста, введите число")
			continue
		}

		if num < 1 || num > len(cfg.Environments) {
			clearScreen()
			fmt.Println("Неверный номер окружения")
			continue
		}

		env := &cfg.Environments[num-1]
		clearScreen()
		fmt.Printf("Выбрано окружение: %s\n", env.Name)
		return env, nil
	}
}

// ShowMenu отображает главное меню и возвращает выбранную команду
func ShowMenu(env *config.Environment, dbConn *config.DbConnection) string {
	fmt.Printf("\nТекущее окружение: %s\n", env.Name)

	// Показываем информацию о локальной базе
	fmt.Println("\nПараметры подключения к локальной базе:")
	fmt.Printf("DSN: %s\n", dbConn.GetDSN())
	fmt.Printf("Хост:     %s\n", dbConn.Host)
	fmt.Printf("Порт:     %s\n", dbConn.Port)
	fmt.Printf("Пользователь: %s\n", dbConn.User)
	fmt.Printf("Пароль:   %s\n", dbConn.Password)
	fmt.Printf("База:     %s\n", dbConn.Database)
	fmt.Printf("SSL режим: %s\n", dbConn.SslMode)

	fmt.Println("\nДоступные команды:")
	fmt.Println("1. Сделать дамп")
	fmt.Println("2. Загрузить дамп в локальную базу")
	fmt.Println("3. Сменить окружение")
	fmt.Println("4. Выход")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nВыберите команду: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Ошибка чтения ввода: %v\n", err)
			continue
		}
		clearScreen()
		return strings.TrimSpace(input)
	}
}
