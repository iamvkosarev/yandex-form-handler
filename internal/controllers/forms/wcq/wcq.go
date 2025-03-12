package wcq

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"math"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "wcq.Handle"

	const kopingKey = "Конфронтационный копинг"
	const distancingKey = "Дистанцирование"
	const selfControlKey = "Самоконтроль"
	const socialSupportKey = "Поиск социальной поддержки"
	const respAcceptKey = "Принятие ответственности"
	const avoidanceKey = "Бегство-избегание"
	const planingKey = "Планирование решения проблемы"
	const positiveKey = "Положительная переоценка"

	const answerPrefix = "answer_"
	const totalAnswersNum = 50

	answers := map[string]int{
		"Никогда": 0,
		"Редко":   1,
		"Иногда":  2,
		"Часто":   3,
	}

	conditions := map[string]struct {
		questions []int
	}{
		kopingKey: {
			questions: []int{2, 3, 13, 21, 26, 37},
		},
		distancingKey: {
			questions: []int{8, 9, 11, 16, 32, 35},
		},
		selfControlKey: {
			questions: []int{6, 10, 27, 34, 44, 49, 50},
		},
		socialSupportKey: {
			questions: []int{4, 14, 17, 24, 33, 36},
		},
		respAcceptKey: {
			questions: []int{5, 19, 22, 42},
		},
		avoidanceKey: {
			questions: []int{7, 12, 25, 31, 38, 41, 46, 47},
		},
		planingKey: {
			questions: []int{1, 20, 30, 39, 40, 43},
		},
		positiveKey: {
			questions: []int{15, 18, 23, 28, 29, 45, 48},
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
			if slices.Contains(value.questions, answerNum) {
				result := countResults[paramKey]
				result.count += answerValue
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

	const middleStart = 40
	const highStart = 61
	const maxAnswerValue = 3

	for key, value := range countResults {
		level := "не распознан"
		valuePercent := int(math.Round(float64(value.count) / float64(len(conditions[key].questions)*maxAnswerValue) * 100))
		switch {
		case valuePercent < middleStart:

			level = "Редкое использование стратегии"
		case valuePercent >= highStart:
			level = "Выраженное предпочтение стратегии"
		default:
			level = "Умеренное использование стратегии"
		}
		value.level = level
		countResults[key] = value
	}

	answersOrder := [...]string{
		kopingKey,
		distancingKey,
		selfControlKey,
		socialSupportKey,
		respAcceptKey,
		avoidanceKey,
		planingKey,
		positiveKey,
	}

	const startText = "<b>Опросник «Способы совладающего поведения», WCQ</b>"
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
	}, order [8]string,
) string {
	result := "<b>Результаты тестирования</b><br/>"
	for _, key := range order {
		result += fmt.Sprintf(
			"<p><b>%s: </b><br/><br/>Балл - %v<br/><br/>%s<br/><br/>", key,
			results[key].count, results[key].level,
		)
	}
	return result
}
