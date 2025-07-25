package mongo

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// Session mongo session
type Session struct {
	client      *mongo.Client
	collection  *mongo.Collection
	table       *Collection
	db          string
	uri         string
	mu          sync.Mutex
	filter      bson.D
	findOpts    []*options.FindOptions
	findOneOpts []*options.FindOneOptions
	stopChan    chan struct{}
}

// C Collection alias
func (s *Session) C(collection string) *Collection {
	return s.Collection(collection)
}

// Collection returns collection
func (s *Session) Collection(collection string) *Collection {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.db) == 0 {
		s.db = "test"
	}
	if s.client == nil {
		return &Collection{}
	}
	d := &Database{database: s.client.Database(s.db)}
	return &Collection{collection: d.database.Collection(collection)}
}

// Connect mongo client
func (s *Session) Connect() error {
	cs, err := connstring.Parse(s.uri)
	if err != nil {
		return err
	}

	timeout := cs.ConnectTimeout
	if timeout == 0 {
		// 连接超时
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.uri))
	if err != nil {
		return err
	}

	// add error handling
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return err
	}

	s.client = client
	s.db = cs.Database
	return nil
}

// Ping verifies that the client can connect to the topology.
// If readPreference is nil then will use the client's default read
// preference.
func (s *Session) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.client.Ping(ctx, readpref.Primary())
}

// Client return mongo Client
func (s *Session) Client() *mongo.Client {
	return s.client
}

// GetCollection return mongo Collection
func (s *Session) GetCollection() *mongo.Collection {
	return s.collection
}

// get table
func (s *Session) GetTable() *Collection {
	return s.table
}

// DB returns a value representing the named database.
func (s *Session) DB(db string) *Database {
	return &Database{database: s.client.Database(db)}
}

// SetOpts set find options
func (s *Session) SetOpts(opts ...interface{}) *Session {
	s.findOpts = make([]*options.FindOptions, 0)
	s.findOneOpts = make([]*options.FindOneOptions, 0)

	for _, opt := range opts {
		switch o := opt.(type) {
		case *options.FindOptions:
			s.findOpts = append(s.findOpts, o)
		case *options.FindOneOptions:
			s.findOneOpts = append(s.findOneOpts, o)
		}
	}
	return s
}

// And adds an AND condition to the filter
func (s *Session) And(filter bson.D) *Session {
	if s.filter == nil {
		s.filter = filter
	} else {
		s.filter = bson.D{{"$and", []bson.D{s.filter, filter}}}
	}
	return s
}

// Or adds an OR condition to the filter
func (s *Session) Or(filter bson.D) *Session {
	if s.filter == nil {
		s.filter = filter
	} else {
		s.filter = bson.D{{"$or", []bson.D{s.filter, filter}}}
	}
	return s
}

// WhereM adds a filter using bson.M
func (s *Session) WhereM(filter bson.M) *Session {
	var d bson.D
	for k, v := range filter {
		d = append(d, bson.E{Key: k, Value: v})
	}
	return s.And(d)
}

// WhereField adds a field=value condition
func (s *Session) WhereField(field string, value interface{}) *Session {
	return s.And(bson.D{{field, value}})
}

// WhereIn adds a field IN condition
func (s *Session) WhereIn(field string, values []interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$in": values}}})
}

// WhereNotIn adds a field NOT IN condition
func (s *Session) WhereNotIn(field string, values []interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$nin": values}}})
}

// WhereGt adds a field > value condition
func (s *Session) WhereGt(field string, value interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$gt": value}}})
}

// WhereGte adds a field >= value condition
func (s *Session) WhereGte(field string, value interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$gte": value}}})
}

// WhereLt adds a field < value condition
func (s *Session) WhereLt(field string, value interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$lt": value}}})
}

// WhereLte adds a field <= value condition
func (s *Session) WhereLte(field string, value interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$lte": value}}})
}

// WhereRegex adds a field regex condition
func (s *Session) WhereRegex(field string, pattern string) *Session {
	return s.And(bson.D{{field, bson.M{"$regex": pattern}}})
}

// WhereExists adds a field exists condition
func (s *Session) WhereExists(field string) *Session {
	return s.And(bson.D{{field, bson.M{"$exists": true}}})
}

// WhereNotExists adds a field not exists condition
func (s *Session) WhereNotExists(field string) *Session {
	return s.And(bson.D{{field, bson.M{"$exists": false}}})
}

// WhereNull adds a field is null condition
func (s *Session) WhereNull(field string) *Session {
	return s.And(bson.D{{field, nil}})
}

// WhereNotNull adds a field is not null condition
func (s *Session) WhereNotNull(field string) *Session {
	return s.And(bson.D{{field, bson.M{"$ne": nil}}})
}

// WhereBetween adds a field between min and max condition
func (s *Session) WhereBetween(field string, min, max interface{}) *Session {
	return s.And(bson.D{{field, bson.M{"$gte": min, "$lte": max}}})
}

// Sort adds sorting to the query
func (s *Session) Sort(field string, order int) *Session {
	if s.findOpts == nil {
		s.findOpts = make([]*options.FindOptions, 0)
	}
	s.findOpts = append(s.findOpts, options.Find().SetSort(bson.D{{field, order}}))
	return s
}

// SortAsc adds ascending sorting to the query
func (s *Session) SortAsc(field string) *Session {
	return s.Sort(field, 1)
}

// SortDesc adds descending sorting to the query
func (s *Session) SortDesc(field string) *Session {
	return s.Sort(field, -1)
}

// Limit adds limit to the query
func (s *Session) Limit(limit int64) *Session {
	if s.findOpts == nil {
		s.findOpts = make([]*options.FindOptions, 0)
	}
	s.findOpts = append(s.findOpts, options.Find().SetLimit(limit))
	return s
}

// Skip adds skip to the query
func (s *Session) Skip(skip int64) *Session {
	if s.findOpts == nil {
		s.findOpts = make([]*options.FindOptions, 0)
	}
	s.findOpts = append(s.findOpts, options.Find().SetSkip(skip))
	return s
}

// Select adds field selection to the query
func (s *Session) Select(fields bson.D) *Session {
	if s.findOpts == nil {
		s.findOpts = make([]*options.FindOptions, 0)
	}
	s.findOpts = append(s.findOpts, options.Find().SetProjection(fields))
	return s
}

// SelectFields adds field selection using field names
func (s *Session) SelectFields(fields ...string) *Session {
	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	return s.Select(projection)
}

// Exclude excludes fields from the query
func (s *Session) Exclude(fields bson.D) *Session {
	if s.findOpts == nil {
		s.findOpts = make([]*options.FindOptions, 0)
	}
	s.findOpts = append(s.findOpts, options.Find().SetProjection(fields))
	return s
}

// ExcludeFields excludes fields using field names
func (s *Session) ExcludeFields(fields ...string) *Session {
	projection := bson.D{}
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 0})
	}
	return s.Exclude(projection)
}

// Find returns up to one document that matches the model.
func (s *Session) Find(result any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Use Find method with limit(1) to ensure sorting/limiting options work
	fo := options.MergeFindOptions(s.findOpts...)
	if fo.Limit == nil || *fo.Limit != 1 {
		fo.SetLimit(1)
	}

	cur, err := s.collection.Find(ctx, s.filter, fo)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	if cur.Next(ctx) {
		return cur.Decode(result)
	}
	return mongo.ErrNoDocuments
}

// FindOne returns up to one document that matches the model (alias for Find)
func (s *Session) FindOne(result any) error {
	return s.Find(result)
}

// First returns the first document that matches the model
func (s *Session) First(result any) error {
	return s.Limit(1).Find(result)
}

// Last returns the last document that matches the model
func (s *Session) Last(result any) error {
	return s.SortDesc("_id").Limit(1).Find(result)
}

// FetchAll find all
func (s *Session) FetchAll(results any) error {
	// 设置超时时间
	ctx := context.Background()
	fo := options.MergeFindOptions(s.findOpts...)
	if fo.NoCursorTimeout == nil || !*fo.NoCursorTimeout {
		maxTime := 10 * time.Second
		if fo.MaxTime != nil {
			maxTime = *fo.MaxTime
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, maxTime)
		defer cancel()
	}

	cur, err := s.collection.Find(ctx, s.filter, s.findOpts...)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return decode(ctx, cur, results)
}

// All find all (alias for FetchAll)
func (s *Session) All(results any) error {
	return s.FetchAll(results)
}

// Get find all (alias for FetchAll)
func (s *Session) Get(results any) error {
	return s.FetchAll(results)
}

// Pluck retrieves a single field from the documents
func (s *Session) Pluck(field string, results any) error {
	return s.SelectFields(field).FetchAll(results)
}

// Exists checks if any documents match the filter
func (s *Session) Exists() bool {
	return s.Count() > 0
}

// DoesntExist checks if no documents match the filter
func (s *Session) DoesntExist() bool {
	return s.Count() == 0
}

// Update by id
func (s *Session) UpdateID(id primitive.ObjectID, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return s.collection.UpdateOne(context.Background(), bson.D{{"_id", id}}, update, opts...)
}

// Update one
func (s *Session) UpdateOne(update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if s.filter == nil {
		s.filter = bson.D{}
	}
	return s.collection.UpdateOne(context.Background(), s.filter, update, opts...)
}

// Update all
func (s *Session) Update(update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if s.filter == nil {
		s.filter = bson.D{}
	}
	return s.collection.UpdateMany(context.Background(), s.filter, update, opts...)
}

// UpdateMany updates multiple documents (alias for Update)
func (s *Session) UpdateMany(update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return s.Update(update, opts...)
}

// Upsert inserts or updates a document
func (s *Session) Upsert(update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	opts = append(opts, options.Update().SetUpsert(true))
	return s.UpdateOne(update, opts...)
}

// Increment increments a field by the given amount
func (s *Session) Increment(field string, amount interface{}) (*mongo.UpdateResult, error) {
	return s.UpdateOne(bson.M{"$inc": bson.M{field: amount}})
}

// Decrement decrements a field by the given amount
func (s *Session) Decrement(field string, amount interface{}) (*mongo.UpdateResult, error) {
	return s.UpdateOne(bson.M{"$inc": bson.M{field: getNegativeValue(amount)}})
}

// Set sets field values
func (s *Session) Set(update bson.M) (*mongo.UpdateResult, error) {
	return s.UpdateOne(bson.M{"$set": update})
}

// Unset removes fields
func (s *Session) Unset(fields ...string) (*mongo.UpdateResult, error) {
	unsetDoc := bson.M{}
	for _, field := range fields {
		unsetDoc[field] = ""
	}
	return s.UpdateOne(bson.M{"$unset": unsetDoc})
}

// Push adds values to an array field
func (s *Session) Push(field string, values ...interface{}) (*mongo.UpdateResult, error) {
	if len(values) == 1 {
		return s.UpdateOne(bson.M{"$push": bson.M{field: values[0]}})
	}
	return s.UpdateOne(bson.M{"$push": bson.M{field: bson.M{"$each": values}}})
}

// Pull removes values from an array field
func (s *Session) Pull(field string, values ...interface{}) (*mongo.UpdateResult, error) {
	if len(values) == 1 {
		return s.UpdateOne(bson.M{"$pull": bson.M{field: values[0]}})
	}
	return s.UpdateOne(bson.M{"$pull": bson.M{field: bson.M{"$in": values}}})
}

// AddToSet adds values to a set (array with unique values)
func (s *Session) AddToSet(field string, values ...interface{}) (*mongo.UpdateResult, error) {
	if len(values) == 1 {
		return s.UpdateOne(bson.M{"$addToSet": bson.M{field: values[0]}})
	}
	return s.UpdateOne(bson.M{"$addToSet": bson.M{field: bson.M{"$each": values}}})
}

// Remove by ID
func (s *Session) RemoveID(id primitive.ObjectID, opts ...*options.DeleteOptions) error {
	s.filter = bson.D{{"_id", id}}
	return s.Remove(opts...)
}

// Remove
func (s *Session) Remove(opts ...*options.DeleteOptions) error {
	if s.filter == nil {
		return errors.New("filter is nil")
	}
	if _, err := s.collection.DeleteMany(context.Background(), s.filter, opts...); err != nil {
		return err
	}
	return nil
}

// Remove one
func (s *Session) RemoveOne(opts ...*options.DeleteOptions) error {
	if s.filter == nil {
		return errors.New("filter is nil")
	}
	if _, err := s.collection.DeleteOne(context.Background(), s.filter, opts...); err != nil {
		return err
	}
	return nil
}

// Delete deletes multiple documents (alias for Remove)
func (s *Session) Delete(opts ...*options.DeleteOptions) error {
	return s.Remove(opts...)
}

// DeleteOne deletes one document (alias for RemoveOne)
func (s *Session) DeleteOne(opts ...*options.DeleteOptions) error {
	return s.RemoveOne(opts...)
}

// DeleteMany deletes multiple documents (alias for Remove)
func (s *Session) DeleteMany(opts ...*options.DeleteOptions) error {
	return s.Remove(opts...)
}

// Count gets the number of documents matching the filter.
func (s *Session) Count(opts ...*options.CountOptions) int64 {
	if s.filter == nil {
		s.filter = bson.D{}
	}
	if v, err := s.collection.CountDocuments(context.Background(), s.filter, opts...); err == nil {
		return v
	}
	return 0
}

// Pagination pagination
func (s *Session) Pagination(page, limit int, results any) (int64, error) {
	fo := options.MergeFindOptions(s.findOpts...)
	if limit > 0 {
		fo.SetLimit(int64(limit))
		offset := (page - 1) * limit
		fo.SetSkip(int64(offset))
	} else {
		fo.SetNoCursorTimeout(true)
	}
	s.SetOpts(fo)
	return s.table.Where(s.filter).Count(), s.FetchAll(results)
}

// Paginate paginates results with page and limit
func (s *Session) Paginate(page, limit int, results any) (int64, error) {
	return s.Pagination(page, limit, results)
}

// Run runs the given model.
func (s *Session) Run(size int32, callback func(*mongo.Cursor)) error {
	ctx := context.Background()

	fo := options.MergeFindOptions(s.findOpts...)
	fo.SetNoCursorTimeout(true)
	fo.SetBatchSize(size)
	s.SetOpts(fo)

	cur, err := s.collection.Find(ctx, s.filter, s.findOpts...)
	if err != nil {
		return err
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		callback(cur)
	}

	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}

// Each iterates over the results and calls the callback for each document
func (s *Session) Each(callback func(result bson.M) error) error {
	return s.Run(100, func(cursor *mongo.Cursor) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error decoding document: %v", err)
			return
		}
		if err := callback(result); err != nil {
			log.Printf("Error in callback: %v", err)
		}
	})
}

// Chunk processes results in chunks
func (s *Session) Chunk(size int, callback func(results []bson.M) error) error {
	var chunk []bson.M
	err := s.Each(func(result bson.M) error {
		chunk = append(chunk, result)
		if len(chunk) >= size {
			if err := callback(chunk); err != nil {
				return err
			}
			chunk = chunk[:0] // Reset chunk
		}
		return nil
	})

	// Process remaining items in chunk
	if len(chunk) > 0 {
		if err := callback(chunk); err != nil {
			return err
		}
	}

	return err
}

// Distinct gets distinct values for a field
func (s *Session) Distinct(field string, opts ...*options.DistinctOptions) ([]interface{}, error) {
	if s.filter == nil {
		s.filter = bson.D{}
	}
	return s.collection.Distinct(context.Background(), field, s.filter, opts...)
}

// StartSession starts a new session for transactions
func (s *Session) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return s.client.StartSession(opts...)
}

// WithTransaction executes a function within a transaction
func (s *Session) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error) {
	session, err := s.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, fn, opts...)
}

// Clone creates a copy of the session
func (s *Session) Clone() *Session {
	return &Session{
		client:      s.client,
		collection:  s.collection,
		table:       s.table,
		db:          s.db,
		uri:         s.uri,
		filter:      s.filter,
		findOpts:    s.findOpts,
		findOneOpts: s.findOneOpts,
	}
}

// Reset resets the session filters and options
func (s *Session) Reset() *Session {
	s.filter = nil
	s.findOpts = nil
	s.findOneOpts = nil
	return s
}

func (s *Session) backgroundCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Ping(); err != nil {
				log.Printf("Ping failed: %v", err)
				if err := s.Connect(); err != nil {
					log.Printf("Reconnect failed: %v", err)
				}
			}
		case <-s.stopChan:
			// Received signal to stop
			return
		}
	}
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		close(s.stopChan)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.client.Disconnect(ctx)
		s.client = nil
	}

	sessionRWMu.Lock()
	delete(sessions, s.uri)
	sessionRWMu.Unlock()
}

// Helper function to get negative value
func getNegativeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
		return -v
	case int32:
		return -v
	case int64:
		return -v
	case float32:
		return -v
	case float64:
		return -v
	default:
		return value
	}
}
