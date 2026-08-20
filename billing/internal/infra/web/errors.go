package web

import (
	"encoding/json"
	"errors"
	"fmt"
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
			messages[fieldKey(fieldError)] = validationMessage(fieldError)
		}

		return messages
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) && typeError.Field != "" {
		return map[string]string{typeError.Field: "Valor inválido."}
	}

	return nil
}

func fieldKey(fieldError validator.FieldError) string {
	namespace := fieldError.Namespace()

	if _, path, found := strings.Cut(namespace, "."); found {
		return path
	}

	return fieldError.Field()
}

func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "Campo obrigatório."
	case "gte":
		if fieldError.Param() == "0" {
			return "O valor não pode ser negativo."
		}

		return fmt.Sprintf("O valor não pode ser menor que %s.", fieldError.Param())
	case "gt":
		if fieldError.Param() == "0" {
			return "O valor precisa ser maior que zero."
		}

		return fmt.Sprintf("O valor precisa ser maior que %s.", fieldError.Param())
	case "max":
		return fmt.Sprintf("Limite de %s caracteres excedido.", fieldError.Param())
	default:
		return "Valor inválido."
	}
}
