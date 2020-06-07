package validator_test

import (
	"fmt"
	"testing"

	"github.com/asaskevich/govalidator"

	myValidator "github.com/xmlking/configor/validator"
)

type User struct {
	Id    int    `validate:"number,min=1,max=1000" valid:"type(string)"`
	Name  string `validate:"string,min=2,max=10" valid:"alphanum,ascii"`
	Bio   string `validate:"string" valid:"ipv4"`
	Email string `validate:"email" valid:"email"`
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

func TestAbs(t *testing.T) {
	user := User{
		Id:    0,
		Name:  "superlongstring",
		Bio:   "",
		Email: "foobar",
	}

	fmt.Println("Errors:")
	isValid, err := myValidator.ValidateStruct(user)
	if err != nil {
		errs := govalidator.ErrorsByField(err)
		cnt := 0
		for field, err := range errs {
			cnt++
			fmt.Printf("\t%d. %s: %s\n", cnt, field, err)
		}
	}
	println(isValid)

	isValid, err = govalidator.ValidateStruct(user)
	if err != nil {
		errs := govalidator.ErrorsByField(err)
		cnt := 0
		for field, err := range errs {
			cnt++
			fmt.Printf("\t%d. %s: %s\n", cnt, field, err)
		}
	}
	println(isValid)
}
