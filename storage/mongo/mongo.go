package mongo

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/mylukin/EchoPilot/helper"
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
	session, err := Get(URI)
	if err != nil {
		log.Panic(err)
	}

	return session
}

func Get(uri string) (*Session, error) {
	// First, try to read the session with a read lock
	sessionRWMu.RLock()
	if s, exists := sessions[uri]; exists {
		s.mu.Lock()
		s.refCount++
		s.mu.Unlock()
		sessionRWMu.RUnlock()
		return s, nil
	}
	sessionRWMu.RUnlock()

	// If the session doesn't exist, acquire a write lock
	sessionRWMu.Lock()
	defer sessionRWMu.Unlock()

	// Double-check if the session was created while we were waiting for the write lock
	if s, exists := sessions[uri]; exists {
		s.mu.Lock()
		s.refCount++
		s.mu.Unlock()
		return s, nil
	}

	// Create a new session
	s := &Session{
		uri:      uri,
		stopChan: make(chan struct{}),
		refCount: 1,
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
