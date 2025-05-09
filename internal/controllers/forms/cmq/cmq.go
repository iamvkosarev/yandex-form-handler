package cmq

import (
	"fmt"
	"forms-handler/internal/controllers/forms"
	"slices"
	"strconv"
	"strings"
)

func Handle(input forms.HandlerInput) (forms.FormResult, error) {
	const op = "cmq.Handle"
	const startText = "<p><b>Опросник когнитивных ошибок, CMQ</b></p>"
	const scalesNum = 10

	scales := []string{
		"Общая выраженность",
		"Персонализация",
		"Чтение мыслей",
		"Упрямство",
		"Морализация",
		"Катастрофизация",
		"Выученная беспомощность",
		"Максимализм",
		"Преувеличение опасности",
		"Гипернормативность",
	}
	if scalesNum != len(scales) {
		return forms.FormResult{}, fmt.Errorf(
			"%s: not equel num of scales (%v) and elements in slice (%v)",
			op,
			scalesNum, len(scales),
		)
	}

	const answerPrefix = "answer_"
	const totalAnswersNum = 45
	const maxValuePlusOne = 5

	answers := map[string]int{
		"Никогда": 1,
		"Иногда":  2,
		"Часто":   3,
		"Всегда":  4,
	}

	stages := []string{
		"низкая",
		"высокая",
	}

	conditions := [scalesNum]struct {
		directQuestions           []int
		revertQuestions           []int
		stagesFromSecondMinLevels []int
	}{
		{
			directQuestions:           []int{},
			revertQuestions:           []int{27, 41},
			stagesFromSecondMinLevels: []int{94},
		},
		{
			directQuestions:           []int{13, 15, 16, 19, 20},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{6},
		},
		{
			directQuestions:           []int{6, 8, 9, 14, 17},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{9},
		},
		{
			directQuestions:           []int{23, 42, 43, 44, 45},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{12},
		},
		{
			directQuestions:           []int{11, 12, 21, 39, 40},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{13},
		},
		{
			directQuestions:           []int{1, 2, 3, 10, 25},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{9},
		},
		{
			directQuestions:           []int{4, 5, 26, 28, 29, 34, 35, 36, 38},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{16},
		},
		{
			directQuestions:           []int{18, 21, 22, 23, 24, 25},
			revertQuestions:           []int{},
			stagesFromSecondMinLevels: []int{11},
		},
		{
			directQuestions:           []int{9, 23, 31, 33, 34, 35},
			revertQuestions:           []int{27},
			stagesFromSecondMinLevels: []int{14},
		},
		{
			directQuestions:           []int{32, 33, 37, 40},
			revertQuestions:           []int{41},
			stagesFromSecondMinLevels: []int{13},
		},
	}

	for i := 1; i <= totalAnswersNum; i++ {
		isRevert := false
		for _, revertQuestion := range conditions[0].revertQuestions {
			if revertQuestion == i {
				isRevert = true
				break
			}
		}
		if isRevert {
			continue
		}
		commonCondition := conditions[0]
		commonCondition.directQuestions = append(commonCondition.directQuestions, i)
		conditions[0] = commonCondition
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
		level := ""
		for stageI := 0; stageI < len(stages)-1; stageI++ {
			nextStageStart := conditions[key].stagesFromSecondMinLevels[stageI]
			if value.count < nextStageStart {
				level = stages[stageI]
				break
			}
		}
		if level == "" {
			level = stages[len(stages)-1]
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

	descriptions := []string{
		"",
		"<br/>Ожидание враждебного и неодобрительного отношения" +
			" к себе; любое несогласие или замечание воспринимаются" +
			" как отвержение, подавление или унижение.<br/>Проявления:" +
			" фиксация на критических замечаниях и несогласии с другими," +
			" обидчивость, недоверие и настороженность, ожидание негативно " +
			"пристрастного отношения к себе, упреков, обмана, отвержения" +
			" и унижения со стороны других.",
		"<br/>Склонность приходить к недостаточно логически обоснованным " +
			"выводам; тенденция «додумывать» за других людей, опираясь на " +
			"субъективные ожидания, интуитивные оценки и проекции.<br/>" +
			"Проявления: односторонние, как правило, негативные суждения " +
			"о намерениях, поступках и оценках других людей; неумение " +
			"логически и с различных сторон рассмотреть причины и " +
			"обстоятельства поведения окружающих.",
		"<br/>Настойчивое стремление отстаивать свою самооценку, связанное со " +
			"страхом ошибиться, эгоцентрическая иерархизация и сужение " +
			"проблемного поля.<br/>Проявления: негибкость суждений, " +
			"преобладание эгоцентрических защитных суждений и бездействия, " +
			"склонность явно или скрыто оспаривать мнение и предложения других " +
			"людей «из принципа», отождествляя себя с предметом спора.",
		"<br/>Декларирование повышенной моральной ответственности, " +
			"стремление к обеспечению безопасности за счет морального контроля над окружающими." +
			"<br/>Проявления: преобладание моральных суждений и оценок в восприятии явлений и людей, " +
			"представляющих потенциальное неудобство или опасность.",
		"<br/>Склонность преувеличивать значимость проблем и бурно на них реагировать, как правило, " +
			"вследствие прямого столкновения идеализированных представлений о себе и окружающих с реальностью." +
			"<br/>Проявления: обостренное, негативно преувеличенное реагирование на проблемы, " +
			"выражающееся в нереалистичном ожидании угрозы жизни, здоровью, благосостоянию, общественному " +
			"статусу, в потере доверия и в уверенности в обмане со стороны других; склонность к " +
			"аффективно-шоковым и диссоциативным реакциям.",
		"<br/>Обесценивание собственного «я», принижение своих возможностей и способностей, " +
			"сопряженное со стремлением снять с себя ответственность за жизненные неудачи, " +
			"и декларирование пессимистической установки." +
			"<br/>Проявления: повторяющееся очевидное обесценивание своих возможностей, положения " +
			"и достижений, стремление к поиску защиты и покровительства, декларирование своей " +
			"слабости и беспомощности как оправдание неудач и нежелания активно преодолевать " +
			"имеющиеся затруднения.",
		"<br/>Амбициозность и крайность в оценках, потребность в восхищении, " +
			"выражающаяся через нарциссическую безупречность.<br/>Проявления: крайность " +
			"в суждениях, " +
			"тенденция преувеличивать свои достижения и упрекать окружающих в их недооценке, " +
			"комплекс Золушки (фрустрация ожиданий восхищения как награды за трудолюбие " +
			"и безупречность), обесценивание других за лень и необязательность.",
		"<br/>Уклонение от непредвиденных обстоятельств, избегание рисков, " +
			"ответственности и соперничества вследствие преувеличения опасностей." +
			"<br/>" +
			"Проявления: самоограничения и повышенный самоконтроль со ссылками на " +
			"многочисленные или преувеличенные опасности, неблагоприятные обстоятельства " +
			"и/или недоброжелательное отношение; избегающая осторожность и пассивность.",
		"<br/>Отождествление себя с социальными нормами, перфекционизм, " +
			"стремление обезопасить себя за счет тщательного следования нормам и социальным предписаниям." +
			"<br/>Проявления: безусловная и не всегда критичная приверженность правилам, " +
			"нормам поведения и этикету, " +
			"исполнительность и тщательность, избыточная вежливость и аккуратность во " +
			"взаимоотношениях, тенденция к вынесению оценок исходя из принятых в данной " +
			"группе социальных нормативов.",
	}
	for i, scale := range scales {
		result += fmt.Sprintf(
			"<p><b>%s:</b>%s</p><p>Балл шкалы - %v<br/>Выраженность - %s</p>",
			scale, descriptions[i], results[i].count, results[i].level,
		)
	}
	return result
}
