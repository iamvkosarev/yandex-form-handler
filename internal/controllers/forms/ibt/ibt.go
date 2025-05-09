package ibt

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "ibt.Handle"
	const startText = "<p><b>Тест иррациональных убеждений, IBT</b></p>"
	const scalesNum = 10

	scales := []string{
		"Потребность в одобрении",
		"Высокие самоожидания",
		"Склонность к обвинениям",
		"Низкая фрустрационная толерантность",
		"Эмоциональная безответственность",
		"Тревожная сверхозабоченность",
		"Избегание проблем",
		"Зависимость от других",
		"Беспомощность в отношении изменений",
		"Перфекционизм",
	}
	if scalesNum != len(scales) {
		return forms.FormResult{}, fmt.Errorf(
			"%s: not equel num of scales (%v) and elements in slice (%v)",
			op,
			scalesNum, len(scales),
		)
	}

	const answerPrefix = "answer_"
	const totalAnswersNum = 100
	const maxValuePlusOne = 6

	answers := map[string]int{
		"Совершенно не согласен":    1,
		"Относительно не согласен":  2,
		"Не знаю, согласен или нет": 3,
		"Относительно согласен":     4,
		"Совершенно согласен":       5,
	}

	conditions := [scalesNum]struct {
		directQuestions  []int
		revertQuestions  []int
		middleLevelStart int
		highLevelStart   int
	}{
		{
			directQuestions:  []int{1, 21, 51, 71, 81},
			revertQuestions:  []int{11, 31, 41, 61, 91},
			middleLevelStart: 22,
			highLevelStart:   38,
		},
		{
			directQuestions:  []int{2, 12, 42, 62, 72, 82},
			revertQuestions:  []int{22, 32, 52, 92},
			middleLevelStart: 22,
			highLevelStart:   38,
		},
		{
			directQuestions:  []int{3, 13, 23, 33, 53, 73},
			revertQuestions:  []int{43, 63, 83, 93},
			middleLevelStart: 24,
			highLevelStart:   37,
		},
		{
			directQuestions:  []int{24, 34, 84},
			revertQuestions:  []int{4, 14, 44, 54, 64, 74, 94},
			middleLevelStart: 22,
			highLevelStart:   37,
		},
		{
			directQuestions:  []int{55, 75},
			revertQuestions:  []int{5, 15, 25, 35, 45, 65, 85, 95},
			middleLevelStart: 20,
			highLevelStart:   32,
		},
		{
			directQuestions:  []int{6, 16, 26, 46, 66, 76, 96},
			revertQuestions:  []int{36, 56, 86},
			middleLevelStart: 22,
			highLevelStart:   40,
		},
		{
			directQuestions:  []int{7, 27, 37, 47, 67},
			revertQuestions:  []int{17, 57, 77, 87, 97},
			middleLevelStart: 19,
			highLevelStart:   32,
		},
		{
			directQuestions:  []int{8, 18, 28, 38, 78},
			revertQuestions:  []int{48, 58, 68, 88, 98},
			middleLevelStart: 26,
			highLevelStart:   36,
		},
		{
			directQuestions:  []int{9, 19, 49, 69, 79, 89},
			revertQuestions:  []int{29, 39, 59, 99},
			middleLevelStart: 19,
			highLevelStart:   32,
		},
		{
			directQuestions:  []int{10, 30, 40, 50, 90},
			revertQuestions:  []int{20, 60, 70, 80, 100},
			middleLevelStart: 22,
			highLevelStart:   34,
		},
	}

	checkedAnswers := make(map[int]struct{})
	countResults := make(
		[]struct {
			count int
			level string
		}, scalesNum,
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
		for scaleIndex, value := range conditions {
			if slices.Contains(value.directQuestions, answerNum) {
				countResult := countResults[scaleIndex]
				countResult.count += answerValue
				countResults[scaleIndex] = countResult
				checkedAnswers[answerNum] = struct{}{}
			}
			if slices.Contains(value.revertQuestions, answerNum) {
				countResult := countResults[scaleIndex]
				countResult.count += maxValuePlusOne - answerValue
				countResults[scaleIndex] = countResult
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
		case value.count >= conditions[key].middleLevelStart && value.count < conditions[key].highLevelStart:
			level = "средний"
		default:
			level = "высокий"
		}
		value.level = level
		countResults[key] = value
	}

	resultText := getResultText(countResults, scales)
	couchBodyHTML := startText + forms.GetTextCouch(input.ClientEmail) + resultText
	clientBodyHTML := startText + forms.GetTextClient() + resultText

	return forms.FormResult{
		CouchResult:  forms.PersonalFormResult{BodyText: couchBodyHTML, BodyHTML: couchBodyHTML},
		ClientResult: forms.PersonalFormResult{BodyText: clientBodyHTML, BodyHTML: clientBodyHTML},
	}, nil
}

func getResultText(
	results []struct {
		count int
		level string
	}, scales []string,
) string {
	result := "<p><b>Результаты тестирования</b></p>"
	result += "<p><i>*Низкий балл указывает на более " +
		"рациональный уровень убеждений в каждой " +
		"сфере, а высокий балл – на более иррациональный.</i></p>"

	descriptions := []string{
		"Вера в то, что вы нуждаетесь в поддержке и одобрении каждого, кого вы знаете или о ком заботитесь.",
		"Вера в то,  что вы должны быть удачливы, успешны и компетентны в любом деле, за которое беретесь, " +
			"и вы судите о своей ценности как личности на основе успешности ваших достижений.",
		"Вера в то, что все люди, включая вас, заслуживают обвинений и наказаний за их ошибки и проступки.",
		"Вера в то, что когда события разворачиваются не так, как им следовало бы быть, то это совершенно ужасно," +
			" кошмарно и катастрофично. Поэтому чувствуется, что это нормально – расстраиваться, " +
			"когда события складываются не в вашу пользу, или когда люди ведут себя не так, как вам того хочется.",
		"Убеждение, что вы слабо контролируете свои неудачи, неприятности, " +
			"эмоциональные расстройства. Все они происходят по вине других людей или событий в этом мире. Если бы только «они» (события или люди) изменились, вы бы чувствовали себя хорошо и всё было бы нормально.",
		"Вера в то, что может случиться нечто плохое или опасное, " +
			"поэтому вы должны быть крайне озабочены этим и побеспокоиться о предотвращении возможности возникновения этого.",
		"Убеждение, что гораздо легче избегать определенных трудностей и ответственности, " +
			"и вместо этого сначала сделать то, что гораздо приятнее.",
		"Вера в то, что должны иметь кого-нибудь более сильного, чем вы, на кого можно положиться.",
		"Убеждение, что вы являетесь результатом развития вашей собственной истории, " +
			"и поэтому вы почти ничего не можете сделать, чтобы предотвратить ее влияние. «Мой путь – это и есть я, и я ничего не могу с этим поделать». Поэтому вы убеждены, что неспособны измениться.",
		"Вера в то, что каждая проблема имеет «правильное» или безупречное решение. И в дальнейшем, " +
			"пока вы не найдете это совершенное решение, вы не можете быть удовлетворены или счастливы.",
	}
	for i, scale := range scales {
		result += fmt.Sprintf(
			"<p><b>%s:</b><br/>%s</p><p>Балл шкалы - %v<br/>Уровень - %s</p>",
			scale, descriptions[i], results[i].count, results[i].level,
		)
	}
	return result
}
