package base

import (
	"errors"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	vtzh "gopkg.in/go-playground/validator.v9/translations/zh"
	"os"
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

	//注册自定义校验
	_ = validate.RegisterValidation("path", func(fl validator.FieldLevel) bool {
		// 检查配置文件位置
		_, err := os.Stat(fl.Field().String())
		if err != nil {
			return false
		} else {
			return true
		}
	})

	RegisterTagTranslation("path", map[string]string{
		"zh": "{0}不是一个有效路径或无权限访问。",
	})

	return validate
}

// 自定义翻译
func RegisterTagTranslation(tag string, messages map[string]string) {
	for _, message := range messages {
		_ = validate.RegisterTranslation(tag, translator, func(ut ut.Translator) error {
			return ut.Add(tag, message, false)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T(fe.Tag(), fe.Field())
			if err != nil {
				return fe.(error).Error()
			}
			return t
		})
	}
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
