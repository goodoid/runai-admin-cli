package ui

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"strconv"
)

// DECLERATION: currently sub grouping are not supported

const ( 

	groupSeperation =  "┌⦿ "
	titleSeperation =  "├⚬ "
	rowsSeperation =   "│  "

	// tag names
	titleTagName  = "title"
	defultTagName = "def"
	formatTagName = "format"
	groupTagName  = "group"

	// group flags
	flattenGroupFlag	= "flatten"

)

var (
	groupId = 0
)

type (
	FormatFunction = func(value interface{}, model interface{}) (string, error)
	FormatterMap = map[string]FormatFunction

	Column struct {
		Formmater                 FormatFunction
		Path 					  []string
		Key, GroupID, Title, Defult string
	}

	tableData struct {
		columns   	[]Column
		modelType 	reflect.Type
		groups 		[]GroupTag
		opt 	  	TableOpt
		err       	error
	}

	Table interface {
		Render(w io.Writer, rows interface{}) Table
		RenderHeader(w io.Writer) Table
		RenderRows(w io.Writer, rows interface{}) Table
		Error() error
	}
)


func NewTag(tag string) Tag {
	t :=  Tag {
		Flags: map[string]bool{},
		Keys: map[string]string{},
	}
	tag = strings.TrimSpace(tag)
	tagSegments := strings.Split(tag, ",")
	for i,s := range tagSegments {
		if (i == 0) {
			t.Val = s
			continue
		}
		sub := strings.SplitN(s, "=", 1)
		// check if it is a feature or a flag
		if len(sub) == 2 {
			t.Keys[sub[0]] = sub[1]
		} else {
			t.Flags[sub[0]] = true
		}
	}

	return t;

}

// Tag is a general tag structure
type Tag struct {
	Val string // the first value 
	Flags map[string]bool
	Keys map[string]string
}

type GroupTag struct {
	Name string
	// to diff groups with the same name
	ID string
	// flags
	Flatten bool
	// keys
	Prefix string
}

type TableOpt struct {
	// set the default for the root struct (any root fields will be hidden by default if is true)
	HideAllByDefault bool
	// which field paths to show
	Show         []string
	// which field paths to hide
	Hide         []string
	// map format name into a function 
	Formatts FormatterMap
}

func CreateTable(model interface{}, opt TableOpt) Table {
	columns := []Column{}

	td := tableData {
		columns: columns,
		modelType: reflect.TypeOf(model),
		opt: opt,
	}

	isShowAllByDefault := true

	if opt.HideAllByDefault {
		isShowAllByDefault = false
	} else if opt.Show != nil {
		// if there is at least one filed on the root of the struct
		for _, path := range opt.Show {
			if !strings.Contains(path, ".") {
				isShowAllByDefault = false
				break
			}
		}
	}

	defaultGroup := NewGroupTag("")
	td.groups = []GroupTag{defaultGroup}

	td.addFields(td.modelType, []string{},defaultGroup , isShowAllByDefault)

	return &td
}

func (td *tableData) addFields(modelType reflect.Type, path []string, groupTag GroupTag, showByDefult bool) {
	fieldsCount := modelType.NumField()
	for i := 0; i < fieldsCount; i++ {
		td.addField( modelType.Field(i), path, groupTag, showByDefult)
	}
}

func (td *tableData) addField(fieldType reflect.StructField, path []string, groupTag GroupTag, showByDefult bool) {
	// if need to hide the field
	pathStr := strings.Join(append(path, fieldType.Name), ".")
	if td.opt.Hide != nil {
		if contains(td.opt.Hide, pathStr) {
			showByDefult = false
		}
	} 
	if td.opt.Show != nil {
		if contains(td.opt.Show, pathStr) {
			showByDefult = true
		}
	}
	if (isStructGroup(fieldType)) {
		td.addGroup(fieldType, path, groupTag, showByDefult)
		return
	} 
	if !showByDefult {
		return
	}
	column, err := toColumn(fieldType, td.opt.Formatts, path, groupTag)
	if err != nil {
		td.err = err
		return
	}
	td.columns = append(td.columns, column)
}

func (td *tableData) addGroup(field reflect.StructField, path []string, groupTag GroupTag, showByDefult bool) {
	groupTag = NewGroupTag(field.Tag.Get(groupTagName))
	groupPath := append(path, field.Name)
	td.groups=  append(td.groups, groupTag)

	td.addFields(UnwrapTypePtr(field.Type), groupPath, groupTag, showByDefult)
}

func (td *tableData) Render(w io.Writer, rows interface{}) Table {
	return td.RenderHeader(w).RenderRows(w, rows)
}

func (td *tableData) RenderHeader(w io.Writer) Table {
	if td.err != nil {
		return td
	}

	// add the groups
	if len(td.groups) > 1 {
		groupsCount := map[string]int{};
		groups := []string{}
		for _, c := range td.columns {
			groupsCount[c.GroupID] = groupsCount[c.GroupID] + 1
		}
		for i, tag := range td.groups {
			groupName := tag.Name
			if tag.Flatten {
				groupName = ""
			}
			spaces := groupsCount[tag.ID]
			if spaces == 0 {
				continue
			}
			tabs := make([]string, spaces)
			for i := range tabs{
				tabs[i]="\t"
			}
			if i > 0 && !tag.Flatten {
				groupName = groupSeperation + groupName
			}
			groups = append(groups, groupName + strings.Join(tabs, ""))
			i++
		}
		if len(groups) > 0 {
			fmt.Fprintln(w, strings.Join(groups, ""))
		}
	}

	titles := make([]string, len(td.columns))
	titlesBottomBorder := make([]string, len(td.columns))

	previousGroup := "1"
	for i, c := range td.columns {
		title := c.Title
		border := multiStr("─", len(title))
		if i > 0 && previousGroup != c.GroupID {
			title = titleSeperation + title
			border = rowsSeperation + border
		}
		previousGroup = c.GroupID
		titles[i] = title
		titlesBottomBorder[i] = border
	}

	fmt.Fprintln(w, strings.Join(titles, "\t"))
	fmt.Fprintln(w, strings.Join(titlesBottomBorder, "\t"))

	return td
}

func (td *tableData) RenderRows(w io.Writer, rows interface{}) Table {
	if td.err != nil {
		return td
	}
	var err error
	data, err := interfaceToArrayOfInterface(rows)
	if err != nil {
		td.err = err
		return td
	}

	values := make([]string, len(td.columns))
	for _, r := range data {
		t := reflect.ValueOf(r)
		previousGroup := ""
		for i, c := range td.columns {
			var val string

			ftp := getNesstedVal(t, append(c.Path, c.Key))
			// if the value is not nil
			if ftp != nil {
				ft := *ftp
				if c.Formmater != nil {
					val, err = c.Formmater(ft.Interface(), r)
					if err != nil {
						td.err = err
						return td
					}
				} else {
					val = StringifyValue(ft)
				}
			}

			// set default value if it is an empty
			if len(val) == 0 {
				val = c.Defult
			}

			if i > 0 && previousGroup != c.GroupID {
				val = rowsSeperation + val
			}
			previousGroup = c.GroupID

			values[i] = val
		}

		buffer := strings.Join(values, "\t")
		fmt.Fprintln(w, buffer)

	}
	return td
}

func (td *tableData) Error() error {
	return td.err
}

//// helpers

func isStructGroup(field reflect.StructField) bool {
	isStruct := UnwrapTypePtr(field.Type).Kind() == reflect.Struct;

	group := field.Tag.Get(groupTagName)
	format := field.Tag.Get(formatTagName)

	return isStruct && len(group) > 0 && len(format) == 0
}

func NewGroupTag(tagStr string) GroupTag{
	tag := NewTag(tagStr)
	groupId+=1
	return GroupTag {
		ID: strconv.Itoa(groupId),
		Name: tag.Val,
		Flatten: tag.Flags[flattenGroupFlag] || len(tag.Val)==0,
	}
}

func toColumn(field reflect.StructField, formatMap FormatterMap, path []string, groupTag GroupTag) (Column, error) {
	var formaterFunc FormatFunction
	key := field.Name
	title := field.Tag.Get(titleTagName)
	def := field.Tag.Get(defultTagName)
	format := field.Tag.Get(formatTagName)

	if len(format) != 0 {
		f, found := formatMap[format]
		// if not found search in the default format
		if !found {
			f, found =DefaultTableFormat[format]
		}

		if !found {
			return Column{}, fmt.Errorf("[Table] Not found format function for format name: %s  on field name: %s . Please make sure to include it in the TableOpt.CustomFormat", format, key)
		}
		formaterFunc = f
	}

	if len(title) == 0 {
		title = key
	}

	return Column{
		Title:     title,
		Defult:    def,
		GroupID:   groupTag.ID,
		Key:       key,
		Path:      path,
		Formmater: formaterFunc,
	}, nil
}