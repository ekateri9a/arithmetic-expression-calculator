package utils

import (
	op "arithmetic-expression-calculator/internal/entities"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Stack []op.Chunk

// IsEmpty: check if stack is empty
func (st *Stack) IsEmpty() bool {
	return len(*st) == 0
}

// Push a new value onto the stack
func (st *Stack) Push(chunk op.Chunk) {
	*st = append(*st, chunk) //Simply append the new value to the end of the stack
}

// Remove top element of stack. Return false if stack is empty.
func (st *Stack) Pop() bool {
	if st.IsEmpty() {
		return false
	} else {
		index := len(*st) - 1 // Get the index of top most element.
		*st = (*st)[:index]   // Remove it from the stack by slicing it off.
		return true
	}
}

// Return top element of stack. Return false if stack is empty.
func (st *Stack) Top() op.Chunk {
	//if st.IsEmpty() {
	//	return nil
	//} else {
	index := len(*st) - 1   // Get the index of top most element.
	element := (*st)[index] // Index onto the slice and obtain the element.
	return element
	//}
}

// Function to return precedence of operators
func prec(s string) int {
	if s == "^" {
		return 3
	} else if (s == "/") || (s == "*") {
		return 2
	} else if (s == "+") || (s == "-") {
		return 1
	} else {
		return -1
	}
}

func InfixToPostfix(infix []op.Chunk) []op.Chunk {
	var sta Stack
	var postfix []op.Chunk
	for _, chunk := range infix {
		if chunk.OpFlag == op.Num {
			postfix = append(postfix, chunk)
		} else if chunk.Val == "(" {
			sta.Push(chunk)
		} else if chunk.Val == ")" {
			for !sta.IsEmpty() && sta.Top().Val != "(" {
				postfix = append(postfix, sta.Top())
				sta.Pop()
			}
			sta.Pop()
		} else {
			for !sta.IsEmpty() && prec(chunk.Val) <= prec(sta.Top().Val) {
				postfix = append(postfix, sta.Top())
				sta.Pop()
			}
			sta.Push(chunk)
		}
	}
	// Pop all the remaining elements from the stack
	for !sta.IsEmpty() {
		postfix = append(postfix, sta.Top())
		sta.Pop()
	}
	return postfix
}

func calculate(arg1, arg2 float64, operation string) (float64, error) {
	result := 0.0
	switch operation {
	case "+":
		result = arg1 + arg2
	case "-":
		result = arg1 - arg2
	case "*":
		result = arg1 * arg2
	case "/":
		if arg2 == 0 {
			return 0.0, fmt.Errorf("div by zero")
		}
		result = arg1 / arg2
	case "^":
		result = math.Pow(arg1, arg2)
	default:
		return 0.0, fmt.Errorf("invalid operator")
	}

	return result, nil
}

func CalculatePostfix(postfix []op.Chunk, id string, tasks *[]op.Task) float64 {
	mm := 200 // todo
	for len(postfix) > 1 && mm > 0 {
		mm--
		//tasks := make([]op.Task, 0)

		// find tasks
		for i := 2; i < len(postfix); i++ {
			if postfix[i].OpFlag == op.Operation && postfix[i-1].OpFlag == op.Num && postfix[i-2].OpFlag == op.Num {
				arg1, _ := strconv.ParseFloat(postfix[i-2].Val, 64)
				arg2, _ := strconv.ParseFloat(postfix[i-1].Val, 64)
				*tasks = append(*tasks, op.Task{
					Id:            id + "-" + strconv.Itoa(i-2),
					Arg1:          arg1,
					Arg2:          arg2,
					Operation:     postfix[i].Val,
					OperationTime: 0,
				})
			}
		}

		fmt.Println("Tasks: ", tasks)

		// calculate
		responses := make([]op.CalcResult, 0)

		for _, task := range *tasks {
			res, _ := calculate(task.Arg1, task.Arg2, task.Operation)
			responses = append(responses, op.CalcResult{
				Id:     task.Id,
				Result: res,
			})
		}

		fmt.Println("Respose: ", responses)

		// replace

		for j, response := range responses {
			rr := strings.Split(response.Id, "-")
			//idt := rr[0]
			pp, _ := strconv.Atoi(rr[1])
			postfix = append(postfix[:pp-j*2], postfix[pp+2-j*2:]...)
			postfix[pp-j*2].Val = fmt.Sprintf("%g", response.Result)
			postfix[pp-j*2].OpFlag = op.Num
		}
		fmt.Println("Postfix: ", postfix)
	}

	result, _ := strconv.ParseFloat(postfix[0].Val, 64)
	return result
}
