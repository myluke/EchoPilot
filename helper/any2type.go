package helper

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToInt64 is interface to int64
func ToInt64(v interface{}) int64 {
	var r int64
	switch _v := v.(type) {
	case string:
		r, _ = strconv.ParseInt(_v, 10, 64)
	default:
		r = _v.(int64)
	}
	return r
}

// ToFloat64 is interface to float64
func ToFloat64(v interface{}) float64 {
	var r float64
	switch _v := v.(type) {
	case string:
		r, _ = strconv.ParseFloat(_v, 64)
	default:
		r = _v.(float64)
	}
	return r
}

// ToObjectID is interface to ObjectID
func ToObjectID(v interface{}) primitive.ObjectID {
	var r primitive.ObjectID
	switch _v := v.(type) {
	case string:
		r, _ = primitive.ObjectIDFromHex(_v)
	case primitive.ObjectID:
		r = _v
	}
	return r
}

// ToString is interface to string
func ToString(v interface{}) string {
	var r string
	switch _v := v.(type) {
	case float32:
		r = strconv.FormatFloat(float64(_v), 'f', -1, 64)
	case float64:
		r = strconv.FormatFloat(_v, 'f', -1, 64)
	case int:
		r = strconv.Itoa(_v)
	case int64:
		r = strconv.FormatInt(_v, 10)
	default:
		r = _v.(string)
	}
	return r
}
