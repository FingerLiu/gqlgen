package main

/* 解析 json， 提取类型并存入结构体
假定 json 文件最外层是大括号
先不处理有 list 的情况
文件名不支持中文
 */
import (
	"io/ioutil"
	"encoding/json"
		"reflect"
	"fmt"
	"unicode"
	"path/filepath"
	)

func unmarshal(filename string) (interface{}, error) {
	var jsonRaw []byte
	var result interface{}
	var err error
	jsonRaw, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonRaw, &result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

type Node struct {
	Name string
	ValueType reflect.Kind
	Children *[]Node
}

type GqlType struct {
	Name string
	Children *[] Node
}
/*
		"Int":          {Model: "github.com/99designs/gqlgen/graphql.Int"},
		"Float":        {Model: "github.com/99designs/gqlgen/graphql.Float"},
		"String":       {Model: "github.com/99designs/gqlgen/graphql.String"},
		"Boolean":      {Model: "github.com/99designs/gqlgen/graphql.Boolean"},
		"ID":           {Model: "github.com/99designs/gqlgen/graphql.ID"},
		"Time":         {Model: "github.com/99designs/gqlgen/graphql.Time"},
		"Map":          {Model: "github.com/99designs/gqlgen/graphql.Map"},
*/

func uppercaseFirst(s string) string {
	if s == "" {
		return ""
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func getRootType(s string) string {
	ext := filepath.Ext(s)
	return uppercaseFirst(s[0:len(s) - len(ext)])
}

func ensureAndAppend(ptr *[]Node, node Node) *[]Node  {
	if ptr == nil {
		arr := make([]Node, 0, 1000)
		ptr = &arr
	}
	*ptr = append(*ptr, node)
	return ptr
}

func Parse(obj interface{}, gqlTypesPtr *[]GqlType, gqlType GqlType, node Node) {
	for key, value := range obj.(map[string]interface{}) {
		//fmt.Printf("key %v \n value %v\n", key, value)

		var valueType reflect.Kind
		if value == nil {
			valueType = reflect.String
		} else {
			valueType = reflect.TypeOf(value).Kind()
		}

		child := Node{Name:key, ValueType:valueType}

		if value != nil && valueType == reflect.Map{

			childGqlType := GqlType{Name:uppercaseFirst(key)}
			Parse(value, gqlTypesPtr, childGqlType, child)


		} else if valueType == reflect.Slice {
			// TODO
			panic("unsupported value type: list is not supported yet.")
		}
		// 为父类型添加子类型
		gqlType.Children = ensureAndAppend(gqlType.Children, child)

		// 为父节点添加当前节点
		node.Children = ensureAndAppend(node.Children, child)
	}
	*gqlTypesPtr = append(*gqlTypesPtr, gqlType)
}

func main()  {
	filename := "platform.json"
	rootTypeName := getRootType(filename)
	rootObj, err := unmarshal(filename)

	if err != nil {
		panic(err)
	}
	fmt.Printf("rootObj is %v\ntype is %v\n", rootObj, reflect.TypeOf(rootObj))

	t := reflect.TypeOf(rootObj)

	mapType := t

	root := Node{Name:"root", ValueType:t.Kind()}
	rootType := GqlType{Name: rootTypeName+"Result"}
	gqlTypes := make([]GqlType, 0, 1000)
	gqlTypesPtr := &gqlTypes

	if t == mapType {
		Parse(rootObj, gqlTypesPtr, rootType, root)
		fmt.Printf("types %v \n", *(*gqlTypesPtr)[1].Children)
	} else {
		// TODO
		panic("unsupported json format: root object is not map.")
	}
}
