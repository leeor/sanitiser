package sanitiser

import "testing"

type testStruct1 struct {
	StringField    string                  `sanitise:"*"`
	IntField       int                     `sanitise:"testContext1"`
	FloatField     float64                 `sanitise:"testContext2"`
	ByteSliceField []byte                  `sanitise:"testContext1"`
	MapField       map[string]*testStruct1 `sanitise:"testContext1"`
	StructPtrField *testStruct1            `sanitise:"testContext2"`
	StructField    testStruct2             `sanitise:"testContext2"`
	InterfaceField interface{}             `sanitise:"testContext1"`
}

type testStruct2 struct {
	StringField    string  `sanitise:"*"`
	IntField       int     `sanitise:"testContext1"`
	FloatField     float64 `sanitise:"testContext2"`
	ByteSliceField []byte  `sanitise:"testContext1"`
}

func newTestObj(depth int) (obj *testStruct1) {

	obj = &testStruct1{"String Value", 4, 5.27, []byte("bytes"), map[string]*testStruct1{}, nil, testStruct2{}, "some string"}

	if depth > 0 {

		obj.MapField["testObj"] = newTestObj(depth - 1)
		obj.StructPtrField = newTestObj(depth - 1)
		obj.StructField = testStruct2{"Another String", 6, 90.27, []byte("and some more bytes")}
	}

	return
}

func (obj *testStruct1) expectContext1() *testStruct1 {

	obj.StringField = ""
	obj.IntField = 0
	obj.ByteSliceField = []byte("")
	obj.MapField = map[string]*testStruct1{}
	obj.InterfaceField = nil

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

	for _, v := range obj.MapField {

		v.expectContext2()
	}

	return obj
}

func (obj testStruct2) expectContext1() *testStruct2 {

	obj.StringField = ""
	obj.IntField = 0
	obj.ByteSliceField = []byte("")

	return &obj
}

func (obj testStruct2) expectContext2() *testStruct2 {

	obj.StringField = ""
	obj.FloatField = 0.0

	return &obj
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

	if this.InterfaceField != that.InterfaceField {

		t.Logf("Interface fields differ: %+v != %+v\n", this.InterfaceField, that.InterfaceField)
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

	return
}

func TestSimpleC1(t *testing.T) {

	testObjDepth := 1

	if testing.Verbose() {

		SetLogger(t.Logf)
	}

	o := newTestObj(testObjDepth)
	if err := Sanitise(o, "testContext1"); err != nil {

		t.Errorf("Call to Sanitise returned with error:\n%v", err)
	} else {

		expected := newTestObj(testObjDepth).expectContext1()

		if !expected.equals(*o, t) {

			t.Errorf("[testContext1] Sanitised object does not match expected results:\nExpected %+v\nSanitised %+v\n", expected, o)
		}
	}
}

func TestSimpleC2(t *testing.T) {

	testObjDepth := 1

	if testing.Verbose() {

		SetLogger(t.Logf)
	}

	o := newTestObj(testObjDepth)
	if err := Sanitise(o, "testContext2"); err != nil {

		t.Errorf("Call to Sanitise returned with error:\n%v", err)
	} else {

		expected := newTestObj(testObjDepth).expectContext2()

		if !expected.equals(*o, t) {

			t.Errorf("[testContext2] Sanitised object does not match expected results:\nExpected %+v\nSanitised %+v\n", expected, o)
		}
	}
}
