package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultServerPort = "8081"

	defaultCountGoroutines = "3"

	defaultTimeAdditionMs        = "2000"
	defaultTimeSubtractionMs     = "2000"
	defaultTimeMultiplicationsMs = "2000"
	defaultTimeDivisionsMs       = "2000"
	defaultTimeExponentiationMs  = "2000"
)

/*
COMPUTING_POWER - количество горутин
TIME_ADDITION_MS - время выполнения операции сложения в милисекундах
TIME_SUBTRACTION_MS - время выполнения операции вычитания в милисекундах
TIME_MULTIPLICATIONS_MS - время выполнения операции умножения в милисекундах
TIME_DIVISIONS_MS - время выполнения операции деления в милисекундах
TIME_EXPONENTIATION_MS - время выполнения операции возведения в степень в милисекундах
*/

type Config struct {
	ServerPort            int
	CountGoroutines       int
	TimeAdditionMs        int
	TimeSubtractionMs     int
	TimeMultiplicationsMs int
	TimeDivisionsMs       int
	TimeExponentiationMs  int
}

func LoadFromEnv() (*Config, error) {
	conf := &Config{}
	var err error

	//SERVER_PORT
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = defaultServerPort
	}

	conf.ServerPort, err = strconv.Atoi(serverPort)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("SERVER_PORT"), err)
	}

	//COMPUTING_POWER
	countGoroutines := os.Getenv("COMPUTING_POWER")
	if countGoroutines == "" {
		countGoroutines = defaultCountGoroutines
	}

	conf.CountGoroutines, err = strconv.Atoi(countGoroutines)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("COMPUTING_POWER"), err)
	}

	//TIME_ADDITION_MS
	timeAdditionMs := os.Getenv("TIME_ADDITION_MS")
	if timeAdditionMs == "" {
		timeAdditionMs = defaultTimeAdditionMs
	}

	conf.TimeAdditionMs, err = strconv.Atoi(timeAdditionMs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("TIME_ADDITION_MS"), err)
	}

	//TIME_SUBTRACTION_MS
	timeSubtractionMs := os.Getenv("TIME_SUBTRACTION_MS")
	if timeSubtractionMs == "" {
		timeSubtractionMs = defaultTimeSubtractionMs
	}

	conf.TimeSubtractionMs, err = strconv.Atoi(timeSubtractionMs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("TIME_SUBTRACTION_MS"), err)
	}

	//TIME_MULTIPLICATIONS_MS
	timeMultiplicationsMs := os.Getenv("TIME_MULTIPLICATIONS_MS")
	if timeMultiplicationsMs == "" {
		timeMultiplicationsMs = defaultTimeMultiplicationsMs
	}

	conf.TimeMultiplicationsMs, err = strconv.Atoi(timeMultiplicationsMs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("TIME_MULTIPLICATIONS_MS"), err)
	}

	//TIME_DIVISIONS_MS - время выполнения операции деления в милисекундах
	timeDivisionsMs := os.Getenv("TIME_DIVISIONS_MS")
	if timeDivisionsMs == "" {
		timeDivisionsMs = defaultTimeDivisionsMs
	}

	conf.TimeDivisionsMs, err = strconv.Atoi(timeDivisionsMs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("TIME_DIVISIONS_MS"), err)
	}

	//TIME_EXPONENTIATION_MS
	timeExponentiationMs := os.Getenv("TIME_EXPONENTIATION_MS")
	if timeExponentiationMs == "" {
		timeExponentiationMs = defaultTimeExponentiationMs
	}

	conf.TimeExponentiationMs, err = strconv.Atoi(timeExponentiationMs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as int: %w", os.Getenv("TIME_EXPONENTIATION_MS"), err)
	}
	return conf, nil
}
