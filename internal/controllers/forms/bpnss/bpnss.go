package bpnss

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "bpnss.Handle"
	const avtoKey = "Автономия"
	const kompKey = "Компетентность"
	const prinKey = "Принадлежность"
	const answerPrefix = "answer_"
	const totalAnswersNum = 21
	const maxValuePlusOne = 8

	answers := map[string]int{
		"Полностью не согласен": 1,
		"Не согласен":           2,
		"Скорее не согласен":    3,
		"Затрудняюсь ответить":  4,
		"В целом согласен":      5,
		"Согласен":              6,
		"Полностью согласен":    7,
	}

	conditions := map[string]struct {
		directQuestions  []int
		revertQuestions  []int
		middleLevelStart int
		highLevelStart   int
	}{
		avtoKey: {
			directQuestions:  []int{1, 8, 11, 14, 17},
			revertQuestions:  []int{4, 20},
			middleLevelStart: 30,
			highLevelStart:   44,
		},
		kompKey: {
			directQuestions:  []int{5, 10, 13},
			revertQuestions:  []int{3, 15, 19},
			middleLevelStart: 25,
			highLevelStart:   36,
		},
		prinKey: {
			directQuestions:  []int{2, 9, 12, 21},
			revertQuestions:  []int{6, 7, 16, 18},
			middleLevelStart: 31,
			highLevelStart:   48,
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := map[string]struct {
		count int
		level string
	}{
		avtoKey: {},
		kompKey: {},
		prinKey: {},
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
			if slices.Contains(value.revertQuestions, answerNum) {
				countResult := countResults[paramKey]
				countResult.count += maxValuePlusOne - answerValue
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
			level = "низкий"
		case value.count > conditions[key].middleLevelStart && value.count < conditions[key].highLevelStart:
			level = "средний"
		default:
			level = "высокий"
		}
		value.level = level
		countResults[key] = value
	}

	answersOrder := [...]string{avtoKey, kompKey, prinKey}

	const startText = "<b>Шкала удовлетворения базовых психологических потребностей, BPNSS</b>"
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
	}, order [3]string,
) string {
	result := "<b>Результаты тестирования</b>"
	for _, key := range order {
		result += fmt.Sprintf(
			"<p><b>%s: </b><br/><br/>Балл шкалы - %v<br/>Уровень показателя - %s<br/>", key,
			results[key].count, results[key].level,
		)
	}
	return result
}
