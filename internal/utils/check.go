package utils

import (
	op "arithmetic-expression-calculator/internal/entities"
	"arithmetic-expression-calculator/internal/logger"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type StackStr struct {
	items []string
}

func (s *StackStr) Push(data string) {
	s.items = append(s.items, data)
}

func (s *StackStr) Pop() {
	if s.IsEmpty() {
		return
	}
	s.items = s.items[:len(s.items)-1]
}

func (s *StackStr) Top() (string, error) {
	if s.IsEmpty() {
		return "", fmt.Errorf("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

func (s *StackStr) IsEmpty() bool {
	if len(s.items) == 0 {
		return true
	}
	return false
}

func CheckBalance(exp string) bool {
	s := StackStr{}

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

func CheckBalance2(exp string) bool {
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

func IsOperator(char string) bool {
	switch char {
	case "+":
		return true
	case "-":
		return true
	case "*":
		return true
	case "/":
		return true
	case "^":
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

	parenthesis := CheckBalance(exp)
	logger.Info("expression", exp, ", check parenthesis - (): ", parenthesis)
	if !parenthesis {
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

	if len(result) < 3 {
		return nil, errors.New("invalid expression - short")
	}

	for i := 0; i < len(result)-1; i++ {
		// check expression hasn't operation after open ( and operation before )
		if result[i].OpFlag == result[i+1].OpFlag && !(result[i+1].Val == "(" || result[i].Val == ")") {
			return nil, errors.New("invalid expression repeated")
		}

		// check div zero
		if result[i].Val == "/" && result[i+1].OpFlag == op.Num {
			f, err := strconv.ParseFloat(result[i+1].Val, 64)
			if err != nil {
				return nil, errors.New("invalid expression repeated")
			}
			if f == 0.0 {
				return nil, errors.New("invalid expression div zero")
			}
		}
	}

	return result, nil
}
