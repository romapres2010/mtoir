package xlsx

import (
	"bytes"

	"github.com/xuri/excelize/v2"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

var gTitleStyle *excelize.Style
var gCellStyle *excelize.Style

func init() {

	gTitleStyle = &excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "centerContinuous",
			//Indent:          1,
			//JustifyLastLine: true,
			//ReadingOrder:    0,
			//RelativeIndent:  1,
			//ShrinkToFit:     true,
			//TextRotation:    45,
			Vertical: "center",
			//WrapText:        true,
		},
		Font: &excelize.Font{
			Bold:   true,
			Italic: false,
			//Family: "Times New Roman",
			Size: 12,
			//Color:  "#777777",
		},
	}

	gCellStyle = &excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			//Indent:          1,
			//JustifyLastLine: true,
			//ReadingOrder:    0,
			//RelativeIndent:  1,
			//ShrinkToFit:     true,
			//TextRotation:    45,
			Vertical: "center",
			//WrapText:        true,
		},
		Font: &excelize.Font{
			Bold:   false,
			Italic: false,
			//Family: "Times New Roman",
			Size: 12,
			//Color:  "#777777",
		},
	}
}

type WriteOption struct {
	SetTitles bool
	NewRow    bool
	//SheetName      string
	TitlePrefix    string
	GroupTitles    bool
	TitleGroupName string
	Transpose      bool // поменять столбцы и строки
	AutoFilter     bool
	CascadeStruct  bool // каскадно выводить все вложенные структуры
	FloatPrecision int  // точность округления данных
}

type Xlsx struct {
	reqID  uint64
	file   *excelize.File
	sheets map[string]*Sheet
}

func NewXlsx(reqID uint64, file *excelize.File) *Xlsx {
	var xlsx = Xlsx{reqID: reqID}

	xlsx.sheets = make(map[string]*Sheet)

	if file == nil {
		xlsx.file = excelize.NewFile()
	} else {
		xlsx.file = file
	}

	return &xlsx
}

func (xls *Xlsx) File() *excelize.File {
	return xls.file
}

func (xls *Xlsx) WriteToBuffer() (*bytes.Buffer, error) {
	if xls != nil && xls.file != nil {
		return xls.file.WriteToBuffer()
	} else {
		return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if xls != nil && xls.file != nil {}", []interface{}{xls}).PrintfError()
	}
}

func (xls *Xlsx) Close() error {
	if xls != nil && xls.file != nil {
		return xls.file.Close()
	} else {
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if xls != nil && xls.file != nil {}", []interface{}{xls}).PrintfError()
	}
}
