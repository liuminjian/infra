package base

import (
	"errors"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	vtzh "gopkg.in/go-playground/validator.v9/translations/zh"
	"strings"
	"sync"
)

var validate *validator.Validate
var translator ut.Translator
var mutex sync.Mutex

func InitValidator() *validator.Validate {
	if validate != nil {
		return validate
	}
	mutex.Lock()
	defer mutex.Unlock()

	if validate != nil {
		return validate
	}
	return NewValidator()
}

func NewValidator() *validator.Validate {
	validate = validator.New()
	//翻译
	cn := zh.New()
	uni := ut.New(cn, cn)
	var found bool
	translator, found = uni.GetTranslator("zh")
	if found {
		err := vtzh.RegisterDefaultTranslations(validate, translator)
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Error("Not found translator: zh")
	}
	return validate
}

func ValidateStruct(v interface{}) error {
	err := validate.Struct(v)
	if err != nil {
		_, ok := err.(*validator.InvalidValidationError)
		if ok {
			log.Error("验证错误:", err)
		}
		errs, ok := err.(validator.ValidationErrors)
		var msg []string
		if ok {
			for _, e := range errs {
				msg = append(msg, e.Translate(translator))
			}
			log.Error(msg)
		}
		return errors.New(strings.Join(msg, ","))
	}
	return nil
}
