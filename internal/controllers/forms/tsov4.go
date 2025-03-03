package forms

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func HandleTSOV4(input HandlerInput) (FormResult, error) {
	const op = "forms.HandleTSOV4"
	const polenezKey = "Поленезависимость"
	const lieKey = "Лживость"
	const answerPrefix = "answer_"
	const totalAnswersNum = 54

	answers := map[string]int{
		"Нет, это не так":  1,
		"Пожалуй, так":     2,
		"Верно":            3,
		"Совершенно верно": 4,
	}

	conditions := map[string]struct {
		directQuestions  []int
		middleLevelStart int
		highLevelStart   int
		lowLevelText     string
		middleLevelText  string
		highLevelText    string
	}{
		polenezKey: {
			directQuestions: []int{
				1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 13, 14, 15,
				16, 17, 18, 19, 20, 22, 23, 24, 25, 26, 27, 28, 29, 31, 32, 33,
				34, 35, 36, 37, 38, 40, 41, 42, 43, 44, 45, 46, 47, 49, 50, 51,
				52, 53, 54,
			},
			middleLevelStart: 97,
			highLevelStart:   144,
			lowLevelText:     "высокий уровень поленезависимости",
			middleLevelText:  "средний уровень развития поленезависимости",
			highLevelText:    "низкий уровень поленезависимости (полезависимость)",
		},
		lieKey: {
			directQuestions:  []int{3, 12, 21, 30, 39, 48},
			middleLevelStart: 13,
			highLevelStart:   19,
			lowLevelText:     "результаты можно использовать безоговорочно",
			middleLevelText:  "результаты можно использовать условно",
			highLevelText:    "результаты теста недостоверны",
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := map[string]struct {
		count int
		level string
	}{
		polenezKey: {},
		lieKey:     {},
	}

AnswerLoop:
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
				countResult := countResults[paramKey]
				countResult.count += answerValue
				countResults[paramKey] = countResult
				checkedAnswers[answerNum] = struct{}{}
				continue AnswerLoop
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
		switch {
		case value.count < conditions[key].middleLevelStart:
			level = conditions[key].lowLevelText
		case value.count >= conditions[key].highLevelStart:
			level = conditions[key].highLevelText
		default:
			level = conditions[key].middleLevelText
		}
		resultHTML += fmt.Sprintf("<p>Значение: %v, уровень: %s</p>", value.count, level)
	}

	return FormResult{
		CouchResult:  PersonalFormResult{BodyText: resultHTML, BodyHTML: resultHTML},
		ClientResult: PersonalFormResult{BodyText: resultHTML, BodyHTML: resultHTML},
	}, nil

}
