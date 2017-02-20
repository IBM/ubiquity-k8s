package terminal

import (
	"fmt"
	"io"
	"strings"
)

type Table interface {
	Add(row ...string)
	Print()
}

type PrintableTable struct {
	writer        io.Writer
	headers       []string
	headerPrinted bool
	maxSizes      []int
	rows          [][]string //each row is single line
}

func NewTable(w io.Writer, headers []string) Table {
	return &PrintableTable{
		writer:   w,
		headers:  headers,
		maxSizes: make([]int, len(headers)),
	}
}

func (t *PrintableTable) Add(row ...string) {
	var maxLines int

	var columns [][]string
	for _, value := range row {
		lines := strings.Split(value, "\n")
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
		columns = append(columns, lines)
	}

	for i := 0; i < maxLines; i++ {
		var row []string
		for _, col := range columns {
			if i >= len(col) {
				row = append(row, "")
			} else {
				row = append(row, col[i])
			}
		}
		t.rows = append(t.rows, row)
	}
}

func (t *PrintableTable) Print() {
	for _, row := range append(t.rows, t.headers) {
		t.calculateMaxSize(row)
	}

	if t.headerPrinted == false {
		t.printHeader()
		t.headerPrinted = true
	}

	for _, line := range t.rows {
		t.printRow(line)
	}

	t.rows = [][]string{}
}

func (t *PrintableTable) calculateMaxSize(row []string) {
	for index, value := range row {
		cellLength := calculateStringWidth(Decolorize(value))
		if t.maxSizes[index] < cellLength {
			t.maxSizes[index] = cellLength
		}
	}
}

func (t *PrintableTable) printHeader() {
	output := ""
	for col, value := range t.headers {
		output = output + t.cellValue(col, HeaderColor(value))
	}
	fmt.Fprintln(t.writer, output)
}

func (t *PrintableTable) printRow(row []string) {
	output := ""
	for columnIndex, value := range row {
		if columnIndex == 0 {
			value = TableContentHeaderColor(value)
		}

		output = output + t.cellValue(columnIndex, value)
	}
	fmt.Fprintln(t.writer, output)
}

func (t *PrintableTable) cellValue(col int, value string) string {
	padding := ""
	if col < len(t.headers)-1 {
		padding = strings.Repeat(" ", t.maxSizes[col]-calculateStringWidth(Decolorize(value)))
	}
	return fmt.Sprintf("%s%s   ", value, padding)
}

func calculateStringWidth(s string) int {
	r := strings.NewReader(s)

	var width int
	for range s {
		_, runeSize, err := r.ReadRune()
		if err != nil {
			panic(fmt.Sprintf("error when calculating visible size of: %s", s))
		}

		if runeSize == 3 {
			width += 2
		} else {
			width++
		}
	}

	return width
}
