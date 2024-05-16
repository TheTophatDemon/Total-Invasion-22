package main

import (
	"os"
	"strconv"
	"strings"
	"text/template"
)

const TEMPLATE = `info face="Total Invasion II Font" size=16 bold=0 italic=0 charset="" unicode=1 stretchH=100 smooth=0 aa=1 padding=0,0,0,0 spacing=2,2 outline=0
common lineHeight=24 base=19 scaleW=256 scaleH=288 pages=1 packed=0 alphaChnl=0 redChnl=4 greenChnl=4 blueChnl=4
page id=0 file="font.png"
chars count={{len .Chars}}
{{range .Chars -}}
char id={{.ID}} x={{.X}} y={{.Y}} width={{.Width}} height={{.Height}} xoffset=0 yoffset=0 xadvance={{.Width}} page=0 chnl=15
{{end}}
{{- /* Delete trailing whitespice or the parser will complain */ -}}
`

type Char struct {
	ID                  uint
	X, Y, Width, Height int
}

func main() {
	if len(os.Args) != 3 {
		panic("Missing arguments! Argument 1: Output file path, argument 2: texture width")
	}

	imgWidth, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		panic(err)
	}

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
	err = tpl.Execute(builder, struct {
		Chars []Char
	}{
		Chars: chars,
	})
	if err != nil {
		panic(err)
	}

	// Write
	file, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = file.WriteString(builder.String())
}
