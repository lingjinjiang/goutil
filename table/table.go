package table

import (
	"fmt"
	"log"
	"reflect"

	"github.com/lingjinjiang/goutil/common"
)

type Table struct {
	size    int
	columes map[string]*colume
}

type colume struct {
	Name string
	Type reflect.Type
	Data []reflect.Value
}

func NewTable[T any](objs []T) Table {
	o := objs[0]
	objType := reflect.TypeOf(o)
	headerIndex := make(map[int]string)
	ignoreIndex := make(map[int]bool)
	columes := make(map[string]*colume)
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		ignore := !field.IsExported()
		ignoreIndex[i] = ignore
		if ignore {
			continue
		}
		col := colume{
			Name: field.Name,
			Type: field.Type,
			Data: make([]reflect.Value, 0),
		}
		columes[col.Name] = &col
		headerIndex[i] = col.Name
	}

	for _, o := range objs {
		values := reflect.ValueOf(o)
		for i := 0; i < values.NumField(); i++ {
			if ignoreIndex[i] {
				continue
			}
			col := columes[headerIndex[i]]
			col.Data = append(col.Data, values.Field(i))
		}
	}

	return Table{columes: columes, size: len(objs)}
}

func (tab Table) ShowSchema() {
	data := make([][]string, 0)
	for _, col := range tab.columes {
		data = append(data, []string{col.Name, col.Type.Name()})
	}
	fmt.Println(common.BuildTableStr([]string{"name", "type"}, data))
}

func (tab Table) Show() {
	data := make([][]string, tab.size)
	header := make([]string, 0)
	for _, col := range tab.columes {
		header = append(header, col.Name)
		for i, d := range col.Data {
			if data[i] == nil {
				data[i] = make([]string, 0)
			}
			data[i] = append(data[i], fmt.Sprint(d))
		}
	}
	fmt.Println(common.BuildTableStr(header, data))
}

func (tab Table) Select(cols ...string) Table {
	columes := make(map[string]*colume)
	for _, cName := range cols {
		col := tab.columes[cName]
		if col != nil {
			columes[cName] = col
		}
	}
	return Table{columes: columes, size: tab.size}
}

func (tab Table) Where(c condition) Table {
	rows := findRow(c, tab.columes[c.getColume()])
	if rows == nil {
		return Table{}
	}
	columes := make(map[string]*colume)
	for i := range rows {
		for name, col := range tab.columes {
			newCol := columes[name]
			if newCol == nil {
				newCol = &colume{Name: col.Name, Type: col.Type, Data: make([]reflect.Value, 0)}
				columes[name] = newCol
			}
			newCol.Data = append(newCol.Data, col.Data[i])
		}
	}
	return Table{len(rows), columes}
}

func findRow(c condition, col *colume) []int {
	if col == nil {
		log.Printf("The table doesn't contain colume named '%s'", c.getColume())
		return nil
	}
	vType := reflect.TypeOf(c.getValue())
	if vType != col.Type {
		log.Printf("The value's type '%s' doesn't match colume's type '%s'", vType.Name(), col.Type.Name())
		return nil
	}
	rows := make([]int, 0)
	for i, value := range col.Data {
		if c.evaluate(value.Interface()) {
			rows = append(rows, i)
		}
	}
	return rows
}

func (tab Table) Unmarshal(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("should be a pointer slice")
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("should be a pointer slice")
	}
	newv := reflect.MakeSlice(elem.Type(), tab.size, tab.size)
	for i := 0; i < tab.size; i++ {
		n := newv.Index(i)
		for cName, cData := range tab.columes {
			tmp := n.FieldByName(cName)
			tmp.Set((*cData).Data[i])
		}
	}
	elem.Set(newv)
	return nil
}
