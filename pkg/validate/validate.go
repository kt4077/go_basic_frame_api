// Package validate 提供请求绑定、参数校验和面向用户的校验错误转换能力。
package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	playground "github.com/go-playground/validator/v10"
)

// Bind 根据请求方法绑定 JSON 请求体或查询参数，并返回使用 validate tag 描述的错误。
func Bind(c *gin.Context, target interface{}) error {
	var err error
	if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
		err = c.ShouldBindQuery(target)
	} else {
		err = c.ShouldBindJSON(target)
	}
	return translateError(err, target)
}

// BindURI 绑定路径参数，并返回使用 validate tag 描述的错误。
func BindURI(c *gin.Context, target interface{}) error {
	return translateError(c.ShouldBindUri(target), target)
}

// Struct 对已经赋值的请求结构执行 binding 规则校验。
func Struct(target interface{}) error {
	engine, err := engine()
	if err != nil {
		return err
	}
	return translateError(engine.Struct(target), target)
}

// RegisterValidation 注册字段级自定义 binding 校验规则。应在服务启动、处理请求前调用。
func RegisterValidation(tag string, fn playground.Func) error {
	engine, err := engine()
	if err != nil {
		return err
	}
	return engine.RegisterValidation(tag, fn)
}

// RegisterStructValidation 注册结构体级自定义校验规则。应在服务启动、处理请求前调用。
func RegisterStructValidation(fn playground.StructLevelFunc, types ...interface{}) error {
	engine, err := engine()
	if err != nil {
		return err
	}
	engine.RegisterStructValidation(fn, types...)
	return nil
}

func engine() (*playground.Validate, error) {
	validator, ok := binding.Validator.Engine().(*playground.Validate)
	if !ok {
		return nil, errors.New("参数验证器初始化失败")
	}
	return validator, nil
}

func translateError(err error, target interface{}) error {
	if err == nil {
		return nil
	}
	var validationErrors playground.ValidationErrors
	if errors.As(err, &validationErrors) && len(validationErrors) > 0 {
		fieldError := validationErrors[0]
		return errors.New(validationMessage(fieldSemantic(target, fieldError.StructNamespace()), fieldError))
	}
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		return fmt.Errorf("%s格式不正确", fieldSemantic(target, typeError.Field))
	}
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return errors.New("请求参数格式错误")
	}
	return errors.New("请求参数格式错误")
}

func fieldSemantic(target interface{}, namespace string) string {
	typeOf := reflect.TypeOf(target)
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	parts := strings.Split(namespace, ".")
	if len(parts) > 0 && parts[0] == typeOf.Name() {
		parts = parts[1:]
	}
	for _, part := range parts {
		part = strings.Split(part, "[")[0]
		if typeOf.Kind() != reflect.Struct {
			break
		}
		field, ok := findField(typeOf, part)
		if !ok {
			break
		}
		if semantic := strings.TrimSpace(field.Tag.Get("validate")); semantic != "" {
			return semantic
		}
		if semantic := semanticFromComment(field.Tag.Get("comment")); semantic != "" {
			return semantic
		}
		typeOf = field.Type
		for typeOf.Kind() == reflect.Ptr || typeOf.Kind() == reflect.Slice || typeOf.Kind() == reflect.Array {
			typeOf = typeOf.Elem()
		}
	}
	return "请求参数"
}

func findField(typeOf reflect.Type, name string) (reflect.StructField, bool) {
	if field, ok := typeOf.FieldByName(name); ok {
		return field, true
	}
	for index := 0; index < typeOf.NumField(); index++ {
		field := typeOf.Field(index)
		for _, tagName := range []string{"json", "form", "uri"} {
			tagValue := strings.Split(field.Tag.Get(tagName), ",")[0]
			if tagValue == name {
				return field, true
			}
		}
	}
	return reflect.StructField{}, false
}

func semanticFromComment(comment string) string {
	comment = strings.TrimSpace(comment)
	if index := strings.IndexAny(comment, "：，；（("); index >= 0 {
		comment = comment[:index]
	}
	return strings.TrimSpace(comment)
}

func validationMessage(semantic string, fieldError playground.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return semantic + "不能为空"
	case "oneof":
		return semantic + "取值不正确"
	case "email":
		return semantic + "格式不正确"
	case "url":
		return semantic + "必须是有效的URL地址"
	case "alphanum":
		return semantic + "只能包含字母和数字"
	case "len":
		return semantic + "长度必须为" + fieldError.Param()
	case "min":
		if fieldError.Kind() == reflect.String {
			return semantic + "长度不能少于" + fieldError.Param() + "个字符"
		}
		if isCollection(fieldError.Kind()) {
			return semantic + "至少包含" + fieldError.Param() + "项"
		}
		return semantic + "不能小于" + fieldError.Param()
	case "max":
		if fieldError.Kind() == reflect.String {
			return semantic + "长度不能超过" + fieldError.Param() + "个字符"
		}
		if isCollection(fieldError.Kind()) {
			return semantic + "最多包含" + fieldError.Param() + "项"
		}
		return semantic + "不能超过" + fieldError.Param()
	case "gt":
		return semantic + "必须大于" + fieldError.Param()
	case "gte":
		return semantic + "不能小于" + fieldError.Param()
	case "lt":
		return semantic + "必须小于" + fieldError.Param()
	case "lte":
		return semantic + "不能大于" + fieldError.Param()
	case "eq":
		return semantic + "必须为" + fieldError.Param()
	default:
		if fieldError.Param() != "" {
			return semantic + "不符合" + fieldError.Tag() + "规则（" + strconv.Quote(fieldError.Param()) + "）"
		}
		return semantic + "不符合要求"
	}
}

func isCollection(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map
}
