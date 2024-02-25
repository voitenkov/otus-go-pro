package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var (
	errorsWrapped        error
	validationErrors     ValidationErrors
	ErrNilValue          = errors.New("nil value in input")
	ErrStructureExpected = errors.New("structure kind expected in input")
	ErrEmptyStructure    = errors.New("empty structure")
	ErrInvalidValidator  = errors.New("invalid validator")
	ErrUnknownValidator  = errors.New("unknown validator")
	ErrValidatorMatching = errors.New("validator doesn't match field type")
	ErrLenValidator      = errors.New("len validator failed")
	ErrRegexpValidator   = errors.New("regexp validator failed")
	ErrInValidator       = errors.New("in validator failed")
	ErrMinValidator      = errors.New("min validator failed")
	ErrMaxValidator      = errors.New("max validator failed")
	ErrNestedValidator   = errors.New("nested validator failed")
)

func (v ValidationErrors) Error() string {
	var result string
	if len(v) > 0 {
		var sb strings.Builder
		sb.WriteString("validation error:\n")
		for _, validationError := range v {
			sb.WriteString(fmt.Sprintf("%v:%v\n", validationError.Field, validationError.Err))
		}
		result = sb.String()
	}
	return result
}

func (v ValidationErrors) Errorf() error {
	if len(v) > 0 {
		for i, validationError := range v {
			if i == 0 {
				errorsWrapped = fmt.Errorf("%v:%w", validationError.Field, validationError.Err)
			} else {
				errorsWrapped = fmt.Errorf("%w,%v:%w", errorsWrapped, validationError.Field, validationError.Err)
			}
		}
	}
	return errorsWrapped
}

func Validate(v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	structure := reflect.ValueOf(v)
	if structure.Kind() != reflect.Struct {
		return ErrStructureExpected
	}

	fieldCount := structure.Type().NumField()
	if fieldCount == 0 {
		return ErrEmptyStructure
	}

	for i := 0; i < fieldCount; i++ {
		field := structure.Type().Field(i)
		fieldValue := structure.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag
		if tag == "" {
			continue
		}

		tagValue, ok := tag.Lookup("validate")
		if !ok || tagValue == "" {
			continue
		}

		validatorSlice := strings.Split(tagValue, "|")
		for _, validator := range validatorSlice {
			validatorParsed := strings.Split(validator, ":")
			validatorType := validatorParsed[0]
			var intValidatorValue int
			var err error
			switch validatorType {
			case "len":
				if len(validatorParsed) != 2 {
					return ErrInvalidValidator
				}

				intValidatorValue, err = strconv.Atoi(validatorParsed[1])
				if err != nil {
					return ErrInvalidValidator
				}

				err = LenValidator(field, fieldValue, intValidatorValue)
				if err != nil {
					return err
				}
			case "regexp":
				if len(validatorParsed) != 2 {
					return ErrInvalidValidator
				}

				validatorValue := validatorParsed[1]
				if validatorValue == "" {
					return ErrInvalidValidator
				}

				err = RegexpValidator(field, fieldValue, validatorValue)
				if err != nil {
					return err
				}
			case "in":
				if len(validatorParsed) != 2 {
					return ErrInvalidValidator
				}

				validatorValue := validatorParsed[1]
				if validatorValue == "" {
					return ErrInvalidValidator
				}

				err = InValidator(field, fieldValue, validatorValue)
				if err != nil {
					return err
				}
			case "min":
				if len(validatorParsed) != 2 {
					return ErrInvalidValidator
				}

				intValidatorValue, err = strconv.Atoi(validatorParsed[1])
				if err != nil {
					return ErrInvalidValidator
				}

				err = MinValidator(field, fieldValue, intValidatorValue)
				if err != nil {
					return err
				}
			case "max":
				if len(validatorParsed) != 2 {
					return ErrInvalidValidator
				}

				intValidatorValue, err = strconv.Atoi(validatorParsed[1])
				if err != nil {
					return ErrInvalidValidator
				}

				err = MaxValidator(field, fieldValue, intValidatorValue)
				if err != nil {
					return err
				}
			case "nested":
				if len(validatorParsed) != 1 {
					return ErrInvalidValidator
				}
				NestedValidator(field)
			default:
				return ErrUnknownValidator
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors.Errorf()
	}

	return nil
}

func LenValidator(field reflect.StructField, fieldValue reflect.Value, intValidatorValue int) error {
	switch {
	case field.Type.Kind() == reflect.String:
		if len(fieldValue.String()) != intValidatorValue {
			validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrLenValidator})
		}
	case (field.Type.Kind() == reflect.Slice) && (field.Type.Elem().Kind() == reflect.String):
		for j := 0; j < fieldValue.Len(); j++ {
			if len(fieldValue.Index(j).String()) != intValidatorValue {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrLenValidator})
				break
			}
		}
	default:
		return ErrValidatorMatching
	}
	return nil
}

func RegexpValidator(field reflect.StructField, fieldValue reflect.Value, validatorValue string) error {
	re, err := regexp.Compile(validatorValue)
	if err != nil {
		return ErrInvalidValidator
	}

	switch {
	case field.Type.Kind() == reflect.String:
		stringValidatedValue := fieldValue.String()
		if !re.MatchString(stringValidatedValue) {
			validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrRegexpValidator})
		}
	case (field.Type.Kind() == reflect.Slice) && (field.Type.Elem().Kind() == reflect.String):
		for j := 0; j < fieldValue.Len(); j++ {
			stringValidatedValue := fieldValue.Index(j).String()
			if !re.MatchString(stringValidatedValue) {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrRegexpValidator})
				break
			}
		}
	default:
		return ErrValidatorMatching
	}
	return nil
}

func InValidator(field reflect.StructField, fieldValue reflect.Value, validatorValue string) error {
	inValidatorParsed := strings.Split(validatorValue, ",")
	switch {
	case field.Type.Kind() == reflect.String:
		if !StringIn(fieldValue.String(), inValidatorParsed) {
			validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrInValidator})
		}
	case field.Type.Kind() == reflect.Slice && (field.Type.Elem().Kind() == reflect.String):
		for j := 0; j < fieldValue.Len(); j++ {
			if !StringIn(fieldValue.Index(j).String(), inValidatorParsed) {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrInValidator})
				break
			}
		}
	case field.Type.Kind() == reflect.Int:
		ok, err := IntIn(fieldValue.Int(), inValidatorParsed)
		if err != nil {
			return err
		}

		if !ok {
			validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrInValidator})
		}
	case field.Type.Kind() == reflect.Slice && (field.Type.Elem().Kind() == reflect.Int):
		for j := 0; j < fieldValue.Len(); j++ {
			ok, err := IntIn(fieldValue.Index(j).Int(), inValidatorParsed)
			if err != nil {
				return err
			}

			if !ok {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrInValidator})
				break
			}
		}
	default:
		return ErrValidatorMatching
	}
	return nil
}

func MinValidator(field reflect.StructField, fieldValue reflect.Value, intValidatorValue int) error {
	switch {
	case field.Type.Kind() == reflect.Int:
		if fieldValue.Int() < int64(intValidatorValue) {
			validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrMinValidator})
		}
	case field.Type.Kind() == reflect.Slice && (field.Type.Elem().Kind() == reflect.Int):
		for j := 0; j < fieldValue.Len(); j++ {
			if fieldValue.Index(j).Int() < int64(intValidatorValue) {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrMinValidator})
				break
			}
		}
	default:
		return ErrValidatorMatching
	}
	return nil
}

func MaxValidator(field reflect.StructField, fieldValue reflect.Value, intValidatorValue int) error {
	switch {
	case field.Type.Kind() == reflect.Int:
		if fieldValue.Int() > int64(intValidatorValue) {
			validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrMaxValidator})
		}
	case field.Type.Kind() == reflect.Slice && (field.Type.Elem().Kind() == reflect.Int):
		for j := 0; j < fieldValue.Len(); j++ {
			if fieldValue.Index(j).Int() > int64(intValidatorValue) {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrMaxValidator})
				break
			}
		}
	default:
		return ErrValidatorMatching
	}
	return nil
}

func NestedValidator(field reflect.StructField) {
	if field.Type.Kind() != reflect.Struct {
		validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: ErrNestedValidator})
	}
}

func StringIn(str string, in []string) bool {
	var validated bool
	for _, matchString := range in {
		if str == matchString {
			validated = true
			break
		}
	}
	return validated
}

func IntIn(i int64, in []string) (bool, error) {
	var validated bool
	for _, matchString := range in {
		intValue, err := strconv.Atoi(matchString)
		if err != nil {
			return false, ErrInvalidValidator
		}
		if i == int64(intValue) {
			validated = true
			break
		}
	}
	return validated, nil
}
