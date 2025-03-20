package ego

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "ego.Handle"

	const criticalParentKey = "Критический Родитель"
	const сaringParentKey = "Заботливый Родитель"
	const adultParentKey = "Взрослый"
	const naturalChildKey = "Естественный Ребенок"
	const adaptedChildKey = "Адаптированный Ребенок"
	const rebelliousChildKey = "Бунтующий Ребенок"

	const answerPrefix = "answer_"
	const totalAnswersNum = 60

	answers := map[string]int{
		"едва":      1,
		"немного":   2,
		"примерно":  3,
		"почти":     4,
		"полностью": 5,
	}

	conditions := map[string]struct {
		directQuestions []int
	}{
		criticalParentKey: {
			directQuestions: []int{
				6, 8, 15, 22, 28, 36, 40, 45, 51, 56,
			},
		},
		сaringParentKey: {
			directQuestions: []int{
				3, 7, 17, 21, 26, 35, 42, 46, 54, 57,
			},
		},
		adultParentKey: {
			directQuestions: []int{
				5, 12, 16, 20, 29, 33, 39, 43, 53, 58,
			},
		},
		naturalChildKey: {
			directQuestions: []int{
				1, 11, 13, 23, 30, 31, 38, 48, 52, 59,
			},
		},
		adaptedChildKey: {
			directQuestions: []int{
				4, 10, 14, 24, 25, 32, 41, 47, 50, 55,
			},
		},
		rebelliousChildKey: {
			directQuestions: []int{
				2, 9, 18, 19, 27, 34, 37, 44, 49, 60,
			},
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := make(
		map[string]struct {
			count   int
			percent float64
		}, len(conditions),
	)

	totalResult := 0

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
				totalResult += answerValue
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

	for key, value := range countResults {
		value.percent = float64(value.count) / float64(totalResult) * 100
		countResults[key] = value
	}

	answersOrder := [...]string{
		criticalParentKey,
		сaringParentKey,
		adultParentKey,
		naturalChildKey,
		adaptedChildKey,
		rebelliousChildKey,
	}

	const startText = "<b>Тест-эгограмма Д.Дюсея</b>"
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
		count   int
		percent float64
	}, order [6]string,
) string {
	result := "<br/><b>Результаты тестирования</b>"
	for _, key := range order {
		result += fmt.Sprintf(
			"<p><b>%s: </b><br/><br/>Балл шкалы - %v<br/>Процентная выраженность - %.1f%%<br/><br/>", key,
			results[key].count, results[key].percent,
		)
	}
	return result
}
