package forms

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func HandleReana(input HandlerInput) (FormResult, error) {
	const op = "forms.HandleReana"
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
			} else if slices.Contains(value.revertQuestions, answerNum) {
				countResult := countResults[paramKey]
				countResult.count += maxValuePlusOne - answerValue
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
				level += " (тенденция метизации на неудачу)"
			case count >= conditions[key].middleLevelHighSubLevelStart:
				level += " (тенденция мотивации на успех)"
			}
		}
		resultHTML += fmt.Sprintf("<p>Значение: %v, уровень: %s</p>", count, level)
	}

	return FormResult{
		CouchResult:  PersonalFormResult{BodyText: resultHTML, BodyHTML: resultHTML},
		ClientResult: PersonalFormResult{BodyText: resultHTML, BodyHTML: resultHTML},
	}, nil

}
