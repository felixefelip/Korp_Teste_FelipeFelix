package web

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	engine.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
}

func bindErrors(err error) map[string]string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		messages := make(map[string]string, len(validationErrors))

		for _, fieldError := range validationErrors {
			messages[fieldError.Field()] = validationMessage(fieldError)
		}

		return messages
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) && typeError.Field != "" {
		return map[string]string{typeError.Field: "tipo invalido"}
	}

	return nil
}

func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "obrigatorio"
	case "gte":
		return "nao pode ser menor que " + fieldError.Param()
	case "max":
		return "excede " + fieldError.Param() + " caracteres"
	case "oneof":
		return "precisa ser um de: " + strings.ReplaceAll(fieldError.Param(), " ", ", ")
	default:
		return "invalido"
	}
}
