package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	ValidatorError = iota
	ProgramError
)

type CustomError struct {
	ErrorType int
	Error     error
}

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

type Validator struct{}

var (
	errorsWrapped    error
	validationErrors ValidationErrors
	// Program and validator tags syntax errors.
	ErrNilValue          = errors.New("nil value in input")
	ErrStructureExpected = errors.New("structure kind expected in input")
	ErrEmptyStructure    = errors.New("empty structure")
	ErrInvalidValidator  = errors.New("invalid validator")
	ErrUnknownValidator  = errors.New("unknown validator")
	ErrConvertingValue   = errors.New("could not convert field value to field type")
	// Validation errors.
	ErrValidatorMatching = errors.New("validator doesn't match field type")
	ErrLenValidator      = errors.New("len validator failed")
	ErrRegexpValidator   = errors.New("regexp validator failed")
	ErrInValidator       = errors.New("in validator failed")
	ErrMinValidator      = errors.New("min validator failed")
	ErrMaxValidator      = errors.New("max validator failed")
	ErrNestedValidator   = errors.New("nested validator failed")
)

func (v ValidationErrors) Error() string {
	result := make([]string, 0)
	if len(v) > 0 {
		result = append(result, "validation errors:")
		for _, validationError := range v {
			errString := fmt.Sprintf("field: %v, error: %v", validationError.Field, validationError.Err)
			result = append(result, errString)
		}
		return strings.Join(result, "\n")
	}
	return ""
}

func (v ValidationErrors) Errorf() error {
	if len(v) > 0 {
		for i, validationError := range v {
			if i == 0 {
				errorsWrapped = fmt.Errorf("field: %v, error: %w", validationError.Field, validationError.Err)
			} else {
				errorsWrapped = fmt.Errorf("%w; field: %v, error: %w", errorsWrapped, validationError.Field, validationError.Err)
			}
		}
	}
	return errorsWrapped
}

func (v Validator) Len(validatorValue string, validatedValue any) CustomError {
	var (
		fieldKind, sliceKind reflect.Kind
		strValidatedValue    string
	)
	intValidatorValue, err := strconv.Atoi(validatorValue)
	if err != nil {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	fieldKind = reflect.TypeOf(validatedValue).Kind()
	if fieldKind == reflect.Slice {
		sliceKind = reflect.TypeOf(validatedValue).Elem().Kind()
	}

	switch {
	case fieldKind == reflect.String:
		strValidatedValue = reflect.ValueOf(validatedValue).String()
		if len(strValidatedValue) != intValidatorValue {
			err = ErrLenValidator
		}
	case fieldKind == reflect.Slice && sliceKind == reflect.String:
		strSlice, ok := validatedValue.([]string)
		if !ok {
			return CustomError{ProgramError, ErrConvertingValue}
		}
		for _, strValidatedValue = range strSlice {
			if len(strValidatedValue) != intValidatorValue {
				err = ErrLenValidator
				break
			}
		}
	default:
		return CustomError{ProgramError, ErrValidatorMatching}
	}

	return CustomError{ValidatorError, err}
}

func (v Validator) Regexp(validatorValue string, validatedValue any) CustomError {
	var (
		fieldKind, sliceKind reflect.Kind
		strValidatedValue    string
	)
	strValidatorValue := validatorValue
	if strValidatorValue == "" {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	re, err := regexp.Compile(strValidatorValue)
	if err != nil {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	fieldKind = reflect.TypeOf(validatedValue).Kind()
	if fieldKind == reflect.Slice {
		sliceKind = reflect.TypeOf(validatedValue).Elem().Kind()
	}

	switch {
	case fieldKind == reflect.String:
		strValidatedValue = reflect.ValueOf(validatedValue).String()
		if !re.MatchString(strValidatedValue) {
			err = ErrRegexpValidator
		}
	case fieldKind == reflect.Slice && sliceKind == reflect.String:
		strSlice, ok := validatedValue.([]string)
		if !ok {
			return CustomError{ProgramError, ErrConvertingValue}
		}
		for _, strValidatedValue = range strSlice {
			if !re.MatchString(strValidatedValue) {
				err = ErrRegexpValidator
				break
			}
		}
	default:
		return CustomError{ProgramError, ErrValidatorMatching}
	}

	return CustomError{ValidatorError, err}
}

func (v Validator) In(validatorValue string, validatedValue any) CustomError {
	var (
		fieldKind, sliceKind reflect.Kind
		ok                   bool
		err                  error
		strValidatedValue    string
	)
	if validatorValue == "" || validatorValue == "," {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	validatorValueParsed := strings.Split(validatorValue, ",")
	fieldKind = reflect.TypeOf(validatedValue).Kind()
	if fieldKind == reflect.Slice {
		sliceKind = reflect.TypeOf(validatedValue).Elem().Kind()
	}

	switch {
	case fieldKind == reflect.String:
		strValidatedValue = reflect.ValueOf(validatedValue).String()
		if !StringIn(strValidatedValue, validatorValueParsed) {
			err = ErrInValidator
		}
	case fieldKind == reflect.Slice && sliceKind == reflect.String:
		strSlice, ok := validatedValue.([]string)
		if !ok {
			return CustomError{ProgramError, ErrConvertingValue}
		}
		for _, strValidatedValue = range strSlice {
			if !StringIn(strValidatedValue, validatorValueParsed) {
				err = ErrInValidator
				break
			}
		}
	case fieldKind == reflect.Int:
		ok, err = IntIn(validatedValue.(int), validatorValueParsed)
		if err != nil {
			return CustomError{ProgramError, err}
		}

		if !ok {
			err = ErrInValidator
		}
	case fieldKind == reflect.Slice && sliceKind == reflect.Int:
		intSlice, ok := validatedValue.([]int)
		if !ok {
			return CustomError{ProgramError, ErrConvertingValue}
		}
		for _, intValidatedValue := range intSlice {
			ok, err = IntIn(intValidatedValue, validatorValueParsed)
			if err != nil {
				return CustomError{ProgramError, err}
			}

			if !ok {
				err = ErrInValidator
				break
			}
		}
	default:
		return CustomError{ProgramError, ErrValidatorMatching}
	}
	return CustomError{ValidatorError, err}
}

func (v Validator) Min(validatorValue string, validatedValue any) CustomError {
	var fieldKind, sliceKind reflect.Kind
	intValidatorValue, err := strconv.Atoi(validatorValue)
	if err != nil {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	fieldKind = reflect.TypeOf(validatedValue).Kind()
	if fieldKind == reflect.Slice {
		sliceKind = reflect.TypeOf(validatedValue).Elem().Kind()
	}

	switch {
	case fieldKind == reflect.Int:
		if validatedValue.(int) < intValidatorValue {
			err = ErrMinValidator
		}
	case fieldKind == reflect.Slice && sliceKind == reflect.Int:
		intSlice, ok := validatedValue.([]int)
		if !ok {
			return CustomError{ProgramError, ErrConvertingValue}
		}
		for _, intValidatedValue := range intSlice {
			if intValidatedValue < intValidatorValue {
				err = ErrMinValidator
				break
			}
		}
	default:
		return CustomError{ProgramError, ErrValidatorMatching}
	}
	return CustomError{ValidatorError, err}
}

func (v Validator) Max(validatorValue string, validatedValue any) CustomError {
	var fieldKind, sliceKind reflect.Kind
	intValidatorValue, err := strconv.Atoi(validatorValue)
	if err != nil {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	fieldKind = reflect.TypeOf(validatedValue).Kind()
	if fieldKind == reflect.Slice {
		sliceKind = reflect.TypeOf(validatedValue).Elem().Kind()
	}

	switch {
	case fieldKind == reflect.Int:
		if validatedValue.(int) > intValidatorValue {
			err = ErrMaxValidator
		}
	case fieldKind == reflect.Slice && sliceKind == reflect.Int:
		intSlice, ok := validatedValue.([]int)
		if !ok {
			return CustomError{ProgramError, ErrConvertingValue}
		}
		for _, intValidatedValue := range intSlice {
			if intValidatedValue > intValidatorValue {
				err = ErrMaxValidator
				break
			}
		}
	default:
		return CustomError{ProgramError, ErrValidatorMatching}
	}
	return CustomError{ValidatorError, err}
}

func (v Validator) Nested(validatorValue string, validatedValue any) CustomError {
	var err error
	if validatorValue != "" {
		return CustomError{ProgramError, ErrInvalidValidator}
	}

	if reflect.TypeOf(validatedValue).Kind() != reflect.Struct {
		err = ErrNestedValidator
	}
	return CustomError{ValidatorError, err}
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

func IntIn(i int, in []string) (bool, error) {
	var validated bool
	for _, matchString := range in {
		intValue, err := strconv.Atoi(matchString)
		if err != nil {
			return false, ErrInvalidValidator
		}
		if i == intValue {
			validated = true
			break
		}
	}
	return validated, nil
}

func Validate(v interface{}) error {
	structure, fieldCount, err := PrepareStructure(v)
	if err != nil {
		return err
	}

	validatorObj := Validator{}
	validatorObjValue := reflect.ValueOf(validatorObj)
	validatorObjType := reflect.TypeOf(validatorObj)

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
		if !ok || tagValue == "" || tagValue == "|" {
			continue
		}

		for _, validator := range strings.Split(tagValue, "|") {
			if validator == "" || strings.HasPrefix(validator, ":") {
				continue
			}

			validatorType, validatorValue, err := ParseValidator(validator)
			if err != nil {
				return err
			}

			_, validatorExist := validatorObjType.MethodByName(validatorType)
			if !validatorExist {
				return ErrUnknownValidator
			}

			params := []reflect.Value{reflect.ValueOf(validatorValue), fieldValue}
			output := validatorObjValue.MethodByName(validatorType).Call(params)[0].Interface().(CustomError)

			if output.ErrorType == ProgramError {
				return output.Error
			}

			if output.ErrorType == ValidatorError && output.Error != nil {
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: output.Error})
			}
		}
	}
	if len(validationErrors) > 0 {
		return validationErrors.Errorf()
	}
	return nil
}

func PrepareStructure(v interface{}) (structure reflect.Value, fieldCount int, err error) {
	if v == nil {
		return structure, fieldCount, ErrNilValue
	}

	structure = reflect.ValueOf(v)
	if structure.Kind() != reflect.Struct {
		return structure, fieldCount, ErrStructureExpected
	}

	fieldCount = structure.Type().NumField()
	if fieldCount == 0 {
		return structure, fieldCount, ErrEmptyStructure
	}
	return structure, fieldCount, nil
}

func ParseValidator(validator string) (validatorType, validatorValue string, err error) {
	validatorParsed := strings.Split(validator, ":")
	if len(validatorParsed) > 2 {
		return validatorType, validatorValue, ErrInvalidValidator
	}

	caser := cases.Title(language.Und)
	validatorType = caser.String(validatorParsed[0])
	if len(validatorParsed) == 2 {
		validatorValue = validatorParsed[1]
	}

	return validatorType, validatorValue, nil
}
