package forms

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func HandleBelov(input HandlerInput) (FormResult, error) {
	const op = "forms.HandleBelov"
	const prefix = "answer_"
	const totalAnswersNum = 80
	const flegmaticKey = "Флегматик"
	const flegmaticFooter = "флегматический"
	const melanholicKey = "Меланхолик"
	const melanholicFooter = "меланхолический"
	const holericKey = "Холерик"
	const holericFooter = "холерический"
	const sangvinikKey = "Сангвиник"
	const sangvinikFooter = "сангвинистический"

	answersCount := map[string]struct {
		key     string
		count   int
		percent float64
		postfix string
		answers []int
		footer  string
	}{
		flegmaticKey: {
			answers: []int{1, 5, 9, 13, 17, 21, 25, 29, 33, 37, 41, 45, 49, 53, 57, 61, 65, 69, 73, 77},
			count:   0,
			footer:  flegmaticFooter,
		},
		melanholicKey: {
			answers: []int{2, 6, 10, 14, 18, 22, 26, 30, 34, 38, 42, 46, 50, 54, 58, 62, 66, 70, 74, 78},
			count:   0,
			footer:  melanholicFooter,
		},
		holericKey: {
			answers: []int{3, 7, 11, 15, 19, 23, 27, 31, 35, 39, 43, 47, 51, 55, 59, 63, 67, 71, 75, 79},
			count:   0,
			footer:  holericFooter,
		},
		sangvinikKey: {
			answers: []int{4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64, 68, 72, 76, 80},
			count:   0,
			footer:  sangvinikFooter,
		},
	}

	checkedAnswersNum := make(map[int]struct{})
	var sumCount float64

	req := input.Request

	for qui, data := range req.Answer.Data {
		if !strings.HasPrefix(qui, prefix) {
			continue
		}
		answerNum, err := strconv.Atoi(qui[len(prefix):])
		if err != nil {
			return FormResult{}, fmt.Errorf("%s: %w", op, err)
		}
		isAnswerYes, ok := data.Value.(bool)
		if !ok {
			return FormResult{}, fmt.Errorf("%s: in qui %v expacting value with type bool", op, qui)
		}

		checkedAnswersNum[answerNum] = struct{}{}

		if !isAnswerYes {
			continue
		}
		sumCount += 1
		for key, v := range answersCount {
			if slices.Contains(v.answers, answerNum) {
				v.count += 1
				answersCount[key] = v
			}
		}
		checkedAnswersNum[answerNum] = struct{}{}
	}

	if len(checkedAnswersNum) < totalAnswersNum {
		failedToFind := make([]int, totalAnswersNum-len(checkedAnswersNum))
		for i := 0; i < totalAnswersNum; i++ {
			if _, ok := checkedAnswersNum[i]; !ok {
				failedToFind = append(failedToFind, i)
			}
		}
		return FormResult{}, fmt.Errorf("failed to find all answers: %v", failedToFind)
	}

	for key, v := range answersCount {
		v.percent = float64(v.count) / sumCount * 100
		if v.percent >= 40 {
			v.postfix = " (доминирующий)"
		}
		answersCount[key] = v
	}

	answersOrder := [...]string{flegmaticKey, melanholicKey, holericKey, sangvinikKey}
	couchBodyHTML := prepareTextCouch(answersCount, answersOrder, input.ClientEmail)
	clientBodyHTML := prepareTextClient(answersCount, answersOrder)

	return FormResult{
		CouchResult:  PersonalFormResult{BodyText: couchBodyHTML, BodyHTML: couchBodyHTML},
		ClientResult: PersonalFormResult{BodyText: clientBodyHTML, BodyHTML: clientBodyHTML},
	}, nil
}

func prepareTextCouch(
	count map[string]struct {
		key     string
		count   int
		percent float64
		postfix string
		answers []int
		footer  string
	}, order [4]string, clientEmail string,
) string {
	first := fmt.Sprintf(
		`<h3><b>Тестирование на темперамент А. Белова</b></h3>
<p><span style="font-weight: 400;">Ваш клиент </span><span style="font-weight: 400;">%s</span><span style="font-weight
: 400;"> получил результаты:&nbsp;</span></p>
<h3><b>Результаты тестирования</b><b></b></h3>`, clientEmail,
	)

	middle := ""
	for _, key := range order {
		middle += fmt.Sprintf(
			`<ul>
<li aria-level="1">
<h4><b>%s: <br /><span style="font-weight: 400;"><br /></span><span style="font-weight: 400;">Балл шкалы - %v</span
><span style="font-weight: 400;"><br /></span><span style="font-weight: 400;">Процент выраженности - %.1f%%%s</span><br /></b></h4>
</li>
</ul>
`, key, count[key].count, count[key].percent, count[key].postfix,
		)
	}

	last := fmt.Sprintf(
		`<p><b>Общее распределение: </b><span style="font-weight: 400;"><br /></span><span style
="font-weight: 400;">Темперамент на %.1f%% %s, %.1f%% %s, %.1f%% %s, 
%.1f%% %s.</span><span style="font-weight: 400;"><br /></span></p>`,
		count[order[0]].percent, count[order[0]].footer,
		count[order[1]].percent, count[order[1]].footer,
		count[order[2]].percent, count[order[2]].footer,
		count[order[3]].percent, count[order[3]].footer,
	)
	return first + middle + last
}

func prepareTextClient(
	count map[string]struct {
		key     string
		count   int
		percent float64
		postfix string
		answers []int
		footer  string
	}, order [4]string,
) string {
	first :=
		`<h3><b>Тестирование на темперамент А. Белова</b></h3>
<p><span style="font-weight: 400;">Вы прошли методику и получили следующие результаты тестирования:</span></p>
<h3><b>Результаты тестирования</b><b></b></h3>`

	middle := ""
	for _, key := range order {
		middle += fmt.Sprintf(
			`<ul>
<li aria-level="1">
<h4><b>%s: <br /><span style="font-weight: 400;"><br /></span><span style="font-weight: 400;">Балл шкалы - %v</span
><span style="font-weight: 400;"><br /></span><span style="font-weight: 400;">Процент выраженности - %.1f%%%s</span><br /></b></h4>
</li>
</ul>
`, key, count[key].count, count[key].percent, count[key].postfix,
		)
	}

	last := fmt.Sprintf(
		`<p><b>Общее распределение: </b><span style="font-weight: 400;"><br /></span><span style
="font-weight: 400;">Темперамент на %.1f%% %s, %.1f%% %s, %.1f%% %s, 
%.1f%% %s.</span><span style="font-weight: 400;"><br /></span></p>`,
		count[order[0]].percent, count[order[0]].footer,
		count[order[1]].percent, count[order[1]].footer,
		count[order[2]].percent, count[order[2]].footer,
		count[order[3]].percent, count[order[3]].footer,
	)
	return first + middle + last
}
