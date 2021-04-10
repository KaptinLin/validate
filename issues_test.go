package validate

import (
	"fmt"
	"testing"
	"time"

	"github.com/gookit/goutil/dump"
	"github.com/stretchr/testify/assert"
)

func TestIssue2(t *testing.T) {
	type Fl struct {
		A float64 `validate:"float"`
	}

	fl := Fl{123}
	v := Struct(fl)
	assert.True(t, v.Validate())
	assert.Equal(t, float64(123), v.SafeVal("A"))

	val, ok := v.Raw("A")
	assert.True(t, ok)
	assert.Equal(t, float64(123), val)

	// Set value
	err := v.Set("A", float64(234))
	assert.Error(t, err)
	// field not exist
	err = v.Set("B", 234)
	assert.Error(t, err)

	// NOTICE: Must use ptr for set value
	v = Struct(&fl)
	err = v.Set("A", float64(234))
	assert.Nil(t, err)

	// check new value
	val, ok = v.Raw("A")
	assert.True(t, ok)
	assert.Equal(t, float64(234), val)

	// int will convert to float
	err = v.Set("A", 23)
	assert.Nil(t, err)

	// type is error
	err = v.Set("A", "abc")
	assert.Error(t, err)
	assert.Equal(t, errConvertFail.Error(), err.Error())
}

// https://github.com/gookit/validate/issues/19
func TestIssues19(t *testing.T) {
	is := assert.New(t)

	// use tag name: country_code
	type smsReq struct {
		CountryCode string `json:"country_code" validate:"required" filter:"trim|lower"`
		Phone       string `json:"phone" validate:"required" filter:"trim"`
		Type        string `json:"type" validate:"required|in:register,forget_password,set_pay_password,reset_pay_password,reset_password" filter:"trim"`
	}

	req := &smsReq{
		" ABcd   ", "13677778888  ", "register",
	}

	v := New(req)
	is.True(v.Validate())
	sd := v.SafeData()
	is.Equal("abcd", sd["CountryCode"])
	is.Equal("13677778888", sd["Phone"])

	// Notice: since 1.2, filtered value will update to struct
	// err := v.BindSafeData(req)
	// is.NoError(err)
	is.Equal("abcd", req.CountryCode)
	is.Equal("13677778888", req.Phone)

	// use tag name: countrycode
	type smsReq1 struct {
		// CountryCode string `json:"countryCode" validate:"required" filter:"trim|lower"`
		CountryCode string `json:"countrycode" validate:"required" filter:"trim|lower"`
		Phone       string `json:"phone" validate:"required" filter:"trim"`
		Type        string `json:"type" validate:"required|in:register,forget_password,set_pay_password,reset_pay_password,reset_password" filter:"trim"`
	}

	req1 := &smsReq1{
		" ABcd   ", "13677778888  ", "register",
	}

	v = New(req1)
	is.True(v.Validate())
	sd = v.SafeData()
	is.Equal("abcd", sd["CountryCode"])

	is.Equal("abcd", req1.CountryCode)
	is.Equal("13677778888", req1.Phone)
}

// https://github.com/gookit/validate/issues/20
func TestIssues20(t *testing.T) {
	is := assert.New(t)
	type setProfileReq struct {
		Nickname string `json:"nickname" validate:"string" filter:"trim"`
		Avatar   string `json:"avatar" validate:"required|url" filter:"trim"`
	}

	req := &setProfileReq{"123nickname111", "123"}
	v := New(req)
	is.True(v.Validate())

	type setProfileReq1 struct {
		Nickname string `json:"nickname" validate:"string" filter:"trim"`
		Avatar   string `json:"avatar" validate:"required|fullUrl" filter:"trim"`
	}
	req1 := &setProfileReq1{"123nickname111", "123"}

	Config(func(opt *GlobalOption) {
		opt.FieldTag = ""
	})
	v = New(req1)
	is.False(v.Validate())
	is.Len(v.Errors, 1)
	is.Equal("Avatar must be an valid full URL address", v.Errors.One())

	ResetOption()
	v = New(req1)
	is.False(v.Validate())
	is.Len(v.Errors, 1)
	is.Equal("avatar must be an valid full URL address", v.Errors.One())
}

// https://github.com/gookit/validate/issues/22
func TestIssues22(t *testing.T) {
	type userInfo0 struct {
		Nickname string `validate:"minLen:6" message:"OO! nickname min len is 6"`
		Avatar   string `validate:"maxLen:6" message:"OO! avatar max len is %d"`
	}

	is := assert.New(t)
	u0 := &userInfo0{
		Nickname: "tom",
		Avatar:   "https://github.com/gookit/validate/issues/22",
	}
	v := Struct(u0)
	is.False(v.Validate())
	is.Equal("OO! nickname min len is 6", v.Errors.FieldOne("Nickname"))
	u0 = &userInfo0{
		Nickname: "inhere",
		Avatar:   "some url",
	}
	v = Struct(u0)
	is.False(v.Validate())
	is.Equal("OO! avatar max len is 6", v.Errors.FieldOne("Avatar"))

	// multi messages
	type userInfo1 struct {
		Nickname string `validate:"required|minLen:6" message:"required:OO! nickname cannot be empty!|minLen:OO! nickname min len is %d"`
	}

	u1 := &userInfo1{Nickname: ""}
	v = Struct(u1)
	is.False(v.Validate())
	is.Equal("OO! nickname cannot be empty!", v.Errors.FieldOne("Nickname"))

	u1 = &userInfo1{Nickname: "tom"}
	v = Struct(u1)
	is.False(v.Validate())
	is.Equal("OO! nickname min len is 6", v.Errors.FieldOne("Nickname"))
}

// https://github.com/gookit/validate/issues/30
func TestIssues30(t *testing.T) {
	v := JSON(`{
   "cost_type": 10
}`)

	v.StringRule("cost_type", "str_num")

	assert.True(t, v.Validate())
	assert.Len(t, v.Errors, 0)
}

// https://github.com/gookit/validate/issues/34
func TestIssues34(t *testing.T) {
	type STATUS int32
	var s1 STATUS = 1

	// use custom validator
	v := New(M{
		"age": s1,
	})
	v.AddValidator("checkAge", func(val interface{}, ints ...int) bool {
		return Enum(int32(val.(STATUS)), ints)
	})
	v.StringRule("age", "required|checkAge:1,2,3,4")
	assert.True(t, v.Validate())

	// TODO refer https://golang.org/src/database/sql/driver/types.go?s=1210:1293#L29
	v = New(M{
		"age": s1,
	})
	v.StringRules(MS{
		"age": "required|in:1,2,3,4",
	})

	assert.NotContains(t, []int{1, 2, 3, 4}, s1)

	dump.Println(Enum(s1, []int{1, 2, 3, 4}), Enum(int32(s1), []int{1, 2, 3, 4}))

	assert.True(t, v.Validate())
	dump.Println(v.Errors)

	type someMode string
	var m1 someMode = "abc"
	v = New(M{
		"mode": m1,
	})
	v.StringRules(MS{
		"mode": "required|in:abc,def",
	})
	assert.True(t, v.Validate())

}

type issues36Form struct {
	Name  string `form:"username" json:"name" validate:"required|minLen:7"`
	Email string `form:"email" json:"email" validate:"email"`
	Age   int    `form:"age" validate:"required|int|min:18|max:150" json:"age"`
}

func (f issues36Form) Messages() map[string]string {
	return MS{
		"required":      "{field}不能为空",
		"Name.minLen":   "用户名最少7位",
		"Name.required": "用户名不能为空",
		"Email.email":   "邮箱格式不正确",
		"Age.min":       "年龄最少18岁",
		"Age.max":       "年龄最大150岁",
	}
}

func (f issues36Form) Translates() map[string]string {
	return MS{
		"Name":  "用户名",
		"Email": "邮箱",
		"Age":   "年龄",
	}
}

// https://github.com/gookit/validate/issues/36
func TestIssues36(t *testing.T) {
	f := issues36Form{Age: 10, Name: "i am tom", Email: "adc@xx.com"}

	v := Struct(&f)
	ok := v.Validate()

	assert.False(t, ok)
	assert.Equal(t, v.Errors.One(), "年龄最少18岁")
	assert.Contains(t, v.Errors.String(), "年龄最少18岁")
}

// https://github.com/gookit/validate/issues/60
func TestIssues60(t *testing.T) {
	is := assert.New(t)
	m := map[string]interface{}{
		"title": "1",
	}

	v := Map(m)
	v.StringRule("title", "in:2,3")
	v.AddMessages(map[string]string{
		"in": "自定义错误",
	})

	is.False(v.Validate())
	is.Equal("自定义错误", v.Errors.One())
}

// https://github.com/gookit/validate/issues/64
func TestPtrFieldValidation(t *testing.T) {
	type Foo struct {
		Name *string `validate:"in:henry,jim"`
	}

	name := "henry"
	v := New(&Foo{Name: &name})
	assert.True(t, v.Validate())

	name = "fish"
	valid := New(&Foo{Name: &name})
	assert.False(t, valid.Validate())
}

// ----- test case structs

type Org struct {
	Company string `validate:"in:A,B,C,D"`
}

type Info struct {
	Email string `validate:"email"  filter:"trim|lower"`
	Age   *int   `validate:"in:1,2,3,4"`
}

// anonymous struct nested
type User struct {
	*Info `validate:"required"`
	Org
	Name string `validate:"required|string" filter:"trim|lower"`
	Sex  string `validate:"string"`
	Time time.Time
}

// non-anonymous struct nested
type User2 struct {
	Name string `validate:"required|string" filter:"trim|lower"`
	In   Info
	Sex  string `validate:"string"`
	Time time.Time
}

type Info2 struct {
	Org
	Sub *Info
}

// gt 2 level struct nested
type User3 struct {
	In2 *Info2 `validate:"required"`
}

// https://github.com/gookit/validate/issues/58
func TestStructNested(t *testing.T) {
	// anonymous field test
	age := 3
	u := &User{
		Name: "fish",
		Info: &Info{
			Email: "fish_yww@163.com",
			Age:   &age,
		},
		Org:  Org{Company: "E"},
		Sex:  "male",
		Time: time.Now(),
	}

	// anonymous field test
	v := Struct(u)
	if v.Validate() {
		assert.True(t, v.Validate())
	} else {
		// Print error msg,verify valid
		fmt.Println("--- anonymous field test\n", v.Errors)
		assert.False(t, v.Validate())
	}

	// non-anonymous field test
	age = 3
	user2 := &User2{
		Name: "fish",
		In: Info{
			Email: "fish_yww@163.com",
			Age:   &age,
		},
		Sex:  "male",
		Time: time.Now(),
	}

	v2 := Struct(user2)
	if v2.Validate() {
		assert.True(t, v2.Validate())
	} else {
		// Print error msg,verify valid
		fmt.Printf("%v\n", v2.Errors)
		assert.False(t, v2.Validate())
	}
}

func TestStructNested_gt2level(t *testing.T) {
	age := 3
	u := &User3{
		In2: &Info2{
			Org: Org{Company: "E"},
			Sub: &Info{
				Email: "SOME@163.com ",
				Age:   &age,
			},
		},
	}

	v := Struct(u)
	ok := v.Validate()
	assert.False(t, ok)
	assert.Equal(t, "In2.Org.Company value must be in the enum [A B C D]", v.Errors.Random())
	fmt.Println(v.Errors)

	u.In2.Org.Company = "A"
	v = Struct(u)
	ok = v.Validate()
	assert.True(t, ok)
	assert.Equal(t, "some@163.com", u.In2.Sub.Email)
}

// https://github.com/gookit/validate/issues/78
func TestIssue78(t *testing.T) {
	type UserDto struct {
		Name string `validate:"required"`
		Sex  *bool  `validate:"required"`
	}

	// sex := true
	u := UserDto{
		Name: "abc",
		Sex:  nil,
	}

	// 创建 Validation 实例
	v := Struct(&u)
	if !v.Validate() {
		fmt.Println(v.Errors)
	} else {
		assert.True(t, v.Validate())
		fmt.Println("Success...")
	}
}

// https://gitee.com/inhere/validate/issues/I36T2B
func TestIssues_I36T2B(t *testing.T) {
	m := map[string]interface{}{
		"a": 0,
	}

	// 创建 Validation 实例
	v := Map(m)
	v.AddRule("a", "gt", 100)

	ok := v.Validate()
	assert.True(t, ok)

	v = Map(m)
	v.AddRule("a", "gt", 100).SetSkipEmpty(false)

	ok = v.Validate()
	assert.False(t, ok)
	assert.Equal(t, "a value should greater the 100", v.Errors.One())

	v = Map(m)
	v.AddRule("a", "required")
	v.AddRule("a", "gt", 100)

	ok = v.Validate()
	assert.False(t, ok)
	assert.Equal(t, "a is required and not empty", v.Errors.One())
}

// https://gitee.com/inhere/validate/issues/I3B3AV
func TestIssues_I3B3AV(t *testing.T) {
	m := map[string]interface{}{
		"a": 0.01,
		"b": float32(0.03),
	}

	v := Map(m)
	v.AddRule("a", "gt", 0)
	v.AddRule("b", "gt", 0)

	assert.True(t, v.Validate())
}

// https://github.com/gookit/validate/issues/92
func TestIssue_92(t *testing.T) {
	m := map[string]interface{}{
		"t": 1.1,
	}

	v := Map(m)
	v.FilterRule("t", "float")
	ok := v.Validate()

	assert.True(t, ok)
}
