package tsov4

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "tsov4.Handle"
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
			lowLevelText:     "Высокий уровень поленезависимости",
			middleLevelText:  "Средний уровень развития поленезависимости",
			highLevelText:    "Высокий уровень полезависимости",
		},
		lieKey: {
			directQuestions:  []int{3, 12, 21, 30, 39, 48},
			middleLevelStart: 13,
			highLevelStart:   19,
			lowLevelText:     "(результаты можно использовать безоговорочно)",
			middleLevelText:  "(результаты можно использовать условно)",
			highLevelText:    "(результаты теста недостоверны)",
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

	for qui, data := range input.Request.Answer.Data {
		if !strings.HasPrefix(qui, answerPrefix) {
			continue
		}
		answerNum, err := strconv.Atoi(qui[len(answerPrefix):])
		if err != nil {
			return forms.FormResult{}, fmt.Errorf("%s: %w", op, err)
		}
		vList, ok := data.Value.([]interface{})
		if !ok {
			return forms.FormResult{}, fmt.Errorf("%s: in qui %v expacting value of type []interface{}", op, qui)
		}
		if len(vList) == 0 {
			return forms.FormResult{}, fmt.Errorf("%s: qui %v is empty", op, qui)
		}
		vFirst := vList[0]
		vMap, ok := vFirst.(map[string]interface{})
		if !ok {
			return forms.FormResult{}, fmt.Errorf(
				"%s: in qui %v expacting value of type map[string]interface{}",
				op,
				qui,
			)
		}
		valueKey, ok := vMap["text"].(string)
		if !ok {
			return forms.FormResult{}, fmt.Errorf("%s: in qui %v expacting value of type string", op, qui)
		}
		answerValue := answers[valueKey]
		for paramKey, value := range conditions {
			if slices.Contains(value.directQuestions, answerNum) {
				countResult := countResults[paramKey]
				countResult.count += answerValue
				countResults[paramKey] = countResult
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
		return forms.FormResult{}, fmt.Errorf(
			"%s: there is not enoght answers in form. not checked: %v",
			op,
			notCheckedAnswers,
		)
	}

	if len(checkedAnswers) > totalAnswersNum {
		return forms.FormResult{}, fmt.Errorf(
			"%s: answers more (%v) then need (%v)", op, len(checkedAnswers),
			totalAnswersNum,
		)
	}

	for key, value := range countResults {
		level := "не распознан"
		switch {
		case value.count < conditions[key].middleLevelStart:
			level = conditions[key].lowLevelText
		case value.count >= conditions[key].highLevelStart:
			level = conditions[key].highLevelText
		default:
			level = conditions[key].middleLevelText
		}
		value.level = level
		countResults[key] = value
	}

	answersOrder := [...]string{polenezKey, lieKey}

	const startText = "<b>Диагностика когнитивного стиля ТСОВ-4</b>"
	resultText := getResultText(countResults, answersOrder)
	couchBodyHTML := startText + forms.GetTextCouch(input.ClientEmail) + resultText
	clientBodyHTML := startText + forms.GetTextClient() + resultText

	return forms.FormResult{
		CouchResult:  forms.PersonalFormResult{BodyText: couchBodyHTML, BodyHTML: couchBodyHTML},
		ClientResult: forms.PersonalFormResult{BodyText: clientBodyHTML, BodyHTML: clientBodyHTML},
	}, nil
}

func getResultText(
	results map[string]struct {
		count int
		level string
	}, order [2]string,
) string {
	result := "<b>Результаты тестирования</b>"
	poleResult := results[order[0]]
	lieResult := results[order[1]]
	result += fmt.Sprintf(
		"<p>Балл - %v<br/><br/>%s<br/><br/>%s", poleResult.count,
		poleResult.level, lieResult.level,
	)

	return result
}
