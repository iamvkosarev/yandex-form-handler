package gse

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "gse.Handle"
	const selfEffectiveKey = "Самоэффективность"
	const answerPrefix = "answer_"
	const totalAnswersNum = 10

	answers := map[string]int{
		"Абсолютно неверно":  1,
		"Едва ли это верно":  2,
		"Скорее всего верно": 3,
		"Совершенно верно":   4,
	}

	conditions := map[string]struct {
		directQuestions  []int
		middleLevelStart int
		highLevelStart   int
		lowLevelText     string
		middleLevelText  string
		highLevelText    string
	}{
		selfEffectiveKey: {
			directQuestions: []int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
			},
			middleLevelStart: 25,
			highLevelStart:   36,
			lowLevelText:     "Низкий уровень самоэффективности",
			middleLevelText:  "Средний уровень самоэффективности",
			highLevelText:    "Высокий уровень самоэффективности",
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := map[string]struct {
		count int
		level string
	}{
		selfEffectiveKey: {},
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

	const startText = "<b>Шкала общей самоэффективности, GSE</b>"
	resultText := getResultText(countResults[selfEffectiveKey])
	couchBodyHTML := startText + forms.GetTextCouch(input.ClientEmail) + resultText
	clientBodyHTML := startText + forms.GetTextClient() + resultText

	return forms.FormResult{
		CouchResult:  forms.PersonalFormResult{BodyText: couchBodyHTML, BodyHTML: couchBodyHTML},
		ClientResult: forms.PersonalFormResult{BodyText: clientBodyHTML, BodyHTML: clientBodyHTML},
	}, nil

}

func getResultText(
	value struct {
		count int
		level string
	},
) string {
	result := "<b>Результаты тестирования</b>"
	result += fmt.Sprintf(
		"<p>Балл - %v<br/><br>%s<br/></p>",
		value.count, value.level,
	)
	return result
}
