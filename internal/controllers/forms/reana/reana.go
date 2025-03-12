package reana

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "reana.Handle"
	const unluckKey = "мотивация на неудачу (боязнь неудачи)"
	const answerPrefix = "answer_"
	const totalAnswersNum = 20
	const maxValuePlusOne = 1

	answers := map[string]int{
		"Да":  1,
		"Нет": 0,
	}

	conditions := map[string]struct {
		directQuestions                []int
		revertQuestions                []int
		middleLevelStart               int
		highLevelStart                 int
		middleLevelMiddleSubLevelStart int
		middleLevelHighSubLevelStart   int
	}{
		unluckKey: {
			directQuestions:                []int{1, 2, 3, 6, 8, 10, 11, 12, 14, 16, 18, 19, 20},
			revertQuestions:                []int{4, 5, 7, 9, 13, 15, 17.},
			middleLevelStart:               8,
			highLevelStart:                 14,
			middleLevelMiddleSubLevelStart: 10,
			middleLevelHighSubLevelStart:   12,
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := map[string]struct {
		count int
		level string
	}{
		unluckKey: {},
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
		count := value.count
		switch {
		case count < conditions[key].middleLevelStart:
			level = "мотивация на неудачу (боязнь неудачи)"
		case count >= conditions[key].highLevelStart:
			level = "мотивация на успех (надежда на успех)"
		default:
			level = "мотивационный полюс ярко не выражен"
			switch {
			case count < conditions[key].middleLevelMiddleSubLevelStart:
				level += ": тенденция метизации на неудачу"
			case count >= conditions[key].middleLevelHighSubLevelStart:
				level += ": тенденция мотивации на успех"
			}
		}
		value.level = level
		countResults[key] = value
	}

	const startText = "<b>Опросник А. Реана «Мотивация успеха и боязнь неудачи» (МУН)</b>"
	resultText := getResultText(countResults[unluckKey])
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
		"<p>Балл - %v<br/><br>Описание - %s<br/></p>",
		value.count, value.level,
	)
	return result
}
