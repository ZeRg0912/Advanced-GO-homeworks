package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func Validate(v interface{}) error {
	value := reflect.ValueOf(v)
	valueType := reflect.TypeOf(v)

	if !value.IsValid() {
		return fmt.Errorf("validation error: invalid value")
	}

	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return fmt.Errorf("validation error: nil pointer")
		}

		value = value.Elem()
		valueType = valueType.Elem()
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("validation error: expected struct, got %s", value.Kind())
	}

	for i := 0; i < value.NumField(); i++ {
		structField := valueType.Field(i)
		fieldValue := value.Field(i)

		validateTag := structField.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rules := strings.Split(validateTag, ";")

		for _, rule := range rules {
			if rule == "" {
				continue
			}

			parts := strings.SplitN(rule, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("validation error: field %s has invalid rule %q", structField.Name, rule)
			}

			ruleName := parts[0]
			ruleValue := parts[1]

			switch ruleName {
			case "min":
				if err := validateMin(structField.Name, fieldValue, ruleValue); err != nil {
					return err
				}

			case "max":
				if err := validateMax(structField.Name, fieldValue, ruleValue); err != nil {
					return err
				}

			case "regexp":
				if err := validateRegexp(structField.Name, fieldValue, ruleValue); err != nil {
					return err
				}

			default:
				return fmt.Errorf("validation error: field %s has unsupported rule %q", structField.Name, ruleName)
			}
		}
	}

	return nil
}

func validateMin(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		minLength, err := strconv.Atoi(ruleValue)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid min value %q", fieldName, ruleValue)
		}

		actualLength := len([]rune(fieldValue.String()))
		if actualLength < minLength {
			return fmt.Errorf(
				"validation error: field %s length must be at least %d, got %d",
				fieldName,
				minLength,
				actualLength,
			)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		minValue, err := strconv.ParseInt(ruleValue, 10, 64)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid min value %q", fieldName, ruleValue)
		}

		actualValue := fieldValue.Int()
		if actualValue < minValue {
			return fmt.Errorf(
				"validation error: field %s must be at least %d, got %d",
				fieldName,
				minValue,
				actualValue,
			)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		minValue, err := strconv.ParseUint(ruleValue, 10, 64)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid min value %q", fieldName, ruleValue)
		}

		actualValue := fieldValue.Uint()
		if actualValue < minValue {
			return fmt.Errorf(
				"validation error: field %s must be at least %d, got %d",
				fieldName,
				minValue,
				actualValue,
			)
		}

	case reflect.Float32, reflect.Float64:
		minValue, err := strconv.ParseFloat(ruleValue, 64)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid min value %q", fieldName, ruleValue)
		}

		actualValue := fieldValue.Float()
		if actualValue < minValue {
			return fmt.Errorf(
				"validation error: field %s must be at least %g, got %g",
				fieldName,
				minValue,
				actualValue,
			)
		}

	default:
		return fmt.Errorf("validation error: field %s does not support min rule", fieldName)
	}

	return nil
}

func validateMax(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		maxLength, err := strconv.Atoi(ruleValue)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid max value %q", fieldName, ruleValue)
		}

		actualLength := len([]rune(fieldValue.String()))
		if actualLength > maxLength {
			return fmt.Errorf(
				"validation error: field %s length must be at most %d, got %d",
				fieldName,
				maxLength,
				actualLength,
			)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		maxValue, err := strconv.ParseInt(ruleValue, 10, 64)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid max value %q", fieldName, ruleValue)
		}

		actualValue := fieldValue.Int()
		if actualValue > maxValue {
			return fmt.Errorf(
				"validation error: field %s must be at most %d, got %d",
				fieldName,
				maxValue,
				actualValue,
			)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		maxValue, err := strconv.ParseUint(ruleValue, 10, 64)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid max value %q", fieldName, ruleValue)
		}

		actualValue := fieldValue.Uint()
		if actualValue > maxValue {
			return fmt.Errorf(
				"validation error: field %s must be at most %d, got %d",
				fieldName,
				maxValue,
				actualValue,
			)
		}

	case reflect.Float32, reflect.Float64:
		maxValue, err := strconv.ParseFloat(ruleValue, 64)
		if err != nil {
			return fmt.Errorf("validation error: field %s has invalid max value %q", fieldName, ruleValue)
		}

		actualValue := fieldValue.Float()
		if actualValue > maxValue {
			return fmt.Errorf(
				"validation error: field %s must be at most %g, got %g",
				fieldName,
				maxValue,
				actualValue,
			)
		}

	default:
		return fmt.Errorf("validation error: field %s does not support max rule", fieldName)
	}

	return nil
}

func validateRegexp(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	if fieldValue.Kind() != reflect.String {
		return fmt.Errorf("validation error: field %s does not support regexp rule", fieldName)
	}

	re, err := regexp.Compile(ruleValue)
	if err != nil {
		return fmt.Errorf("validation error: field %s has invalid regexp %q: %w", fieldName, ruleValue, err)
	}

	actualValue := fieldValue.String()
	if !re.MatchString(actualValue) {
		return fmt.Errorf(
			"validation error: field %s does not match regexp %q",
			fieldName,
			ruleValue,
		)
	}

	return nil
}

type User struct {
	Name  string `validate:"min=3"`
	Age   int    `validate:"min=18;max=65"`
	Email string `validate:"regexp=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
}

func main() {
	if err := Validate(User{
		Name:  "Ив",
		Age:   18,
		Email: "test@example.com",
	}); err != nil {
		fmt.Println("Validation error:", err)
	}

	if err := Validate(User{
		Name:  "Иван",
		Age:   70,
		Email: "test@example.com",
	}); err != nil {
		fmt.Println("Validation error:", err)
	}

	if err := Validate(User{
		Name:  "Иван",
		Age:   35,
		Email: "invalid email",
	}); err != nil {
		fmt.Println("Validation error:", err)
	}

	if err := Validate(User{
		Name:  "Иван",
		Age:   35,
		Email: "test@example.com",
	}); err != nil {
		fmt.Println("Validation error:", err)
	}
}
