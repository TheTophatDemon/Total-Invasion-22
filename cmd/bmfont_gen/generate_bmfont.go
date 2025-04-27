package main

import (
	_ "image/png"
	"os"
	"strings"
	"text/template"
)

const TEMPLATE = `info face="{{.FontName}}" size={{.FontSize}} bold=0 italic=0 charset="" unicode=1 stretchH=100 smooth=0 aa=1 padding=0,0,0,0 spacing={{index .Spacing 0}},{{index .Spacing 1}} outline=0
common lineHeight={{.LineHeight}} base={{.Base}} scaleW={{.ImageWidth}} scaleH={{.ImageHeight}} pages=1 packed=0 alphaChnl=0 redChnl=4 greenChnl=4 blueChnl=4
page id=0 file="{{.ImagePath}}"
chars count={{len .Chars}}
{{range .Chars -}}
char id={{.ID}} x={{.X}} y={{.Y}} width={{.Width}} height={{.Height}} xoffset=0 yoffset=0 xadvance={{.Width}} page=0 chnl=15
{{end}}
{{- /* Delete trailing whitespice or the parser will complain */ -}}
`

type TemplParams struct {
	FontName, ImagePath        string
	ImageWidth, ImageHeight    int
	LineHeight, Base, FontSize int
	Spacing                    [2]int
	Chars                      []Char
}

type Char struct {
	ID                  uint
	X, Y, Width, Height int
}

func main() {
	generateMainFont()
	generateHudFont()
}

func generateMainFont() {
	imgWidth := 256

	chars := make([]Char, 0)

	// Include all ASCII characters
	for i := range 95 {
		c := i + 32
		chars = append(chars, Char{
			ID:     uint(c),
			X:      ((i * 16) % int(imgWidth)),
			Y:      ((i * 16) / int(imgWidth)) * 24,
			Width:  16,
			Height: 24,
		})
	}

	// Include all Cyrillic characters
	cyrillicRunes := []rune{
		'А', 'Б', 'В', 'Г', 'Д', 'Е', 'Ё', 'Ж', 'З', 'И', 'Й', 'К', 'Л', 'М', 'Н', 'О', 'П', 'Р', 'С', 'Т', 'У', 'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ', 'Ъ', 'Ы', 'Ь', 'Э', 'Ю', 'Я',
		'а', 'б', 'в', 'г', 'д', 'е', 'ё', 'ж', 'з', 'и', 'й', 'к', 'л', 'м', 'н', 'о', 'п', 'р', 'с', 'т', 'у', 'ф', 'х', 'ц', 'ч', 'ш', 'щ', 'ъ', 'ы', 'ь', 'э', 'ю', 'я',
	}
	for i, r := range cyrillicRunes {
		chars = append(chars, Char{
			ID:     uint(r),
			X:      ((i * 16) % int(imgWidth)),
			Y:      144 + ((i*16)/int(imgWidth))*24,
			Width:  16,
			Height: 24,
		})
	}

	tpl := template.Must(template.New("BMFont").Parse(TEMPLATE))
	builder := &strings.Builder{}
	err := tpl.Execute(builder, TemplParams{
		FontName:   "Total Invasion 22 Font",
		ImagePath:  "font.png",
		ImageWidth: 256, ImageHeight: 288,
		LineHeight: 24, Base: 0, FontSize: 16,
		Spacing: [2]int{2, 2},
		Chars:   chars,
	})
	if err != nil {
		panic(err)
	}

	// Write
	file, err := os.Create("assets/textures/ui/font.fnt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = file.WriteString(builder.String())
}

func generateHudFont() {
	chars := make([]Char, 0)
	// Digits
	for i := range 10 {
		chars = append(chars, Char{
			ID:    uint('0' + i),
			X:     i*12 + 2,
			Y:     0,
			Width: 9, Height: 12,
		})
	}
	// Infinity
	chars = append(chars, Char{
		ID: uint('∞'),
		X:  0, Y: 12, Width: 24, Height: 12,
	})

	tpl := template.Must(template.New("BMFont").Parse(TEMPLATE))
	builder := &strings.Builder{}
	err := tpl.Execute(builder, TemplParams{
		FontName:   "Total Invasion 22 HUD Counter Font",
		ImagePath:  "hud_counter_font.png",
		ImageWidth: 120, ImageHeight: 24,
		LineHeight: 16, Base: 0, FontSize: 12,
		Spacing: [2]int{2, 1},
		Chars:   chars,
	})
	if err != nil {
		panic(err)
	}

	// Write
	file, err := os.Create("assets/textures/ui/hud_counter_font.fnt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = file.WriteString(builder.String())
}
