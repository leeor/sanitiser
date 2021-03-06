package sanitiser

import (
	"encoding/json"
	"testing"
)

type testStruct1 struct {
	StringField    string                  `sanitise:"*"`
	IntField       int                     `sanitise:"testContext1"`
	FloatField     float64                 `sanitise:"testContext2"`
	ByteSliceField []byte                  `sanitise:"testContext1"`
	MapField       map[string]*testStruct1 `sanitise:"testContext1"`
	ListField      []testStruct2           `sanitise:"testContext1"`
	PtrListField   []*testStruct2          `sanitise:"testContext1"`
	StructPtrField *testStruct1            `sanitise:"testContext2"`
	StructField    testStruct2             `sanitise:"testContext2"`
	InterfaceField interface{}             `sanitise:"testContext1"`
	AnotherInt     int
	hiddenIntField int `sanitise:"testContext1"`
}

type testStruct2 struct {
	StringField    string  `sanitise:"*"`
	IntField       int     `sanitise:"testContext1"`
	FloatField     float64 `sanitise:"testContext2"`
	ByteSliceField []byte  `sanitise:"testContext1"`
	AnotherString  string
	AnotherInt     int
}

func (o *testStruct1) Sanitise(context string) {

	if context == "testContext1" {

		o.AnotherInt = 0
	}
}

func (o *testStruct2) Sanitise(context string) {

	if context == "testContext1" {

		o.AnotherString = ""
	} else if context == "testContext2" {

		o.AnotherInt = 0
	}
}

func newTestObj(depth int) (obj *testStruct1) {

	obj = &testStruct1{"String Value", 4, 5.27, []byte("bytes"), map[string]*testStruct1{}, []testStruct2{}, []*testStruct2{}, nil, testStruct2{}, "some string", 7, 8}

	if depth > 0 {

		obj.MapField["testObj"] = newTestObj(depth - 1)
		obj.ListField = []testStruct2{testStruct2{"Another String", 6, 90.27, []byte("and some more bytes"), "Yet Another String", 8}}
		obj.PtrListField = []*testStruct2{&testStruct2{"Another String", 6, 90.27, []byte("and some more bytes"), "Yet Another String", 8}}
		obj.StructPtrField = newTestObj(depth - 1)
		obj.StructField = testStruct2{"Another String", 6, 90.27, []byte("and some more bytes"), "Yet Another String", 8}
		obj.InterfaceField = &testStruct2{"Another String", 6, 90.27, []byte("and some more bytes"), "Yet Another String", 8}
	}

	return
}

func (obj *testStruct1) expectContext1() *testStruct1 {

	obj.StringField = ""
	obj.IntField = 0
	obj.ByteSliceField = []byte("")
	obj.MapField = map[string]*testStruct1{}
	obj.ListField = []testStruct2{}
	obj.PtrListField = []*testStruct2{}
	obj.InterfaceField = nil
	obj.AnotherInt = 0

	if obj.StructPtrField != nil {

		obj.StructPtrField.expectContext1()
	}

	obj.StructField = *obj.StructField.expectContext1()

	return obj
}

func (obj *testStruct1) expectContext2() *testStruct1 {

	obj.StringField = ""
	obj.FloatField = 0.0
	obj.StructPtrField = nil
	obj.StructField = testStruct2{}

	if o, ok := obj.InterfaceField.(*testStruct2); ok {

		obj.InterfaceField = o.expectContext2()
	}

	for _, v := range obj.MapField {

		v.expectContext2()
	}

	for i, v := range obj.ListField {

		obj.ListField[i] = *v.expectContext2()
	}

	for _, v := range obj.PtrListField {

		v.expectContext2()
	}

	return obj
}

func (obj *testStruct2) expectContext1() *testStruct2 {

	obj.StringField = ""
	obj.IntField = 0
	obj.ByteSliceField = []byte("")
	obj.AnotherString = ""

	return obj
}

func (obj *testStruct2) expectContext2() *testStruct2 {

	obj.StringField = ""
	obj.FloatField = 0.0
	obj.AnotherInt = 0

	return obj
}

func (this testStruct1) equals(that testStruct1, t *testing.T) (equal bool) {

	equal = true

	if this.StringField != that.StringField {

		t.Logf("String fields differ: \"%+v\" != \"%+v\"\n", this.StringField, that.StringField)
		equal = false
	}

	if this.IntField != that.IntField {

		t.Logf("Int fields differ: %+v != %+v\n", this.IntField, that.IntField)
		equal = false
	}

	if this.FloatField != that.FloatField {

		t.Logf("Float fields differ: %+v != %+v\n", this.FloatField, that.FloatField)
		equal = false
	}

	if string(this.ByteSliceField) != string(that.ByteSliceField) {

		t.Logf("ByteSlice fields differ: %+v != %+v\n", this.ByteSliceField, that.ByteSliceField)
		equal = false
	}

	for k, v := range this.MapField {

		if !v.equals(*that.MapField[k], t) {

			t.Logf("Map members %v differ: %+v != %+v\n", k, v, this.MapField[k])
			equal = false
		}
	}

	for i, v := range this.ListField {

		if !v.equals(that.ListField[i], t) {

			t.Logf("List members %v differ: %+v != %+v\n", i, v, this.ListField[i])
			equal = false
		}
	}

	for i, v := range this.PtrListField {

		if !v.equals(*that.PtrListField[i], t) {

			t.Logf("List members %v differ: %+v != %+v\n", i, v, this.PtrListField[i])
			equal = false
		}
	}

	if this.StructPtrField != nil {

		if that.StructPtrField == nil {

			t.Logf("Struct pointer fields differ: %+v != %+v\n", this.StructPtrField, that.StructPtrField)
			equal = false
		} else if !this.StructPtrField.equals(*that.StructPtrField, t) {

			t.Logf("Struct pointer fields differ: %+v != %+v\n", *this.StructPtrField, *that.StructPtrField)
			equal = false
		}
	}

	if !this.StructField.equals(that.StructField, t) {

		t.Logf("Struct fields differ: %+v != %+v\n", this.StructField, that.StructField)
		equal = false
	}

	if this.InterfaceField != nil {

		if that.InterfaceField == nil {

			t.Logf("Interface fields differ: %+v != %+v\n", this.InterfaceField, that.InterfaceField)
			equal = false
		} else if s, ok := this.InterfaceField.(string); ok {

			if s2, ok := that.InterfaceField.(string); !ok || s != s2 {

				t.Logf("Interface fields differ: %+v != %+v\n", this.InterfaceField, that.InterfaceField)
				equal = false
			}
		} else if !this.InterfaceField.(*testStruct2).equals(*that.InterfaceField.(*testStruct2), t) {

			t.Logf("Interface fields differ: %+v != %+v\n", this.InterfaceField, that.InterfaceField)
			equal = false
		}
	}

	if this.AnotherInt != that.AnotherInt {

		t.Logf("Int fields differ: %+v != %+v\n", this.AnotherInt, that.AnotherInt)
		equal = false
	}

	return
}

func (this testStruct2) equals(that testStruct2, t *testing.T) (equal bool) {

	equal = true

	if this.StringField != that.StringField {

		t.Logf("String fields differ: \"%+v\" != \"%+v\"\n", this.StringField, that.StringField)
		equal = false
	}

	if this.IntField != that.IntField {

		t.Logf("Int fields differ: %+v != %+v\n", this.IntField, that.IntField)
		equal = false
	}

	if this.FloatField != that.FloatField {

		t.Logf("Float fields differ: %+v != %+v\n", this.FloatField, that.FloatField)
		equal = false
	}

	if string(this.ByteSliceField) != string(that.ByteSliceField) {

		t.Logf("ByteSlice fields differ: %+v != %+v\n", this.ByteSliceField, that.ByteSliceField)
		equal = false
	}

	if this.AnotherString != that.AnotherString {

		t.Logf("AnotherString fields differ: \"%+v\" != \"%+v\"\n", this.AnotherString, that.AnotherString)
		equal = false
	}

	if this.AnotherInt != that.AnotherInt {

		t.Logf("AnotherInt fields differ: \"%+v\" != \"%+v\"\n", this.AnotherInt, that.AnotherInt)
		equal = false
	}

	return
}

func TestSimpleC1(t *testing.T) {

	testObjDepth := 1

	if testing.Verbose() {

		SetLogger(t.Logf)
	}

	o := newTestObj(testObjDepth)

	sanitised := Sanitise(o, "testContext1")
	expected := newTestObj(testObjDepth).expectContext1()

	if testing.Verbose() {

		expected_json, _ := json.MarshalIndent(expected, "", "    ")
		o_json, _ := json.MarshalIndent(sanitised, "", "    ")

		t.Logf("Expecting: %v\nSanitised object: %v", string(expected_json), string(o_json))
	}

	if ts1, ok := sanitised.(*testStruct1); ok {

		if !expected.equals(*ts1, t) {

			t.Errorf("[testContext1] Sanitised object does not match expected results:\nExpected %+v\nSanitised %+v\n", expected, o)
		}
	} else {

		t.Errorf("[testContext1] Sanitised object is not of the expected type: %T instead of testStruct1\n", sanitised)
	}
}

func TestSimpleC2(t *testing.T) {

	testObjDepth := 1

	if testing.Verbose() {

		SetLogger(t.Logf)
	}

	o := newTestObj(testObjDepth)

	sanitised := Sanitise(o, "testContext2")
	expected := newTestObj(testObjDepth).expectContext2()

	if testing.Verbose() {

		expected_json, _ := json.MarshalIndent(expected, "", "    ")
		o_json, _ := json.MarshalIndent(sanitised, "", "    ")

		t.Logf("Expecting: %v\nSanitised object: %v", string(expected_json), string(o_json))
	}

	if ts1, ok := sanitised.(*testStruct1); ok {

		if !expected.equals(*ts1, t) {

			t.Errorf("[testContext2] Sanitised object does not match expected results:\nExpected %+v\nSanitised %+v\n", expected, o)
		}
	} else {

		t.Errorf("[testContext2] Sanitised object is not of the expected type: %T instead of testStruct1\n", sanitised)
	}
}
