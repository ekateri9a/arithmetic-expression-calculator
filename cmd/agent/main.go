package main

import (
	"arithmetic-expression-calculator/internal/config"
	"arithmetic-expression-calculator/internal/entities"
	"arithmetic-expression-calculator/internal/logger"
	"arithmetic-expression-calculator/internal/utils"
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TaskResult struct {
	Id     string  `json:"id"`
	Result float64 `json:"result"`
}

func Calculate(cl chan bool, tasks chan entities.Task, results chan TaskResult, conf *config.Config) {
	for {
		select {
		case <-cl:
			logger.Info("close calculate goroutines")
			return
		case task := <-tasks:
			logger.Info("new task:", task)
			result := TaskResult{}
			result.Id = task.Id
			var t time.Duration
			switch task.Operation {
			case "*":
				result.Result = task.Arg1 * task.Arg2
				t = time.Duration(conf.TimeMultiplicationsMs) * time.Millisecond
			case "/":
				result.Result = task.Arg1 / task.Arg2
				t = time.Duration(conf.TimeDivisionsMs) * time.Millisecond
			case "+":
				result.Result = task.Arg1 + task.Arg2
				t = time.Duration(conf.TimeAdditionMs) * time.Millisecond
			case "-":
				result.Result = task.Arg1 - task.Arg2
				t = time.Duration(conf.TimeSubtractionMs) * time.Millisecond
			case "^":
				result.Result = math.Pow(task.Arg1, task.Arg2)
				t = time.Duration(conf.TimeExponentiationMs) * time.Millisecond
			}
			time.Sleep(t)
			results <- result
		}
	}
}

func Run() {
	logger.Info("start main agent")
	logger.Info("reading config...")
	conf, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to read config")
		os.Exit(1)
	}

	countGoroutines := conf.CountGoroutines

	cl := make(chan bool)
	tasks := make(chan entities.Task)
	results := make(chan TaskResult)

	for i := 0; i < countGoroutines; i++ {
		go Calculate(cl, tasks, results, conf)
	}

	client := &http.Client{}
	serverPort := conf.ServerPort

	go func() {
		for {
			select {
			// GET /internal/task
			case <-time.After(1 * time.Second):
				resp, err := client.Get("http://localhost:" + strconv.Itoa(serverPort) + "/internal/task")
				if err != nil {
					logger.Error(err, resp.StatusCode)
				}

				if resp.StatusCode == 200 {
					task := new(entities.Task)
					utils.DecodeRespondBody(resp.Body, task)
					defer resp.Body.Close()
					if err != nil {
						logger.Error("failed to decode body", err)
					} else {
						go func() {
							tasks <- *task
						}()
					}
				} else {
					logger.Info("StatusCode: ", resp.StatusCode)
				}
			// POST /internal/task
			case r := <-results:
				logger.Info("new result: ", r)
				body, err := json.Marshal(r)
				if err != nil {
					logger.Error("failed to marshall payload:", r)
				} else {
					resp, err := client.Post("http://localhost:"+strconv.Itoa(serverPort)+"/internal/task", "application/json", bytes.NewBuffer(body))
					if err != nil {
						logger.Error(err)
					}
					resp.Body.Close()
					logger.Info("StatusCode: ", resp.StatusCode)
				}
			}
		}
	}()

	time.Sleep(time.Duration(20) * time.Minute)
}

func main() {
	Run()
}
