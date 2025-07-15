package mongo

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/mylukin/EchoPilot/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrNoDocuments is mongo: no document results
var ErrNoDocuments = mongo.ErrNoDocuments

var (
	sessionRWMu sync.RWMutex
	sessions    = make(map[string]*Session)
)

type (
	BulkWriteResult         = mongo.BulkWriteResult
	InsertOneResult         = mongo.InsertOneResult
	InsertManyResult        = mongo.InsertManyResult
	DeleteResult            = mongo.DeleteResult
	RewrapManyDataKeyResult = mongo.RewrapManyDataKeyResult
	ListDatabasesResult     = mongo.ListDatabasesResult
	DatabaseSpecification   = mongo.DatabaseSpecification
	UpdateResult            = mongo.UpdateResult
	IndexSpecification      = mongo.IndexSpecification
	CollectionSpecification = mongo.CollectionSpecification
	Pipeline                = mongo.Pipeline
	SessionContext          = mongo.SessionContext
	Cursor                  = mongo.Cursor
	Client                  = mongo.Client
)

// New session
//
// Relevant documentation:
//
//	https://docs.mongodb.com/manual/reference/connection-string/
func New(uri ...string) *Session {
	URI := helper.Config("MONGO_URI")
	if len(uri) > 0 {
		URI = uri[0]
	}
	if URI == "" {
		return &Session{}
	}
	session, err := Get(URI)
	if err != nil {
		log.Panic(err)
	}

	return session
}

// Close session
func Close(uri ...string) error {
	URI := helper.Config("MONGO_URI")
	if len(uri) > 0 {
		URI = uri[0]
	}
	session, err := Get(URI)
	if err != nil {
		return fmt.Errorf("failed to get session for URI %s: %v", uri, err)
	}
	session.Close()
	return nil
}

func Get(uri string) (*Session, error) {
	sessionRWMu.RLock()
	if s, exists := sessions[uri]; exists {
		sessionRWMu.RUnlock()
		return s, nil
	}
	sessionRWMu.RUnlock()

	sessionRWMu.Lock()
	defer sessionRWMu.Unlock()

	// Double-check after acquiring the write lock
	if s, exists := sessions[uri]; exists {
		return s, nil
	}

	s := &Session{
		uri:      uri,
		stopChan: make(chan struct{}),
	}
	if err := s.Connect(); err != nil {
		return nil, err
	}

	sessions[uri] = s
	go s.backgroundCheck()

	return s, nil
}

// C Collection alias
func C(collection string, uri ...string) *Collection {
	return New(uri...).Collection(collection)
}

// decode
func decode(ctx context.Context, cur *mongo.Cursor, results any) error {
	resultsVal := reflect.ValueOf(results)
	if resultsVal.Kind() != reflect.Ptr {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a %s", resultsVal.Kind())
	}

	sliceVal := resultsVal.Elem()
	if sliceVal.Kind() == reflect.Interface {
		sliceVal = sliceVal.Elem()
	}

	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a pointer to %s", sliceVal.Kind())
	}

	elementType := sliceVal.Type().Elem()

	defer cur.Close(ctx)
	var index int
	for cur.Next(ctx) {
		data := reflect.New(elementType)
		if err := cur.Decode(data.Interface()); err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, data.Elem()))
		index++
	}

	if err := cur.Err(); err != nil {
		return err
	}

	resultsVal.Elem().Set(sliceVal.Slice(0, index))
	return nil
}

// ObjectID creates a new ObjectID from a hex string
func ObjectID(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}

// IsValidObjectID checks if a string is a valid ObjectID
func IsValidObjectID(hex string) bool {
	return primitive.IsValidObjectID(hex)
}

// NewObjectID generates a new ObjectID
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// RegexOptions creates a regex with options
func RegexOptions(pattern, options string) bson.M {
	return bson.M{"$regex": pattern, "$options": options}
}

// Regex creates a regex pattern
func Regex(pattern string) bson.M {
	return bson.M{"$regex": pattern}
}

// In creates an $in condition
func In(values ...interface{}) bson.M {
	return bson.M{"$in": values}
}

// NotIn creates a $nin condition
func NotIn(values ...interface{}) bson.M {
	return bson.M{"$nin": values}
}

// Gt creates a $gt condition
func Gt(value interface{}) bson.M {
	return bson.M{"$gt": value}
}

// Gte creates a $gte condition
func Gte(value interface{}) bson.M {
	return bson.M{"$gte": value}
}

// Lt creates a $lt condition
func Lt(value interface{}) bson.M {
	return bson.M{"$lt": value}
}

// Lte creates a $lte condition
func Lte(value interface{}) bson.M {
	return bson.M{"$lte": value}
}

// Ne creates a $ne condition
func Ne(value interface{}) bson.M {
	return bson.M{"$ne": value}
}

// Exists creates an $exists condition
func Exists(exists bool) bson.M {
	return bson.M{"$exists": exists}
}

// Between creates a range condition
func Between(min, max interface{}) bson.M {
	return bson.M{"$gte": min, "$lte": max}
}

// Or creates an $or condition
func Or(conditions ...bson.D) bson.M {
	return bson.M{"$or": conditions}
}

// And creates an $and condition
func And(conditions ...bson.D) bson.M {
	return bson.M{"$and": conditions}
}

// Not creates a $not condition
func Not(condition bson.M) bson.M {
	return bson.M{"$not": condition}
}

// Nor creates a $nor condition
func Nor(conditions ...bson.D) bson.M {
	return bson.M{"$nor": conditions}
}

// Size creates a $size condition
func Size(size int) bson.M {
	return bson.M{"$size": size}
}

// All creates an $all condition
func All(values ...interface{}) bson.M {
	return bson.M{"$all": values}
}

// ElemMatch creates an $elemMatch condition
func ElemMatch(condition bson.M) bson.M {
	return bson.M{"$elemMatch": condition}
}

// Set creates a $set update
func Set(updates bson.M) bson.M {
	return bson.M{"$set": updates}
}

// Unset creates an $unset update
func Unset(fields ...string) bson.M {
	unsetDoc := bson.M{}
	for _, field := range fields {
		unsetDoc[field] = ""
	}
	return bson.M{"$unset": unsetDoc}
}

// Inc creates an $inc update
func Inc(field string, amount interface{}) bson.M {
	return bson.M{"$inc": bson.M{field: amount}}
}

// Push creates a $push update
func Push(field string, values ...interface{}) bson.M {
	if len(values) == 1 {
		return bson.M{"$push": bson.M{field: values[0]}}
	}
	return bson.M{"$push": bson.M{field: bson.M{"$each": values}}}
}

// Pull creates a $pull update
func Pull(field string, values ...interface{}) bson.M {
	if len(values) == 1 {
		return bson.M{"$pull": bson.M{field: values[0]}}
	}
	return bson.M{"$pull": bson.M{field: bson.M{"$in": values}}}
}

// AddToSet creates an $addToSet update
func AddToSet(field string, values ...interface{}) bson.M {
	if len(values) == 1 {
		return bson.M{"$addToSet": bson.M{field: values[0]}}
	}
	return bson.M{"$addToSet": bson.M{field: bson.M{"$each": values}}}
}

// Min creates a $min update
func Min(field string, value interface{}) bson.M {
	return bson.M{"$min": bson.M{field: value}}
}

// Max creates a $max update
func Max(field string, value interface{}) bson.M {
	return bson.M{"$max": bson.M{field: value}}
}

// Mul creates a $mul update
func Mul(field string, value interface{}) bson.M {
	return bson.M{"$mul": bson.M{field: value}}
}

// Rename creates a $rename update
func Rename(oldField, newField string) bson.M {
	return bson.M{"$rename": bson.M{oldField: newField}}
}

// CurrentDate creates a $currentDate update
func CurrentDate(field string) bson.M {
	return bson.M{"$currentDate": bson.M{field: true}}
}

// Asc creates ascending sort
func Asc(field string) bson.E {
	return bson.E{Key: field, Value: 1}
}

// Desc creates descending sort
func Desc(field string) bson.E {
	return bson.E{Key: field, Value: -1}
}

// SortBy creates a sort document
func SortBy(sorts ...bson.E) bson.D {
	var sortDoc bson.D
	for _, sort := range sorts {
		sortDoc = append(sortDoc, sort)
	}
	return sortDoc
}

// Project creates a projection document
func Project(fields ...string) bson.D {
	var projection bson.D
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	return projection
}

// Exclude creates an exclusion projection document
func Exclude(fields ...string) bson.D {
	var projection bson.D
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 0})
	}
	return projection
}

// M creates a bson.M from key-value pairs
func M(pairs ...interface{}) bson.M {
	if len(pairs)%2 != 0 {
		panic("pairs must be even")
	}

	result := bson.M{}
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic("key must be string")
		}
		result[key] = pairs[i+1]
	}
	return result
}

// D creates a bson.D from key-value pairs
func D(pairs ...interface{}) bson.D {
	if len(pairs)%2 != 0 {
		panic("pairs must be even")
	}

	var result bson.D
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic("key must be string")
		}
		result = append(result, bson.E{Key: key, Value: pairs[i+1]})
	}
	return result
}

// A creates a bson.A from values
func A(values ...interface{}) bson.A {
	return bson.A(values)
}

// IsConnected checks if a session is connected
func IsConnected(session *Session) bool {
	return session != nil && session.client != nil && session.Ping() == nil
}

// GetAllSessions returns all active sessions
func GetAllSessions() map[string]*Session {
	sessionRWMu.RLock()
	defer sessionRWMu.RUnlock()

	result := make(map[string]*Session)
	for uri, session := range sessions {
		result[uri] = session
	}
	return result
}

// CloseAllSessions closes all active sessions
func CloseAllSessions() {
	sessionRWMu.Lock()
	defer sessionRWMu.Unlock()

	for _, session := range sessions {
		session.Close()
	}
}

// GetSessionCount returns the number of active sessions
func GetSessionCount() int {
	sessionRWMu.RLock()
	defer sessionRWMu.RUnlock()
	return len(sessions)
}
