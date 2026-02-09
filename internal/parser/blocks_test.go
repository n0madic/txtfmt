package parser

import (
	"strings"
	"testing"

	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

func TestIsDialogueLine(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"- Привет", true},
		{"— Привет", true},
		{"  \t- Привет", true},
		{"-", false},
		{"-- Привет", false},
		{"Текст", false},
	}
	for _, tc := range cases {
		if got := isDialogueLine(tc.in); got != tc.want {
			t.Fatalf("isDialogueLine(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsSceneBreak(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"***", true},
		{"* * *", true},
		{"— * —", true},
		{"--***--", true},
		{"-----", true},
		{"x x x", true},
		{"х х х", true},
		{"**", false},
		{"x x", false},
		{"abc", false},
	}
	for _, tc := range cases {
		if got := isSceneBreak(tc.in); got != tc.want {
			t.Fatalf("isSceneBreak(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseSplitsIndentedParagraphs(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"     Первый сектор молчал, но индикатор мигал.\n" +
		"на столе лежал чертеж и карандаш.\n" +
		"     Второй сектор ответил позже.\n" +
		"оператор сделал заметку."

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.Paragraph); !ok {
		t.Fatalf("expected first block Paragraph, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected second block Paragraph, got %T", doc.Blocks[1])
	}
}

func TestParseMixedLayoutBlocks(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"# Журнал  ночной смены\n" +
		"\n" +
		"     В мастерской  пахло озоном, и на табло мигало \"режим 3\"...\n" +
		"техник сверял 1990-2000 на шкале, а рядом лежал список[черновик ].\n" +
		"     Дежурный отметил: \"пуск - через минуту\", но таймер   молчал.\n" +
		"\n" +
		"-  Кто-то видел ключ?\n" +
		"—  Нет , но я слышал: \"щелк\" - и тишина...\n" +
		"- Тогда  проверим  второй  шкаф.\n" +
		"\n" +
		"* * *\n" +
		"\n" +
		"     После проверки команда  вернулась в зал, и кто-то сказал :\n" +
		"\"Все  в порядке !\""

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 6 {
		t.Fatalf("expected 6 blocks, got %d", len(doc.Blocks))
	}

	if _, ok := doc.Blocks[0].(ast.Heading); !ok {
		t.Fatalf("expected block 0 Heading, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.Paragraph); !ok {
		t.Fatalf("expected block 2 Paragraph, got %T", doc.Blocks[2])
	}
	if _, ok := doc.Blocks[3].(ast.DialogueBlock); !ok {
		t.Fatalf("expected block 3 DialogueBlock, got %T", doc.Blocks[3])
	}
	if _, ok := doc.Blocks[4].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 4 SceneBreak, got %T", doc.Blocks[4])
	}
	if _, ok := doc.Blocks[5].(ast.Paragraph); !ok {
		t.Fatalf("expected block 5 Paragraph, got %T", doc.Blocks[5])
	}
}

func TestParseBookLayoutWithoutBlankLines(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"\x01СОДЕРЖАНИЕ\x02\n" +
		"Глава I Пролог\n" +
		"Глава II Развязка\n" +
		"\n" +
		"Первая строка абзаца -----------------------------------------\n" +
		"Вторая строка абзаца.\n" +
		"x x x\n" +
		"После разделителя."

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 6 {
		t.Fatalf("expected 6 blocks, got %d", len(doc.Blocks))
	}

	cb, ok := doc.Blocks[0].(ast.ContentsBlock)
	if !ok {
		t.Fatalf("expected block 0 ContentsBlock, got %T", doc.Blocks[0])
	}
	if len(cb.Entries) != 2 {
		t.Fatalf("expected 2 contents entries, got %d", len(cb.Entries))
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 2 SceneBreak, got %T", doc.Blocks[2])
	}
	if _, ok := doc.Blocks[3].(ast.Paragraph); !ok {
		t.Fatalf("expected block 3 Paragraph, got %T", doc.Blocks[3])
	}
	if _, ok := doc.Blocks[4].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 4 SceneBreak, got %T", doc.Blocks[4])
	}
	if _, ok := doc.Blocks[5].(ast.Paragraph); !ok {
		t.Fatalf("expected block 5 Paragraph, got %T", doc.Blocks[5])
	}
}

func TestParseStandaloneMetaLines(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"OCR: Василий Новиков\n" +
		"Spellcheck: Евгений Морозов -----------------------------------------"

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}
	meta0, ok := doc.Blocks[0].(ast.MetaLineBlock)
	if !ok {
		t.Fatalf("expected block 0 MetaLineBlock, got %T", doc.Blocks[0])
	}
	if meta0.Key != "OCR" {
		t.Fatalf("expected first meta key OCR, got %q", meta0.Key)
	}
	meta1, ok := doc.Blocks[1].(ast.MetaLineBlock)
	if !ok {
		t.Fatalf("expected block 1 MetaLineBlock, got %T", doc.Blocks[1])
	}
	if meta1.Key != "Spellcheck" {
		t.Fatalf("expected second meta key Spellcheck, got %q", meta1.Key)
	}
	if _, ok := doc.Blocks[2].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 2 SceneBreak, got %T", doc.Blocks[2])
	}
}

func TestParseLeadingTitleBeforeMetaAndHeadings(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"\x14Ирина Ветрова. Сигнал с ледяной орбиты\x15\n" +
		"---------------------------------------------------------------\n" +
		"OCR: Алексей Протасов\n" +
		"Spellcheck: Марина Логинова ---------------------------------------------------------------\n" +
		"\n" +
		"\x14 * ЧАСТЬ ПЕРВАЯ * \x15\n"

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 6 {
		t.Fatalf("expected 6 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.TitleBlock); !ok {
		t.Fatalf("expected block 0 TitleBlock, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 1 SceneBreak, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.MetaLineBlock); !ok {
		t.Fatalf("expected block 2 MetaLineBlock, got %T", doc.Blocks[2])
	}
	if _, ok := doc.Blocks[3].(ast.MetaLineBlock); !ok {
		t.Fatalf("expected block 3 MetaLineBlock, got %T", doc.Blocks[3])
	}
	if _, ok := doc.Blocks[4].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 4 SceneBreak, got %T", doc.Blocks[4])
	}
	if _, ok := doc.Blocks[5].(ast.Heading); !ok {
		t.Fatalf("expected block 5 Heading, got %T", doc.Blocks[5])
	}
}

func TestParseLeadingTitleWithSceneBreakThenBody(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := strings.Join([]string{
		"Ирина Ветрова. Сигнал с ледяной орбиты",
		"---------------------------------------------------------------",
		"",
		"Первый абзац основного текста.",
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.TitleBlock); !ok {
		t.Fatalf("expected block 0 TitleBlock, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 1 SceneBreak, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.Paragraph); !ok {
		t.Fatalf("expected block 2 Paragraph, got %T", doc.Blocks[2])
	}
}

func TestParseLeadingTitleTwoLinesSameCandidate(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := strings.Join([]string{
		"АВТОР",
		"НАЗВАНИЕ КНИГИ",
		"",
		"OCR: Test",
		"# ЧАСТЬ ПЕРВАЯ",
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.TitleBlock); !ok {
		t.Fatalf("expected block 0 TitleBlock, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.MetaLineBlock); !ok {
		t.Fatalf("expected block 1 MetaLineBlock, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.Heading); !ok {
		t.Fatalf("expected block 2 Heading, got %T", doc.Blocks[2])
	}
}

func TestParseLeadingTitleTwoLinesSeparateCandidates(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := strings.Join([]string{
		"НАЗВАНИЕ КНИГИ",
		"",
		"АВТОР",
		"",
		"Spellcheck: Test",
		"### Пролог",
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.TitleBlock); !ok {
		t.Fatalf("expected block 0 TitleBlock, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.MetaLineBlock); !ok {
		t.Fatalf("expected block 1 MetaLineBlock, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.Heading); !ok {
		t.Fatalf("expected block 2 Heading, got %T", doc.Blocks[2])
	}
}

func TestParseLongIndentedLinesSplitIntoParagraphs(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	long1 := " " + strings.Repeat("Павлик думал о команде и о возвращении домой, ", 4) + "сердце сжималось."
	long2 := " " + strings.Repeat("Он шел по коридору и видел усталые лица товарищей, ", 4) + "и молчал."
	long3 := " " + strings.Repeat("После смены он долго смотрел в иллюминатор на темную воду, ", 4) + "не находя слов."

	input := strings.Join([]string{
		long1,
		long2,
		" - Марат, ты слышишь меня?",
		" - Слышу, Павлик.",
		long3,
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.Paragraph); !ok {
		t.Fatalf("expected block 0 Paragraph, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}
	db, ok := doc.Blocks[2].(ast.DialogueBlock)
	if !ok {
		t.Fatalf("expected block 2 DialogueBlock, got %T", doc.Blocks[2])
	}
	if len(db.Turns) != 2 {
		t.Fatalf("expected 2 dialogue turns, got %d", len(db.Turns))
	}
	if _, ok := doc.Blocks[3].(ast.Paragraph); !ok {
		t.Fatalf("expected block 3 Paragraph, got %T", doc.Blocks[3])
	}
}

func TestParseDialogueBlocksSplitByNarration(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"— Первая реплика.\n" +
		"— Вторая реплика.\n" +
		"Связующий авторский абзац между репликами.\n" +
		"— Третья реплика.\n" +
		"— Четвертая реплика."

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}

	first, ok := doc.Blocks[0].(ast.DialogueBlock)
	if !ok {
		t.Fatalf("expected block 0 DialogueBlock, got %T", doc.Blocks[0])
	}
	if len(first.Turns) != 2 {
		t.Fatalf("expected first dialogue block with 2 turns, got %d", len(first.Turns))
	}

	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}

	second, ok := doc.Blocks[2].(ast.DialogueBlock)
	if !ok {
		t.Fatalf("expected block 2 DialogueBlock, got %T", doc.Blocks[2])
	}
	if len(second.Turns) != 2 {
		t.Fatalf("expected second dialogue block with 2 turns, got %d", len(second.Turns))
	}
}

func TestParseStructuralLinesWithControlArtifacts(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := "" +
		"\x03СОДЕРЖАНИЕ\x7f\n" +
		"\x03Глава I. Вводная часть\x7f\n" +
		"\x03x x x\x7f\n" +
		"Текст после разделителя."

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}

	cb, ok := doc.Blocks[0].(ast.ContentsBlock)
	if !ok {
		t.Fatalf("expected block 0 ContentsBlock, got %T", doc.Blocks[0])
	}
	if len(cb.Entries) != 1 {
		t.Fatalf("expected 1 contents entry, got %d", len(cb.Entries))
	}
	if _, ok := doc.Blocks[1].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 1 SceneBreak, got %T", doc.Blocks[1])
	}
	if _, ok := doc.Blocks[2].(ast.Paragraph); !ok {
		t.Fatalf("expected block 2 Paragraph, got %T", doc.Blocks[2])
	}
}

func TestParseLongChapterLikeChunkDoesNotCollapseIntoSingleParagraph(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	firstLead := " " + strings.Repeat("Смена тянулась медленно, приборы мерцали в полутьме, ", 4) + "дежурный молчал."
	firstTail := "а в журнале копились короткие заметки о режиме и давлении."
	secondLead := " " + strings.Repeat("Под утро команда слышала только шаги в коридоре и звук насосов, ", 4) + "никто не спорил."
	secondTail := "и каждый думал о том, как завершить рейс без потерь."
	thirdLead := " " + strings.Repeat("Когда тревога утихла, механик снова проверил схему и кивнул, ", 4) + "работа продолжалась."
	thirdTail := "но в голосах оставалась усталость после длинного дня."

	input := strings.Join([]string{
		firstLead,
		firstTail,
		secondLead,
		secondTail,
		"— Доклад готов?",
		"— Готов, передаю в центральный пост.",
		thirdLead,
		thirdTail,
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(ast.Paragraph); !ok {
		t.Fatalf("expected block 0 Paragraph, got %T", doc.Blocks[0])
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}
	db, ok := doc.Blocks[2].(ast.DialogueBlock)
	if !ok {
		t.Fatalf("expected block 2 DialogueBlock, got %T", doc.Blocks[2])
	}
	if len(db.Turns) != 2 {
		t.Fatalf("expected 2 dialogue turns, got %d", len(db.Turns))
	}
	if _, ok := doc.Blocks[3].(ast.Paragraph); !ok {
		t.Fatalf("expected block 3 Paragraph, got %T", doc.Blocks[3])
	}
}

func TestParseEnglishBookLikeLayout(t *testing.T) {
	cfg, err := config.New("en", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := strings.Join([]string{
		"CONTENTS",
		"Chapter I. Interrupted Conversation",
		"",
		"    Before dawn there was almost no time left,",
		"and only the pumps could be heard in the corridor.",
		"",
		"- Are we ready?",
		"- Yes, almost.",
		"",
		"x x x",
		"",
		"After the break the team returned to work.",
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(doc.Blocks))
	}
	cb, ok := doc.Blocks[0].(ast.ContentsBlock)
	if !ok {
		t.Fatalf("expected block 0 ContentsBlock, got %T", doc.Blocks[0])
	}
	if len(cb.Entries) != 1 {
		t.Fatalf("expected 1 contents entry, got %d", len(cb.Entries))
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}
	db, ok := doc.Blocks[2].(ast.DialogueBlock)
	if !ok {
		t.Fatalf("expected block 2 DialogueBlock, got %T", doc.Blocks[2])
	}
	if len(db.Turns) != 2 {
		t.Fatalf("expected 2 dialogue turns, got %d", len(db.Turns))
	}
	if _, ok := doc.Blocks[3].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 3 SceneBreak, got %T", doc.Blocks[3])
	}
	if _, ok := doc.Blocks[4].(ast.Paragraph); !ok {
		t.Fatalf("expected block 4 Paragraph, got %T", doc.Blocks[4])
	}
}

func TestParseUkrainianBookLikeLayout(t *testing.T) {
	cfg, err := config.New("ua", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	input := strings.Join([]string{
		"ЗМІСТ",
		"Розділ I. Перерваний діалог",
		"",
		"    До світанку лишалося вже зовсім трохи,",
		"і в коридорі чулося лише гудіння насосів.",
		"",
		"- Ти готовий?",
		"- Так, починаємо.",
		"",
		"х х х",
		"",
		"Після перерви команда повернулася до роботи.",
	}, "\n")

	doc := Parse(input, cfg)
	if len(doc.Blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(doc.Blocks))
	}
	cb, ok := doc.Blocks[0].(ast.ContentsBlock)
	if !ok {
		t.Fatalf("expected block 0 ContentsBlock, got %T", doc.Blocks[0])
	}
	if len(cb.Entries) != 1 {
		t.Fatalf("expected 1 contents entry, got %d", len(cb.Entries))
	}
	if _, ok := doc.Blocks[1].(ast.Paragraph); !ok {
		t.Fatalf("expected block 1 Paragraph, got %T", doc.Blocks[1])
	}
	db, ok := doc.Blocks[2].(ast.DialogueBlock)
	if !ok {
		t.Fatalf("expected block 2 DialogueBlock, got %T", doc.Blocks[2])
	}
	if len(db.Turns) != 2 {
		t.Fatalf("expected 2 dialogue turns, got %d", len(db.Turns))
	}
	if _, ok := doc.Blocks[3].(ast.SceneBreak); !ok {
		t.Fatalf("expected block 3 SceneBreak, got %T", doc.Blocks[3])
	}
	if _, ok := doc.Blocks[4].(ast.Paragraph); !ok {
		t.Fatalf("expected block 4 Paragraph, got %T", doc.Blocks[4])
	}
}
