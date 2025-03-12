package spb

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "spb.Handle"

	const commonKey = "Общий балл"
	const disasterKey = "Катастрофизация"
	const selfStatusKey = "Должествование в отношении себя"
	const otherStatusKey = "Должествование в отношении других"
	const toleranceKey = "Фрустрационная толерантность"
	const evaluationKey = "Оценочная установка"

	const answerPrefix = "answer_"
	const totalAnswersNum = 50
	const maxAnswerValuePlusOne = 7

	answers := map[string]int{
		"Полностью согласен":     1,
		"В основном согласен":    2,
		"Слегка согласен":        3,
		"Слегка не согласен":     4,
		"В основном не согласен": 5,
		"Полностью не согласен":  6,
	}

	conditions := map[string]struct {
		directQuestions   []int
		reversedQuestions []int
	}{
		commonKey: {
			directQuestions: []int{
				2,
				3,
				5,
				6,
				7,
				8,
				9,
				10,
				11,
				12,
				14,
				15,
				16,
				18,
				19,
				21,
				23,
				24,
				27,
				29,
				30,
				31,
				32,
				33,
				35,
				36,
				37,
				39,
				40,
				41,
				43,
				44,
				45,
				47,
				48,
				50,
			},
			reversedQuestions: []int{1, 4, 13, 17, 20, 22, 25, 26, 28, 34, 38, 42, 46, 49},
		},
		disasterKey: {
			directQuestions:   []int{6, 11, 16, 21, 31, 36, 41},
			reversedQuestions: []int{1, 26, 46},
		},
		selfStatusKey: {
			directQuestions:   []int{2, 7, 12, 27, 32, 37, 47},
			reversedQuestions: []int{17, 22, 42},
		},
		otherStatusKey: {
			directQuestions:   []int{3, 8, 18, 23, 33, 43, 48},
			reversedQuestions: []int{13, 28, 38},
		},
		toleranceKey: {
			directQuestions:   []int{9, 14, 19, 24, 29, 39, 44},
			reversedQuestions: []int{4, 34, 49},
		},
		evaluationKey: {
			directQuestions:   []int{5, 10, 15, 30, 35, 40, 45, 50},
			reversedQuestions: []int{20, 25},
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := make(
		map[string]struct {
			count int
			level string
		}, len(conditions),
	)

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
				result := countResults[paramKey]
				result.count += answerValue
				countResults[paramKey] = result
				checkedAnswers[answerNum] = struct{}{}
			}
			if slices.Contains(value.reversedQuestions, answerNum) {
				result := countResults[paramKey]
				result.count += maxAnswerValuePlusOne - answerValue
				countResults[paramKey] = result
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

	resultHTML := ""

	const commonMiddleStart = 150
	const commonHighStart = 230

	commonValue := countResults[commonKey]
	resultHTML += fmt.Sprintf("<h1>%s</h1>", commonKey)
	level := "не распознан"
	switch {
	case commonValue.count < commonMiddleStart:
		level = "Ярко выраженное наличие иррациональной установки"
	case commonValue.count >= commonHighStart:
		level = "Иррациональные установки отсутствуют"
	default:
		level = "Средняя вероятность наличия иррациональной установки"
	}
	commonValue.level = level
	countResults[commonKey] = commonValue

	const middleStart = 30
	const highStart = 45

	for key, value := range countResults {
		if key == commonKey {
			continue
		}
		resultHTML += fmt.Sprintf("<h1>%s</h1>", key)
		level := "не распознан"
		switch {
		case value.count < middleStart:
			level = "Выраженное наличие иррациональной установки"
		case value.count >= highStart:
			level = "Отсутствие иррациональной установки"
		default:
			level = "Иррациональная установка присутствует"
		}
		value.level = level
		countResults[key] = value
	}

	answersOrder := [...]string{disasterKey, selfStatusKey, otherStatusKey, toleranceKey, evaluationKey}

	const startText = "<b>Методика диагностики иррациональных установок, SPB</b>"
	resultText := getResultText(countResults, commonKey, answersOrder)
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
	}, commonKey string, order [5]string,
) string {
	result := "<b>Результаты тестирования</b>"
	result += fmt.Sprintf(
		"<p><b>%s: </b><br/><br/>Общий балл - %v<br/><br/>%s<br/>", commonKey,
		results[commonKey].count, results[commonKey].level,
	)
	result += "<br/><br/><b>По шкалам:</b><br/>"
	for _, key := range order {
		result += fmt.Sprintf(
			"<p><b>%s: </b><br/><br/>Балл - %v<br/><br/>%s<br/><br/>", key,
			results[key].count, results[key].level,
		)
	}
	return result
}
