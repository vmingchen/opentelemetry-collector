// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pdata

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	otlpcommon "go.opentelemetry.io/collector/internal/data/opentelemetry-proto-gen/common/v1"
)

func TestAttributeValue(t *testing.T) {
	v := NewAttributeValueString("abc")
	assert.EqualValues(t, AttributeValueSTRING, v.Type())
	assert.EqualValues(t, "abc", v.StringVal())

	v = NewAttributeValueInt(123)
	assert.EqualValues(t, AttributeValueINT, v.Type())
	assert.EqualValues(t, 123, v.IntVal())

	v = NewAttributeValueDouble(3.4)
	assert.EqualValues(t, AttributeValueDOUBLE, v.Type())
	assert.EqualValues(t, 3.4, v.DoubleVal())

	v = NewAttributeValueBool(true)
	assert.EqualValues(t, AttributeValueBOOL, v.Type())
	assert.EqualValues(t, true, v.BoolVal())

	v = NewAttributeValueNull()
	assert.EqualValues(t, AttributeValueNULL, v.Type())

	v.SetStringVal("abc")
	assert.EqualValues(t, AttributeValueSTRING, v.Type())
	assert.EqualValues(t, "abc", v.StringVal())

	v.SetIntVal(123)
	assert.EqualValues(t, AttributeValueINT, v.Type())
	assert.EqualValues(t, 123, v.IntVal())

	v.SetDoubleVal(3.4)
	assert.EqualValues(t, AttributeValueDOUBLE, v.Type())
	assert.EqualValues(t, 3.4, v.DoubleVal())

	v.SetBoolVal(true)
	assert.EqualValues(t, AttributeValueBOOL, v.Type())
	assert.EqualValues(t, true, v.BoolVal())
}

func fromVal(v interface{}) AttributeValue {
	switch val := v.(type) {
	case string:
		return NewAttributeValueString(val)
	case int:
		return NewAttributeValueInt(int64(val))
	case float64:
		return NewAttributeValueDouble(val)
	case map[string]interface{}:
		return fromMap(val)
	}
	return NewAttributeValueNull()
}

func fromMap(v map[string]interface{}) AttributeValue {
	m := NewAttributeMap()
	for k, v := range v {
		m.Insert(k, fromVal(v))
	}
	m.Sort()
	av := NewAttributeValueMap()
	av.SetMapVal(m)
	return av
}

func fromJSON(jsonStr string) AttributeValue {
	var src map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &src)
	if err != nil {
		panic("Invalid input jsonStr:" + jsonStr)
	}
	return fromMap(src)
}

func assertMapJSON(t *testing.T, expectedJSON string, actualMap AttributeValue) {
	assert.EqualValues(t, fromJSON(expectedJSON).MapVal(), actualMap.MapVal().Sort())
}

func TestAttributeValueMap(t *testing.T) {
	m1 := NewAttributeValueMap()
	assert.EqualValues(t, fromJSON(`{}`), m1)
	assert.EqualValues(t, AttributeValueMAP, m1.Type())
	assert.EqualValues(t, NewAttributeMap(), m1.MapVal())
	assert.EqualValues(t, 0, m1.MapVal().Len())

	m1.MapVal().InsertDouble("double_key", 123)
	assertMapJSON(t, `{"double_key":123}`, m1)
	assert.EqualValues(t, 1, m1.MapVal().Len())

	v, exists := m1.MapVal().Get("double_key")
	require.True(t, exists)
	assert.EqualValues(t, AttributeValueDOUBLE, v.Type())
	assert.EqualValues(t, 123, v.DoubleVal())

	// Create a second map.
	m2 := NewAttributeValueMap()
	assertMapJSON(t, `{}`, m2)
	assert.EqualValues(t, 0, m2.MapVal().Len())

	// Modify the source map that was inserted.
	m2.MapVal().UpsertString("key_in_child", "somestr")
	assertMapJSON(t, `{"key_in_child": "somestr"}`, m2)
	assert.EqualValues(t, 1, m2.MapVal().Len())

	// Insert the second map as a child. This should perform a deep copy.
	m1.MapVal().Insert("child_map", m2)
	assertMapJSON(t, `{"double_key":123, "child_map": {"key_in_child": "somestr"}}`, m1)
	assert.EqualValues(t, 2, m1.MapVal().Len())

	// Check that the map was correctly copied.
	childMap, exists := m1.MapVal().Get("child_map")
	require.True(t, exists)
	assert.EqualValues(t, AttributeValueMAP, childMap.Type())
	assert.EqualValues(t, 1, childMap.MapVal().Len())

	v, exists = childMap.MapVal().Get("key_in_child")
	require.True(t, exists)
	assert.EqualValues(t, AttributeValueSTRING, v.Type())
	assert.EqualValues(t, "somestr", v.StringVal())

	// Modify the source map m2 that was inserted into m1.
	m2.MapVal().UpdateString("key_in_child", "somestr2")
	assertMapJSON(t, `{"key_in_child": "somestr2"}`, m2)
	assert.EqualValues(t, 1, m2.MapVal().Len())

	// The child map inside m1 should not be modified.
	assertMapJSON(t, `{"double_key":123, "child_map": {"key_in_child": "somestr"}}`, m1)
	childMap, exists = m1.MapVal().Get("child_map")
	require.True(t, exists)
	v, exists = childMap.MapVal().Get("key_in_child")
	require.True(t, exists)
	assert.EqualValues(t, AttributeValueSTRING, v.Type())
	assert.EqualValues(t, "somestr", v.StringVal())

	// Now modify the inserted map (not the source)
	childMap.MapVal().UpdateString("key_in_child", "somestr3")
	assertMapJSON(t, `{"double_key":123, "child_map": {"key_in_child": "somestr3"}}`, m1)
	assert.EqualValues(t, 1, childMap.MapVal().Len())

	v, exists = childMap.MapVal().Get("key_in_child")
	require.True(t, exists)
	assert.EqualValues(t, AttributeValueSTRING, v.Type())
	assert.EqualValues(t, "somestr3", v.StringVal())

	// The source child map should not be modified.
	v, exists = m2.MapVal().Get("key_in_child")
	require.True(t, exists)
	assert.EqualValues(t, AttributeValueSTRING, v.Type())
	assert.EqualValues(t, "somestr2", v.StringVal())

	deleted := m1.MapVal().Delete("double_key")
	assert.True(t, deleted)
	assertMapJSON(t, `{"child_map": {"key_in_child": "somestr3"}}`, m1)
	assert.EqualValues(t, 1, m1.MapVal().Len())
	_, exists = m1.MapVal().Get("double_key")
	assert.False(t, exists)

	deleted = m1.MapVal().Delete("child_map")
	assert.True(t, deleted)
	assert.EqualValues(t, 0, m1.MapVal().Len())
	_, exists = m1.MapVal().Get("child_map")
	assert.False(t, exists)

	// Test nil KvlistValue case for MapVal() func.
	orig := &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_KvlistValue{KvlistValue: nil}}
	m1 = AttributeValue{orig: &orig}
	assert.EqualValues(t, NewAttributeMap(), m1.MapVal())
}

func createNilOrigSetAttributeValue() AttributeValue {
	var orig *otlpcommon.AnyValue
	return AttributeValue{orig: &orig}
}

func TestNilOrigSetAttributeValue(t *testing.T) {
	av := createNilOrigSetAttributeValue()
	av.SetStringVal("abc")
	assert.EqualValues(t, "abc", av.StringVal())

	av = createNilOrigSetAttributeValue()
	av.SetIntVal(123)
	assert.EqualValues(t, 123, av.IntVal())

	av = createNilOrigSetAttributeValue()
	av.SetBoolVal(true)
	assert.EqualValues(t, true, av.BoolVal())

	av = createNilOrigSetAttributeValue()
	av.SetDoubleVal(1.23)
	assert.EqualValues(t, 1.23, av.DoubleVal())

	av = createNilOrigSetAttributeValue()
	av.SetMapVal(NewAttributeMap())
	assert.EqualValues(t, NewAttributeMap(), av.MapVal())

	// Test nil KvlistValue case for MapVal() func.
	orig := &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_KvlistValue{KvlistValue: nil}}
	av = AttributeValue{orig: &orig}
	av.SetMapVal(NewAttributeMap())
	assert.EqualValues(t, NewAttributeMap(), av.MapVal())
}

func TestAttributeValueEqual(t *testing.T) {
	av1 := createNilOrigSetAttributeValue()
	av2 := createNilOrigSetAttributeValue()
	assert.True(t, av1.Equal(av2))

	av2 = NewAttributeValueString("abc")
	assert.False(t, av1.Equal(av2))
	assert.False(t, av2.Equal(av1))

	av1 = NewAttributeValueString("abc")
	assert.True(t, av1.Equal(av2))

	av2 = NewAttributeValueString("edf")
	assert.False(t, av1.Equal(av2))

	av2 = NewAttributeValueInt(123)
	assert.False(t, av1.Equal(av2))
	assert.False(t, av2.Equal(av1))

	av1 = NewAttributeValueInt(234)
	assert.False(t, av1.Equal(av2))

	av1 = NewAttributeValueInt(123)
	assert.True(t, av1.Equal(av2))

	av2 = NewAttributeValueDouble(123)
	assert.False(t, av1.Equal(av2))
	assert.False(t, av2.Equal(av1))

	av1 = NewAttributeValueDouble(234)
	assert.False(t, av1.Equal(av2))

	av1 = NewAttributeValueDouble(123)
	assert.True(t, av1.Equal(av2))

	av2 = NewAttributeValueBool(true)
	assert.False(t, av1.Equal(av2))
	assert.False(t, av2.Equal(av1))

	av1 = NewAttributeValueBool(true)
	assert.True(t, av1.Equal(av2))

	av1 = NewAttributeValueBool(false)
	assert.False(t, av1.Equal(av2))
}

func TestNewAttributeValueSlice(t *testing.T) {
	events := NewAttributeValueSlice(0)
	assert.EqualValues(t, 0, len(events))

	n := rand.Intn(10)
	events = NewAttributeValueSlice(n)
	assert.EqualValues(t, n, len(events))
	for event := range events {
		assert.NotNil(t, event)
	}
}

func TestNilAttributeMap(t *testing.T) {
	assert.EqualValues(t, 0, NewAttributeMap().Len())

	val, exist := NewAttributeMap().Get("test_key")
	assert.False(t, exist)
	assert.EqualValues(t, AttributeValue{nil}, val)

	insertMap := NewAttributeMap()
	insertMap.Insert("k", NewAttributeValueString("v"))
	assert.EqualValues(t, generateTestAttributeMap(), insertMap)

	insertMapString := NewAttributeMap()
	insertMapString.InsertString("k", "v")
	assert.EqualValues(t, generateTestAttributeMap(), insertMapString)

	insertMapNull := NewAttributeMap()
	insertMapNull.InsertNull("k")
	assert.EqualValues(t, generateTestNullAttributeMap(), insertMapNull)

	insertMapInt := NewAttributeMap()
	insertMapInt.InsertInt("k", 123)
	assert.EqualValues(t, generateTestIntAttributeMap(), insertMapInt)

	insertMapDouble := NewAttributeMap()
	insertMapDouble.InsertDouble("k", 12.3)
	assert.EqualValues(t, generateTestDoubleAttributeMap(), insertMapDouble)

	insertMapBool := NewAttributeMap()
	insertMapBool.InsertBool("k", true)
	assert.EqualValues(t, generateTestBoolAttributeMap(), insertMapBool)

	updateMap := NewAttributeMap()
	updateMap.Update("k", NewAttributeValueString("v"))
	assert.EqualValues(t, NewAttributeMap(), updateMap)

	updateMapString := NewAttributeMap()
	updateMapString.UpdateString("k", "v")
	assert.EqualValues(t, NewAttributeMap(), updateMapString)

	updateMapInt := NewAttributeMap()
	updateMapInt.UpdateInt("k", 123)
	assert.EqualValues(t, NewAttributeMap(), updateMapInt)

	updateMapDouble := NewAttributeMap()
	updateMapDouble.UpdateDouble("k", 12.3)
	assert.EqualValues(t, NewAttributeMap(), updateMapDouble)

	updateMapBool := NewAttributeMap()
	updateMapBool.UpdateBool("k", true)
	assert.EqualValues(t, NewAttributeMap(), updateMapBool)

	upsertMap := NewAttributeMap()
	upsertMap.Upsert("k", NewAttributeValueString("v"))
	assert.EqualValues(t, generateTestAttributeMap(), upsertMap)

	upsertMapString := NewAttributeMap()
	upsertMapString.UpsertString("k", "v")
	assert.EqualValues(t, generateTestAttributeMap(), upsertMapString)

	upsertMapInt := NewAttributeMap()
	upsertMapInt.UpsertInt("k", 123)
	assert.EqualValues(t, generateTestIntAttributeMap(), upsertMapInt)

	upsertMapDouble := NewAttributeMap()
	upsertMapDouble.UpsertDouble("k", 12.3)
	assert.EqualValues(t, generateTestDoubleAttributeMap(), upsertMapDouble)

	upsertMapBool := NewAttributeMap()
	upsertMapBool.UpsertBool("k", true)
	assert.EqualValues(t, generateTestBoolAttributeMap(), upsertMapBool)

	deleteMap := NewAttributeMap()
	assert.False(t, deleteMap.Delete("k"))
	assert.EqualValues(t, NewAttributeMap(), deleteMap)

	// Test Sort
	assert.EqualValues(t, NewAttributeMap(), NewAttributeMap().Sort())
}

func TestAttributeMapWithNilValues(t *testing.T) {
	origWithNil := []*otlpcommon.KeyValue{
		nil,
		{
			Key:   "test_key",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "test_value"}},
		},
		nil,
		{
			Key:   "test_key2",
			Value: nil,
		},
		nil,
		{
			Key:   "test_key3",
			Value: &otlpcommon.AnyValue{Value: nil},
		},
	}
	sm := AttributeMap{
		orig: &origWithNil,
	}
	val, exist := sm.Get("test_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "test_value", val.StringVal())

	val, exist = sm.Get("test_key2")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueNULL, val.Type())
	assert.EqualValues(t, "", val.StringVal())

	val, exist = sm.Get("test_key3")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueNULL, val.Type())
	assert.EqualValues(t, "", val.StringVal())

	sm.Insert("other_key", NewAttributeValueString("other_value"))
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "other_value", val.StringVal())

	sm.InsertString("other_key_string", "other_value")
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "other_value", val.StringVal())

	sm.InsertInt("other_key_int", 123)
	val, exist = sm.Get("other_key_int")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueINT, val.Type())
	assert.EqualValues(t, 123, val.IntVal())

	sm.InsertDouble("other_key_double", 1.23)
	val, exist = sm.Get("other_key_double")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueDOUBLE, val.Type())
	assert.EqualValues(t, 1.23, val.DoubleVal())

	sm.InsertBool("other_key_bool", true)
	val, exist = sm.Get("other_key_bool")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueBOOL, val.Type())
	assert.EqualValues(t, true, val.BoolVal())

	sm.Update("other_key", NewAttributeValueString("yet_another_value"))
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "yet_another_value", val.StringVal())

	sm.UpdateString("other_key_string", "yet_another_value")
	val, exist = sm.Get("other_key_string")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "yet_another_value", val.StringVal())

	sm.UpdateInt("other_key_int", 456)
	val, exist = sm.Get("other_key_int")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueINT, val.Type())
	assert.EqualValues(t, 456, val.IntVal())

	sm.UpdateDouble("other_key_double", 4.56)
	val, exist = sm.Get("other_key_double")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueDOUBLE, val.Type())
	assert.EqualValues(t, 4.56, val.DoubleVal())

	sm.UpdateBool("other_key_bool", false)
	val, exist = sm.Get("other_key_bool")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueBOOL, val.Type())
	assert.EqualValues(t, false, val.BoolVal())

	sm.Upsert("other_key", NewAttributeValueString("other_value"))
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "other_value", val.StringVal())

	sm.UpsertString("other_key_string", "other_value")
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "other_value", val.StringVal())

	sm.UpsertInt("other_key_int", 123)
	val, exist = sm.Get("other_key_int")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueINT, val.Type())
	assert.EqualValues(t, 123, val.IntVal())

	sm.UpsertDouble("other_key_double", 1.23)
	val, exist = sm.Get("other_key_double")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueDOUBLE, val.Type())
	assert.EqualValues(t, 1.23, val.DoubleVal())

	sm.UpsertBool("other_key_bool", true)
	val, exist = sm.Get("other_key_bool")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueBOOL, val.Type())
	assert.EqualValues(t, true, val.BoolVal())

	sm.Upsert("yet_another_key", NewAttributeValueString("yet_another_value"))
	val, exist = sm.Get("yet_another_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "yet_another_value", val.StringVal())

	sm.UpsertString("yet_another_key_string", "yet_another_value")
	val, exist = sm.Get("yet_another_key_string")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "yet_another_value", val.StringVal())

	sm.UpsertInt("yet_another_key_int", 456)
	val, exist = sm.Get("yet_another_key_int")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueINT, val.Type())
	assert.EqualValues(t, 456, val.IntVal())

	sm.UpsertDouble("yet_another_key_double", 4.56)
	val, exist = sm.Get("yet_another_key_double")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueDOUBLE, val.Type())
	assert.EqualValues(t, 4.56, val.DoubleVal())

	sm.UpsertBool("yet_another_key_bool", false)
	val, exist = sm.Get("yet_another_key_bool")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueBOOL, val.Type())
	assert.EqualValues(t, false, val.BoolVal())

	assert.True(t, sm.Delete("other_key"))
	assert.True(t, sm.Delete("other_key_string"))
	assert.True(t, sm.Delete("other_key_int"))
	assert.True(t, sm.Delete("other_key_double"))
	assert.True(t, sm.Delete("other_key_bool"))
	assert.True(t, sm.Delete("yet_another_key"))
	assert.True(t, sm.Delete("yet_another_key_string"))
	assert.True(t, sm.Delete("yet_another_key_int"))
	assert.True(t, sm.Delete("yet_another_key_double"))
	assert.True(t, sm.Delete("yet_another_key_bool"))
	assert.False(t, sm.Delete("other_key"))
	assert.False(t, sm.Delete("yet_another_key"))

	// Test that the initial key is still there.
	val, exist = sm.Get("test_key")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueSTRING, val.Type())
	assert.EqualValues(t, "test_value", val.StringVal())

	val, exist = sm.Get("test_key2")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueNULL, val.Type())
	assert.EqualValues(t, "", val.StringVal())

	val, exist = sm.Get("test_key3")
	assert.True(t, exist)
	assert.EqualValues(t, AttributeValueNULL, val.Type())
	assert.EqualValues(t, "", val.StringVal())

	// Test Sort
	assert.EqualValues(t, AttributeMap{orig: &origWithNil}, sm.Sort())
}

func TestAttributeMapIterationNil(t *testing.T) {
	NewAttributeMap().ForEach(func(k string, v AttributeValue) {
		// Fail if any element is returned
		t.Fail()
	})
}

func TestAttributeMap_ForEach(t *testing.T) {
	rawMap := map[string]AttributeValue{
		"k_string": NewAttributeValueString("123"),
		"k_int":    NewAttributeValueInt(123),
		"k_double": NewAttributeValueDouble(1.23),
		"k_bool":   NewAttributeValueBool(true),
		"k_null":   NewAttributeValueNull(),
	}
	am := NewAttributeMap().InitFromMap(rawMap)
	assert.EqualValues(t, 5, am.Len())

	am.ForEach(func(k string, v AttributeValue) {
		assert.True(t, v.Equal(rawMap[k]))
		delete(rawMap, k)
	})
	assert.EqualValues(t, 0, len(rawMap))
}

func TestAttributeMap_ForEach_WithNils(t *testing.T) {
	rawMap := map[string]AttributeValue{
		"k_string": NewAttributeValueString("123"),
		"k_int":    NewAttributeValueInt(123),
		"k_double": NewAttributeValueDouble(1.23),
		"k_bool":   NewAttributeValueBool(true),
		"k_null":   NewAttributeValueNull(),
		"k_null2":  NewAttributeValueNull(),
	}
	rawOrigWithNil := []*otlpcommon.KeyValue{
		nil,
		newAttributeKeyValueString("k_string", "123"),
		nil,
		newAttributeKeyValueInt("k_int", 123),
		nil,
		newAttributeKeyValueDouble("k_double", 1.23),
		nil,
		newAttributeKeyValueBool("k_bool", true),
		nil,
		newAttributeKeyValueNull("k_null"),
		nil,
		{Key: "k_null2", Value: nil},
		nil,
	}
	am := AttributeMap{
		orig: &rawOrigWithNil,
	}
	assert.EqualValues(t, 13, am.Len())

	am.ForEach(func(k string, v AttributeValue) {
		assert.True(t, v.Equal(rawMap[k]))
		delete(rawMap, k)
	})
	assert.EqualValues(t, 0, len(rawMap))
}

func TestAttributeMap_InitFromMap(t *testing.T) {
	am := NewAttributeMap().InitFromMap(map[string]AttributeValue(nil))
	assert.EqualValues(t, NewAttributeMap(), am)

	rawMap := map[string]AttributeValue{
		"k_string": NewAttributeValueString("123"),
		"k_int":    NewAttributeValueInt(123),
		"k_double": NewAttributeValueDouble(1.23),
		"k_bool":   NewAttributeValueBool(true),
		"k_null":   NewAttributeValueNull(),
	}
	rawOrig := []*otlpcommon.KeyValue{
		newAttributeKeyValueString("k_string", "123"),
		newAttributeKeyValueInt("k_int", 123),
		newAttributeKeyValueDouble("k_double", 1.23),
		newAttributeKeyValueBool("k_bool", true),
		newAttributeKeyValueNull("k_null"),
	}
	am = NewAttributeMap().InitFromMap(rawMap)
	assert.EqualValues(t, AttributeMap{orig: &rawOrig}.Sort(), am.Sort())
}

func TestAttributeMap_CopyTo(t *testing.T) {
	dest := NewAttributeMap()
	// Test CopyTo to empty
	NewAttributeMap().CopyTo(dest)
	assert.EqualValues(t, 0, dest.Len())

	// Test CopyTo larger slice
	generateTestAttributeMap().CopyTo(dest)
	assert.EqualValues(t, generateTestAttributeMap(), dest)

	// Test CopyTo same size slice
	generateTestAttributeMap().CopyTo(dest)
	assert.EqualValues(t, generateTestAttributeMap(), dest)

	// Test CopyTo with a nil Value in the destination
	(*dest.orig)[0].Value = nil
	generateTestAttributeMap().CopyTo(dest)
	assert.EqualValues(t, generateTestAttributeMap(), dest)
}

func TestAttributeValue_copyTo(t *testing.T) {
	av := createNilOrigSetAttributeValue()
	destVal := otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{}}
	av.copyTo(&destVal)
	assert.EqualValues(t, nil, destVal.Value)
}

func TestAttributeMap_UpdateWithNilValues(t *testing.T) {
	origWithNil := []*otlpcommon.KeyValue{
		nil,
		{
			Key:   "test_key",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "test_value"}},
		},
		nil,
		{
			Key:   "test_key2",
			Value: nil,
		},
		nil,
		{
			Key:   "test_key3",
			Value: &otlpcommon.AnyValue{Value: nil},
		},
	}
	sm := AttributeMap{
		orig: &origWithNil,
	}

	av, exists := sm.Get("test_key")
	assert.True(t, exists)
	assert.EqualValues(t, AttributeValueSTRING, av.Type())
	assert.EqualValues(t, "test_value", av.StringVal())
	av.SetIntVal(123)

	av2, exists := sm.Get("test_key")
	assert.True(t, exists)
	assert.EqualValues(t, AttributeValueINT, av2.Type())
	assert.EqualValues(t, 123, av2.IntVal())

	av, exists = sm.Get("test_key2")
	assert.True(t, exists)
	assert.EqualValues(t, AttributeValueNULL, av.Type())
	assert.EqualValues(t, "", av.StringVal())
	av.SetIntVal(123)

	av2, exists = sm.Get("test_key2")
	assert.True(t, exists)
	assert.EqualValues(t, AttributeValueINT, av2.Type())
	assert.EqualValues(t, 123, av2.IntVal())

	av, exists = sm.Get("test_key3")
	assert.True(t, exists)
	assert.EqualValues(t, AttributeValueNULL, av.Type())
	assert.EqualValues(t, "", av.StringVal())
	av.SetBoolVal(true)

	av2, exists = sm.Get("test_key3")
	assert.True(t, exists)
	assert.EqualValues(t, AttributeValueBOOL, av2.Type())
	assert.EqualValues(t, true, av2.BoolVal())
}

func TestNilStringMap(t *testing.T) {
	assert.EqualValues(t, 0, NewStringMap().Len())

	val, exist := NewStringMap().Get("test_key")
	assert.False(t, exist)
	assert.EqualValues(t, StringValue{nil}, val)

	insertMap := NewStringMap()
	insertMap.Insert("k", "v")
	assert.EqualValues(t, generateTestStringMap(), insertMap)

	updateMap := NewStringMap()
	updateMap.Update("k", "v")
	assert.EqualValues(t, NewStringMap(), updateMap)

	upsertMap := NewStringMap()
	upsertMap.Upsert("k", "v")
	assert.EqualValues(t, generateTestStringMap(), upsertMap)

	deleteMap := NewStringMap()
	assert.False(t, deleteMap.Delete("k"))
	assert.EqualValues(t, NewStringMap(), deleteMap)

	// Test Sort
	assert.EqualValues(t, NewStringMap(), NewStringMap().Sort())
}

func TestStringMapWithNilValues(t *testing.T) {
	origWithNil := []*otlpcommon.StringKeyValue{
		nil,
		{
			Key:   "test_key",
			Value: "test_value",
		},
		nil,
	}
	sm := StringMap{
		orig: &origWithNil,
	}
	val, exist := sm.Get("test_key")
	assert.True(t, exist)
	assert.EqualValues(t, "test_value", val.Value())

	sm.Insert("other_key", "other_value")
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, "other_value", val.Value())

	sm.Update("other_key", "yet_another_value")
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, "yet_another_value", val.Value())

	sm.Upsert("other_key", "other_value")
	val, exist = sm.Get("other_key")
	assert.True(t, exist)
	assert.EqualValues(t, "other_value", val.Value())

	sm.Upsert("yet_another_key", "yet_another_value")
	val, exist = sm.Get("yet_another_key")
	assert.True(t, exist)
	assert.EqualValues(t, "yet_another_value", val.Value())

	assert.True(t, sm.Delete("other_key"))
	assert.True(t, sm.Delete("yet_another_key"))
	assert.False(t, sm.Delete("other_key"))
	assert.False(t, sm.Delete("yet_another_key"))

	// Test that the initial key is still there.
	val, exist = sm.Get("test_key")
	assert.True(t, exist)
	assert.EqualValues(t, "test_value", val.Value())

	// Test Sort
	assert.EqualValues(t, StringMap{orig: &origWithNil}, sm.Sort())
}

func TestStringMap(t *testing.T) {
	origRawMap := map[string]string{"k0": "v0", "k1": "v1", "k2": "v2"}
	origMap := NewStringMap().InitFromMap(origRawMap)
	sm := NewStringMap().InitFromMap(origRawMap)
	assert.EqualValues(t, 3, sm.Len())

	val, exist := sm.Get("k2")
	assert.True(t, exist)
	assert.EqualValues(t, "v2", val.Value())

	val, exist = sm.Get("k3")
	assert.False(t, exist)
	assert.EqualValues(t, StringValue{nil}, val)

	sm.Insert("k1", "v1")
	assert.EqualValues(t, origMap.Sort(), sm.Sort())
	sm.Insert("k3", "v3")
	assert.EqualValues(t, 4, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k0": "v0", "k1": "v1", "k2": "v2", "k3": "v3"}).Sort(), sm.Sort())
	assert.True(t, sm.Delete("k3"))
	assert.EqualValues(t, 3, sm.Len())
	assert.EqualValues(t, origMap.Sort(), sm.Sort())

	sm.Update("k3", "v3")
	assert.EqualValues(t, 3, sm.Len())
	assert.EqualValues(t, origMap.Sort(), sm.Sort())
	sm.Update("k2", "v3")
	assert.EqualValues(t, 3, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k0": "v0", "k1": "v1", "k2": "v3"}).Sort(), sm.Sort())
	sm.Update("k2", "v2")
	assert.EqualValues(t, 3, sm.Len())
	assert.EqualValues(t, origMap.Sort(), sm.Sort())

	sm.Upsert("k3", "v3")
	assert.EqualValues(t, 4, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k0": "v0", "k1": "v1", "k2": "v2", "k3": "v3"}).Sort(), sm.Sort())
	sm.Upsert("k1", "v5")
	assert.EqualValues(t, 4, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k0": "v0", "k1": "v5", "k2": "v2", "k3": "v3"}).Sort(), sm.Sort())
	sm.Upsert("k1", "v1")
	assert.EqualValues(t, 4, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k0": "v0", "k1": "v1", "k2": "v2", "k3": "v3"}).Sort(), sm.Sort())
	assert.True(t, sm.Delete("k3"))
	assert.EqualValues(t, 3, sm.Len())
	assert.EqualValues(t, origMap.Sort(), sm.Sort())

	assert.EqualValues(t, false, sm.Delete("k3"))
	assert.EqualValues(t, 3, sm.Len())
	assert.EqualValues(t, origMap.Sort(), sm.Sort())

	assert.True(t, sm.Delete("k0"))
	assert.EqualValues(t, 2, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k1": "v1", "k2": "v2"}).Sort(), sm.Sort())
	assert.True(t, sm.Delete("k2"))
	assert.EqualValues(t, 1, sm.Len())
	assert.EqualValues(t, NewStringMap().InitFromMap(map[string]string{"k1": "v1"}).Sort(), sm.Sort())
	assert.True(t, sm.Delete("k1"))
	assert.EqualValues(t, 0, sm.Len())
}

func TestStringMapIterationNil(t *testing.T) {
	NewStringMap().ForEach(func(k string, v StringValue) {
		// Fail if any element is returned
		t.Fail()
	})
}

func TestStringMap_ForEach(t *testing.T) {
	rawMap := map[string]string{"k0": "v0", "k1": "v1", "k2": "v2"}
	sm := NewStringMap().InitFromMap(rawMap)
	assert.EqualValues(t, 3, sm.Len())

	sm.ForEach(func(k string, v StringValue) {
		assert.EqualValues(t, rawMap[k], v.Value())
		delete(rawMap, k)
	})
	assert.EqualValues(t, 0, len(rawMap))
}

func TestStringMap_ForEach_WithNils(t *testing.T) {
	rawMap := map[string]string{"k0": "v0", "k1": "v1", "k2": "v2"}
	rawOrigWithNil := []*otlpcommon.StringKeyValue{
		nil,
		{
			Key:   "k0",
			Value: "v0",
		},
		nil,
		{
			Key:   "k1",
			Value: "v1",
		},
		nil,
		{
			Key:   "k2",
			Value: "v2",
		},
		nil,
	}
	sm := StringMap{
		orig: &rawOrigWithNil,
	}
	assert.EqualValues(t, 7, sm.Len())

	sm.ForEach(func(k string, v StringValue) {
		assert.EqualValues(t, rawMap[k], v.Value())
		delete(rawMap, k)
	})
	assert.EqualValues(t, 0, len(rawMap))
}

func TestStringMap_CopyTo(t *testing.T) {
	dest := NewStringMap()
	// Test CopyTo to empty
	NewStringMap().CopyTo(dest)
	assert.EqualValues(t, 0, dest.Len())

	// Test CopyTo larger slice
	generateTestStringMap().CopyTo(dest)
	assert.EqualValues(t, generateTestStringMap(), dest)

	// Test CopyTo same size slice
	generateTestStringMap().CopyTo(dest)
	assert.EqualValues(t, generateTestStringMap(), dest)
}

func TestStringMap_InitFromMap(t *testing.T) {
	sm := NewStringMap().InitFromMap(map[string]string(nil))
	assert.EqualValues(t, NewStringMap(), sm)

	rawMap := map[string]string{"k0": "v0", "k1": "v1", "k2": "v2"}
	rawOrig := []*otlpcommon.StringKeyValue{
		{
			Key:   "k0",
			Value: "v0",
		},
		{
			Key:   "k1",
			Value: "v1",
		},
		{
			Key:   "k2",
			Value: "v2",
		},
	}
	sm = NewStringMap().InitFromMap(rawMap)
	assert.EqualValues(t, StringMap{orig: &rawOrig}.Sort(), sm.Sort())
}

func BenchmarkAttributeValue_CopyTo(b *testing.B) {
	av := NewAttributeValueString("k")
	c := NewAttributeValueInt(123)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.copyTo(*av.orig)
	}
	if av.IntVal() != 123 {
		b.Fail()
	}
}

func BenchmarkAttributeValue_SetIntVal(b *testing.B) {
	av := NewAttributeValueString("k")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		av.SetIntVal(int64(n))
	}
	if av.IntVal() != int64(b.N-1) {
		b.Fail()
	}
}

func BenchmarkAttributeMap_ForEach(b *testing.B) {
	const numElements = 20
	rawOrigWithNil := make([]*otlpcommon.KeyValue, 2*numElements)
	for i := 0; i < numElements; i++ {
		rawOrigWithNil[i*2] = &otlpcommon.KeyValue{
			Key:   "k" + strconv.Itoa(i),
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "v" + strconv.Itoa(i)}},
		}
	}
	am := AttributeMap{
		orig: &rawOrigWithNil,
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		numEls := 0
		am.ForEach(func(k string, v AttributeValue) {
			numEls++
		})
		if numEls != numElements {
			b.Fail()
		}
	}
}

func BenchmarkAttributeMap_RangeOverMap(b *testing.B) {
	const numElements = 20
	rawOrig := make(map[string]AttributeValue, numElements)
	for i := 0; i < numElements; i++ {
		key := "k" + strconv.Itoa(i)
		rawOrig[key] = NewAttributeValueString("v" + strconv.Itoa(i))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		numEls := 0
		for _, v := range rawOrig {
			if v.orig == nil {
				continue
			}
			numEls++
		}
		if numEls != numElements {
			b.Fail()
		}
	}
}

func BenchmarkStringMap_ForEach(b *testing.B) {
	const numElements = 20
	rawOrigWithNil := make([]*otlpcommon.StringKeyValue, 2*numElements)
	for i := 0; i < numElements; i++ {
		rawOrigWithNil[i*2] = &otlpcommon.StringKeyValue{
			Key:   "k" + strconv.Itoa(i),
			Value: "v" + strconv.Itoa(i),
		}
	}
	sm := StringMap{
		orig: &rawOrigWithNil,
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		numEls := 0
		sm.ForEach(func(s string, value StringValue) {
			numEls++
		})
		if numEls != numElements {
			b.Fail()
		}
	}
}

func BenchmarkStringMap_RangeOverMap(b *testing.B) {
	const numElements = 20
	rawOrig := make(map[string]StringValue, numElements)
	for i := 0; i < numElements; i++ {
		key := "k" + strconv.Itoa(i)
		rawOrig[key] = StringValue{newStringKeyValue(key, "v"+strconv.Itoa(i))}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		numEls := 0
		for _, v := range rawOrig {
			if v.orig == nil {
				continue
			}
			numEls++
		}
		if numEls != numElements {
			b.Fail()
		}
	}
}

func generateTestStringMap() StringMap {
	sm := NewStringMap()
	fillTestStringMap(sm)
	return sm
}

func fillTestStringMap(dest StringMap) {
	dest.InitFromMap(map[string]string{
		"k": "v",
	})
}

func generateTestAttributeMap() AttributeMap {
	am := NewAttributeMap()
	fillTestAttributeMap(am)
	return am
}

func fillTestAttributeMap(dest AttributeMap) {
	dest.InitFromMap(map[string]AttributeValue{
		"k": NewAttributeValueString("v"),
	})
}

func generateTestNullAttributeMap() AttributeMap {
	am := NewAttributeMap()
	am.InitFromMap(map[string]AttributeValue{
		"k": NewAttributeValueNull(),
	})
	return am
}
func generateTestIntAttributeMap() AttributeMap {
	am := NewAttributeMap()
	am.InitFromMap(map[string]AttributeValue{
		"k": NewAttributeValueInt(123),
	})
	return am
}

func generateTestDoubleAttributeMap() AttributeMap {
	am := NewAttributeMap()
	am.InitFromMap(map[string]AttributeValue{
		"k": NewAttributeValueDouble(12.3),
	})
	return am
}

func generateTestBoolAttributeMap() AttributeMap {
	am := NewAttributeMap()
	am.InitFromMap(map[string]AttributeValue{
		"k": NewAttributeValueBool(true),
	})
	return am
}
