package forms

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func HandleUSC(input HandlerInput) (FormResult, error) {
	const op = "forms.HandleUSC"

	const commonKey = "Шкала общей интернальности (Ио)"
	const successKey = "Шкала интернальности в области достижений (Ид)"
	const unluckyKey = "Школа интернальности в области неудач (Ин)"
	const familyKey = "Шкала интернальности в семейных отношениях (Ис)"
	const randomKey = "Шкала интернальности в области производственных отношении (Ип)"
	const interpersonalKey = "Шкала интернальности в области межличностных отношении (Им)"
	const healthKey = "Шкала интернильности в отношении здоровья и болезни (Из)"

	const answerPrefix = "answer_"
	const totalAnswersNum = 44

	answers := map[string]int{
		"Не согласен полностью":            -3,
		"Не согласен частично":             -2,
		"Скорее не согласен, чем согласен": -1,
		"Скорее согласен, чем не согласен": 1,
		"Согласен частично":                2,
		"Согласен полностью":               3,
	}

	conditions := map[string]struct {
		directQuestions    []int
		reversedQuestions  []int
		internalStartValue int
	}{
		commonKey: {
			directQuestions: []int{
				2, 4, 11, 12, 13, 15, 16, 17,
				19, 20, 22, 25, 27, 29, 31, 32,
				34, 36, 37, 39, 42, 44,
			},
			reversedQuestions: []int{
				1, 3, 5, 6, 7, 8, 9, 10, 14,
				18, 21, 23, 24, 26, 28, 30,
				33, 35, 38, 40, 41, 43,
			},
			internalStartValue: 33,
		},
		successKey: {
			directQuestions:    []int{12, 15, 27, 32, 36, 37},
			reversedQuestions:  []int{1, 5, 6, 14, 26, 43},
			internalStartValue: 6,
		},
		unluckyKey: {
			directQuestions:    []int{2, 4, 20, 31, 42, 44},
			reversedQuestions:  []int{7, 24, 33, 38, 40, 41},
			internalStartValue: 8,
		},
		familyKey: {
			directQuestions:    []int{2, 16, 20, 32, 37},
			reversedQuestions:  []int{7, 14, 26, 28, 41},
			internalStartValue: 4,
		},
		randomKey: {
			directQuestions:    []int{19, 22, 25, 31, 42},
			reversedQuestions:  []int{1, 9, 10, 24, 30},
			internalStartValue: 12,
		},
		interpersonalKey: {
			directQuestions:    []int{4, 27},
			reversedQuestions:  []int{6, 38},
			internalStartValue: 2,
		},
		healthKey: {
			directQuestions:    []int{13, 34},
			reversedQuestions:  []int{3, 23},
			internalStartValue: 3,
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := make(map[string]int, len(conditions))

	for qui, data := range input.Request.Answer.Data {
		if !strings.HasPrefix(qui, answerPrefix) {
			continue
		}
		answerNum, err := strconv.Atoi(qui[len(answerPrefix):])
		if err != nil {
			return FormResult{}, fmt.Errorf("%s: %w", op, err)
		}
		vList, ok := data.Value.([]interface{})
		if !ok {
			return FormResult{}, fmt.Errorf("%s: in qui %v expacting value of type []interface{}", op, qui)
		}
		if len(vList) == 0 {
			return FormResult{}, fmt.Errorf("%s: qui %v is empty", op, qui)
		}
		vFirst := vList[0]
		vMap, ok := vFirst.(map[string]interface{})
		if !ok {
			return FormResult{}, fmt.Errorf("%s: in qui %v expacting value of type map[string]interface{}", op, qui)
		}
		valueKey, ok := vMap["text"].(string)
		if !ok {
			return FormResult{}, fmt.Errorf("%s: in qui %v expacting value of type string", op, qui)
		}
		answerValue := answers[valueKey]
		for paramKey, value := range conditions {
			if slices.Contains(value.directQuestions, answerNum) {
				countResults[paramKey] += answerValue
				checkedAnswers[answerNum] = struct{}{}
			}
			if slices.Contains(value.reversedQuestions, answerNum) {
				countResults[paramKey] += -answerValue
				checkedAnswers[answerNum] = struct{}{}
			}
		}
	}

	if len(checkedAnswers) < totalAnswersNum {
		notCheckedAnswers := make([]int, 0, totalAnswersNum-len(checkedAnswers))
		for i := 1; i <= totalAnswersNum; i++ {
			if _, ok := checkedAnswers[i]; !ok {
				notCheckedAnswers = append(notCheckedAnswers, i)
			}
		}
		return FormResult{}, fmt.Errorf(
			"%s: there is not enoght answers in form. not checked: %v",
			op,
			notCheckedAnswers,
		)
	}

	if len(checkedAnswers) > totalAnswersNum {
		return FormResult{}, fmt.Errorf(
			"%s: answers more (%v) then need (%v)", op, len(checkedAnswers),
			totalAnswersNum,
		)
	}

	resultHTML := ""

	for key, value := range countResults {
		resultHTML += fmt.Sprintf("<h1>%s</h1>", key)
		level := "не распознан"

		if value < conditions[key].internalStartValue {
			level = "экстернальность"
		} else {
			level = "интернальность"
		}
		resultHTML += fmt.Sprintf("<p>Значение: %v, уровень: %s</p>", value, level)
	}

	return FormResult{
		CouchResult:  PersonalFormResult{BodyText: resultHTML, BodyHTML: resultHTML},
		ClientResult: PersonalFormResult{BodyText: resultHTML, BodyHTML: resultHTML},
	}, nil
}
