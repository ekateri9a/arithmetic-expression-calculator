package main

import (
	"arithmetic-expression-calculator/internal/config"
	op "arithmetic-expression-calculator/internal/entities"
	"arithmetic-expression-calculator/internal/logger"
	"arithmetic-expression-calculator/internal/utils"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Stack struct {
	items []string
}

func (s *Stack) Push(data string) {
	s.items = append(s.items, data)
}

func (s *Stack) Pop() {
	if s.IsEmpty() {
		return
	}
	s.items = s.items[:len(s.items)-1]
}

func (s *Stack) Top() (string, error) {
	if s.IsEmpty() {
		return "", fmt.Errorf("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

func (s *Stack) IsEmpty() bool {
	if len(s.items) == 0 {
		return true
	}
	return false
}

func checkBalance(exp string) bool {
	s := Stack{}

	for _, char := range exp {
		switch char {
		case '(':
			s.Push(string(char))
		case ')':
			top, _ := s.Top()
			if top == "(" {
				s.Pop()
			} else {
				return false
			}
		}
	}

	if s.IsEmpty() {
		return true
	}
	return false
}

// todo
func checkBalance2(exp string) bool {
	s := make([]string, 0)

	for _, char := range exp {
		switch c := string(char); c {
		case "(":
			s = append(s, c)
		case ")":
			if len(s) == 0 {
				return false
			}

			top := s[len(s)-1]
			if top == "(" {
				s = s[:len(s)-1]
			} else {
				return false
			}
		}
	}

	if len(s) == 0 {
		return true
	}
	return false
}

func isOperator(char string) bool {
	switch char {
	case "+":
		return true
	case "-":
		return true
	case "*":
		return true
	case "/":
		return true
	default:
		return false
	}
}

func CheckExpression(exp string) ([]op.Chunk, error) {
	result := make([]op.Chunk, 0)
	pointFlag := true
	num := ""

	exp = strings.TrimSpace(exp)
	if len(exp) == 0 {
		return nil, errors.New("empty expression")
	}

	logger.Info("expression", exp, ", CheckExpression (): ", checkBalance(exp))
	if !checkBalance(exp) {
		logger.Error("expression", exp, "invalid")
		return nil, errors.New("invalid expression")
	}

	r := []rune(exp)

	for i := 0; i < len(r); i++ {
		if unicode.IsDigit(r[i]) || (r[i] == '.') {
			if r[i] == '.' && pointFlag {
				if num == "" {
					return nil, errors.New("invalid point")
				}
				num += string(r[i])
				pointFlag = false
				continue
			}
			if r[i] == '.' && !pointFlag {
				return nil, errors.New("double point")
			}
			num += string(r[i])
			continue
		} else {
			if i > 0 && r[i-1] == '.' {
				return nil, errors.New("invalid point in the end of number")
			}

			if i > 0 && unicode.IsDigit(r[i-1]) {
				result = append(result, op.Chunk{
					OpFlag: op.Num,
					Val:    num,
				})
				num = ""
				pointFlag = true
			}
		}

		if unicode.IsSpace(r[i]) {
			continue
		}

		if !(r[i] == '(' || r[i] == ')' || r[i] == '/' || r[i] == '*' || r[i] == '+' || r[i] == '-' || r[i] == '^') {
			return nil, errors.New("invalid expression")
		}

		if r[i] == ')' && i > 0 {
			if v := result[len(result)-1]; v.OpFlag == op.Operation && v.Val != ")" {
				return nil, errors.New("invalid expression: operation before )")
			}
		}

		if r[i] == '(' && (r[i+1] == '/' || r[i+1] == '*' || r[i+1] == '+' || r[i+1] == '-' || r[i+1] == '^') {
			if v := result[len(result)-1]; v.OpFlag == op.Operation {
				return nil, errors.New("invalid expression: operation before )")
			}
		}

		result = append(result, op.Chunk{
			OpFlag: op.Operation,
			Val:    string(r[i]),
		})

	}

	if num != "" {
		result = append(result, op.Chunk{
			OpFlag: op.Num,
			Val:    num,
		})
	}

	for i := 0; i < len(result)-1; i++ {
		if result[i].OpFlag == result[i+1].OpFlag && !(result[i+1].Val == "(" || result[i].Val == ")") {
			return nil, errors.New("invalid expression repeated")
		}
	}

	//
	if len(result) < 3 {
		return nil, errors.New("invalid expression - short")
	}

	return result, nil
}

type AddExpressionResponse struct {
	Id int `json:"id"`
}

func (repo *Repo) AddExpressionHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWith400(w, "http method must be POST")
		return
	}

	expression := new(op.Expression)
	err := utils.DecodeBody(w, r, expression)
	if err != nil {
		logger.Error("failed to decode body", err)
		if err = utils.RespondWith400(w, "failed to decode body"); err != nil {
			logger.Error(err)
		}
		return
	}
	defer r.Body.Close()

	fmt.Println(expression.Val)

	// CheckExpression ---------------------------------------------------------
	chunks, e := CheckExpression(expression.Val)
	if e != nil {
		fmt.Println(e)
		utils.RespondWith422(w)
		return
	}

	// InfixToPostfix ---------------------------------------------------------
	chunksPostfix := utils.InfixToPostfix(chunks)
	logger.Info("Postfix:", chunksPostfix)

	//5
	id := len(repo.RepoE) + 1
	repo.RepoE = append(repo.RepoE, Expression{
		Id:               id,
		Status:           "Calculate",
		Result:           0,
		ExpressionChunks: chunksPostfix,
	})

	//
	// find tasks
	countNewTasks := 0
	for i := 2; i < len(chunksPostfix); i++ {
		if chunksPostfix[i].OpFlag == op.Operation && chunksPostfix[i-1].OpFlag == op.Num && chunksPostfix[i-2].OpFlag == op.Num {
			countNewTasks++
			arg1, _ := strconv.ParseFloat(chunksPostfix[i-2].Val, 64)
			arg2, _ := strconv.ParseFloat(chunksPostfix[i-1].Val, 64)
			repo.Tasks = append(repo.Tasks, op.Task{
				Id:            strconv.Itoa(id) + "-" + strconv.Itoa(i-2),
				Arg1:          arg1,
				Arg2:          arg2,
				Operation:     chunksPostfix[i].Val,
				OperationTime: 0,
			})
		}
	}
	logger.Info("Added", countNewTasks, "new tasks")

	// Response
	payload := AddExpressionResponse{Id: id}

	respondErr := utils.SuccessRespondWith201(w, payload)
	if respondErr != nil {
		logger.Error(respondErr)
	}

}

func (repo *Repo) GetExpressionsHandleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	if strings.HasPrefix(r.URL.String(), "/expressions/:") {
		s, _ := strings.CutPrefix(r.URL.String(), "/expressions/:")

		id, err := strconv.Atoi(s)
		if err != nil {
			utils.RespondWith404(w)
			return
		}

		expression, err := repo.GetExpression(id)
		if err != nil {
			utils.RespondWith404(w)
			return
		}

		payload := map[string]Expression{"expression": expression}

		respondErr := utils.SuccessRespondWith200(w, payload)
		if respondErr != nil {
			logger.Error(respondErr)
		}
		return
	}

	payload := repo

	respondErr := utils.SuccessRespondWith200(w, payload)
	if respondErr != nil {
		logger.Error(respondErr)
	}

}

func (repo *Repo) TaskHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		logger.Info("Task POST")
		//Коды ответа:
		//* 200 - успешно записан результа
		//* 404 - нет такой задачи
		//* 422 - невалидные данные
		//* 500 - что-то пошло не так

		calcResult := new(op.CalcResult)
		err := utils.DecodeBody(w, r, calcResult)
		if err != nil {
			logger.Error("failed to decode body", err)
			if err = utils.RespondWith400(w, "failed to decode body"); err != nil {
				logger.Error(err)
			}
			return
		}
		defer r.Body.Close()

		fmt.Println(calcResult)

		// update postfix ---------------------------------------------------
		idArr := strings.Split(calcResult.Id, "-")
		taskId, _ := strconv.Atoi(idArr[0])
		indexReplaceChunk, _ := strconv.Atoi(idArr[1])

		flagCorrectId := false

		// find and delete task
		for i := 0; i < len(repo.Tasks); i++ {
			if repo.Tasks[i].Id == calcResult.Id {
				if len(repo.Tasks) == 1 {
					repo.Tasks = []op.Task{}
				} else if i == len(repo.Tasks)-1 {
					repo.Tasks = repo.Tasks[:i]
				} else {
					repo.Tasks = append(repo.Tasks[:i], repo.Tasks[:i+1]...)
				}
				flagCorrectId = true
			}
		}

		if !flagCorrectId {
			if err = utils.RespondWith404(w); err != nil {
				logger.Error(err)
			}
			return
		}

		// update postfix
		for i := 0; i < len(repo.RepoE); i++ {
			if repo.RepoE[i].Id == taskId {
				logger.Info("Postfix before replace: ", repo.RepoE[i].ExpressionChunks)

				repo.RepoE[i].ExpressionChunks = append(repo.RepoE[i].ExpressionChunks[:indexReplaceChunk], repo.RepoE[i].ExpressionChunks[indexReplaceChunk+2:]...)

				repo.RepoE[i].ExpressionChunks[indexReplaceChunk].Val = fmt.Sprintf("%g", calcResult.Result)
				repo.RepoE[i].ExpressionChunks[indexReplaceChunk].OpFlag = op.Num
				logger.Info("Postfix after replace: ", repo.RepoE[i].ExpressionChunks)

				// Repeat - add new tasks
				if len(repo.RepoE[i].ExpressionChunks) > 1 {
					// find tasks
					countNewTasks := 0
					for ii := 2; ii < len(repo.RepoE[i].ExpressionChunks); ii++ {
						if repo.RepoE[i].ExpressionChunks[ii].OpFlag == op.Operation && repo.RepoE[i].ExpressionChunks[ii-1].OpFlag == op.Num && repo.RepoE[i].ExpressionChunks[ii-2].OpFlag == op.Num {
							countNewTasks++
							arg1, _ := strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[ii-2].Val, 64)
							arg2, _ := strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[ii-1].Val, 64)
							repo.Tasks = append(repo.Tasks, op.Task{
								Id:            strconv.Itoa(taskId) + "-" + strconv.Itoa(ii-2),
								Arg1:          arg1,
								Arg2:          arg2,
								Operation:     repo.RepoE[i].ExpressionChunks[ii].Val,
								OperationTime: 0,
							})
						}
					}
					logger.Info("Added", countNewTasks, "new tasks")
				} else {
					repo.RepoE[i].Status = "Finished"
					repo.RepoE[i].Result, _ = strconv.Atoi(repo.RepoE[i].ExpressionChunks[0].Val)
				}
			}
		}

		//
		utils.Respond200(w)
		return

	}

	if r.Method == "" || r.Method == http.MethodGet {
		logger.Info("Task GET")
		if len(repo.Tasks) > 0 {
			for i := 0; i < len(repo.Tasks); i++ {
				if !repo.Tasks[i].AtWork {
					repo.Tasks[i].AtWork = true
					payload := repo.Tasks[i]

					respondErr := utils.SuccessRespondWith200(w, payload)
					if respondErr != nil {
						logger.Error(respondErr)
					}
					return
				}
			}
		}

		utils.RespondWith404(w)
		return
	}
}

type Expression struct {
	Id               int        `json:"id"`
	Status           string     `json:"status"`
	Result           int        `json:"result"`
	ExpressionChunks []op.Chunk `json:"-"`
}

type Repo struct {
	RepoE []Expression `json:"expressions"`
	Tasks []op.Task    `json:"-"`
}

func (repo Repo) GetExpression(id int) (Expression, error) {

	for _, expression := range repo.RepoE {
		if expression.Id == id {
			return expression, nil
		}
	}
	return Expression{}, errors.New("wrong id")
}

func main() {
	const (
		defaultHTTPServerWriteTimeout = time.Second * 15
		defaultHTTPServerReadTimeout  = time.Second * 15
	)
	logger.Info("start main orchestrator")
	logger.Info("reading config...")
	conf, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to read config")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	repo := &Repo{}

	mux.HandleFunc("/calculate", repo.AddExpressionHandleFunc)
	mux.HandleFunc("/expressions", repo.GetExpressionsHandleFunc)
	mux.HandleFunc("/expressions/", repo.GetExpressionsHandleFunc) //todo
	mux.HandleFunc("/internal/task", repo.TaskHandleFunc)

	server := &http.Server{
		Handler:      mux,
		Addr:         ":" + strconv.Itoa(conf.ServerPort),
		WriteTimeout: defaultHTTPServerWriteTimeout,
		ReadTimeout:  defaultHTTPServerReadTimeout,
	}

	logger.Info("starting http server...")
	server.ListenAndServe()
}

// expression := "1+2*(3^4-5) ^( 6 +  7*8 )-9" // 1234^5-678*+^*+9-
