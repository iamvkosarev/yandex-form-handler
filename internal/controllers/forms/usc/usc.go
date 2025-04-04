package usc

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "usc.Handle"

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
		maxStanValue       []int
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
			maxStanValue:       []int{-14, -3, 9, 21, 32, 44, 56, 68, 79, 132},
		},
		successKey: {
			internalStartValue: 6,
			directQuestions:    []int{12, 15, 27, 32, 36, 37},
			reversedQuestions:  []int{1, 5, 6, 14, 26, 43},
			maxStanValue:       []int{-11, -7, -3, 1, 5, 9, 14, 18, 22, 36},
		},
		unluckyKey: {
			internalStartValue: 8,
			directQuestions:    []int{2, 4, 20, 31, 42, 44},
			reversedQuestions:  []int{7, 24, 33, 38, 40, 41},
			maxStanValue:       []int{-8, -4, 0, 4, 7, 11, 15, 19, 23, 36},
		},
		familyKey: {
			internalStartValue: 4,
			directQuestions:    []int{2, 16, 20, 32, 37},
			reversedQuestions:  []int{7, 14, 26, 28, 41},
			maxStanValue:       []int{-12, -8, -5, -1, 3, 6, 10, 13, 17, 30},
		},
		randomKey: {
			internalStartValue: 12,
			directQuestions:    []int{19, 22, 25, 31, 42},
			reversedQuestions:  []int{1, 9, 10, 24, 30},
			maxStanValue:       []int{-5, -1, 3, 7, 11, 15, 19, 23, 27, 30},
		},
		interpersonalKey: {
			internalStartValue: 2,
			directQuestions:    []int{4, 27},
			reversedQuestions:  []int{6, 38},
			maxStanValue:       []int{-7, -5, -3, -1, 1, 4, 6, 8, 10, 12},
		},
		healthKey: {
			internalStartValue: 3,
			directQuestions:    []int{13, 34},
			reversedQuestions:  []int{3, 23},
			maxStanValue:       []int{-6, -4, -2, 0, 2, 4, 6, 8, 10, 12},
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := make(
		map[string]struct {
			count int
			level string
			stan  int
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
				result.count -= answerValue
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
		level := "не распознан"
		stan := 1

		if value.count < conditions[key].internalStartValue {
			level = "Экстернальность"
		} else {
			level = "Интернальность"
		}
		for i, stanMax := range conditions[key].maxStanValue {
			if value.count <= stanMax {
				stan = i + 1
				break
			}
		}
		value.level = level

		value.stan = stan
		countResults[key] = value
	}

	answersOrder := [...]string{successKey, unluckyKey, familyKey, randomKey, interpersonalKey, healthKey}

	const startText = "<b>Уровень субъективного контроля, УСК</b>"
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
		stan  int
	}, commonKey string, order [6]string,
) string {
	result := "<p><b>Результаты тестирования</b></p>"
	result += fmt.Sprintf(
		"<p>%s</p><p>Балл: %v<br/>Стен: %v</p><p>%s</p><br/>", commonKey,
		results[commonKey].count, results[commonKey].stan, results[commonKey].level,
	)
	result += "<br/><br/><b>По шкалам:</b><br/>"
	for _, key := range order {
		result += fmt.Sprintf(
			"<p><b>%s: </b></p><p>Балл: %v<br/>Стен: %v</p><p>%s</p><br/>", key,
			results[key].count, results[key].stan, results[key].level,
		)
	}
	return result
}
