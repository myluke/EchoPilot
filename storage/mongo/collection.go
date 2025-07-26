package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection mongo-driver collection
type Collection struct {
	collection *mongo.Collection
}

// get collection
func (c *Collection) Get() *mongo.Collection {
	return c.collection
}

// Index creates an index with the given keys and options.
func (c *Collection) Index(keys ...bson.M) ([]string, error) {
	if c.collection == nil {
		return nil, errors.New("collection is nil")
	}
	ctx := context.Background()

	// 检查是否有唯一索引
	curIndex, err := c.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}

	// 已经存在的索引
	indexes := map[string]bool{}
	defer curIndex.Close(ctx)
	for curIndex.Next(ctx) {
		var index bson.M
		curIndex.Decode(&index)

		keys := []string{}
		for k, v := range index["key"].(bson.M) {
			keys = append(keys, fmt.Sprintf("%v:%v", k, v))
		}

		key := strings.Join(keys, "_")
		if key == "" {
			key = index["name"].(string)
		}
		if _, ok := indexes[key]; !ok {
			indexes[key] = true
		}
	}

	var newIndexes []mongo.IndexModel
	for _, val := range keys {
		// get keys
		keys := []string{}
		if vs, ok := val["keys"].(bson.D); ok {
			for _, v := range vs {
				keys = append(keys, fmt.Sprintf("%v:%v", v.Key, v.Value))
			}
		}
		if vs, ok := val["keys"].(bson.M); ok {
			for k, v := range vs {
				keys = append(keys, fmt.Sprintf("%v:%v", k, v))
			}
		}
		key := strings.Join(keys, "_")
		if _, ok := indexes[key]; ok {
			continue
		}
		opts := options.Index().SetName(key)
		if v, ok := val["unique"]; ok && v.(bool) {
			opts.SetUnique(true)
		}
		if v, ok := val["sparse"]; ok && v.(bool) {
			opts.SetSparse(true)
		}
		if v, ok := val["weights"]; ok {
			opts.SetWeights(v.(bson.M))
		}
		if v, ok := val["language"]; ok {
			opts.SetDefaultLanguage(v.(string))
		}
		newIndexes = append(newIndexes, mongo.IndexModel{
			Keys:    val["keys"],
			Options: opts,
		})
	}

	if len(newIndexes) == 0 {
		return nil, nil
	}
	return c.collection.Indexes().CreateMany(ctx, newIndexes)
}

// DropIndex drops a single index from the collection
func (c *Collection) DropIndex(name string) error {
	if c.collection == nil {
		return errors.New("collection is nil")
	}
	ctx := context.Background()
	_, err := c.collection.Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes lists all indexes in the collection
func (c *Collection) ListIndexes() ([]bson.M, error) {
	if c.collection == nil {
		return nil, errors.New("collection is nil")
	}
	ctx := context.Background()
	cursor, err := c.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}
	return indexes, cursor.Err()
}

// EnsureIndex ensures an index exists with the given keys and options
func (c *Collection) EnsureIndex(keys bson.D, unique bool) error {
	if c.collection == nil {
		return errors.New("collection is nil")
	}

	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	if unique {
		indexModel.Options = options.Index().SetUnique(true)
	}

	ctx := context.Background()
	_, err := c.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// Where finds docs by given filter
func (c *Collection) Where(filter bson.D) *Session {
	return &Session{filter: filter, collection: c.collection, table: c}
}

// WhereM finds docs by given filter (using bson.M)
func (c *Collection) WhereM(filter bson.M) *Session {
	var d bson.D
	for k, v := range filter {
		d = append(d, bson.E{Key: k, Value: v})
	}
	return &Session{filter: d, collection: c.collection, table: c}
}

// WhereID finds docs by id
func (c *Collection) WhereID(id primitive.ObjectID) *Session {
	return c.Where(bson.D{{"_id", id}})
}

// WhereField finds docs by field name and value
func (c *Collection) WhereField(field string, value interface{}) *Session {
	return c.Where(bson.D{{field, value}})
}

// WhereIn finds docs where field is in the given values
func (c *Collection) WhereIn(field string, values []interface{}) *Session {
	return c.Where(bson.D{{field, bson.M{"$in": values}}})
}

// WhereGt finds docs where field is greater than value
func (c *Collection) WhereGt(field string, value interface{}) *Session {
	return c.Where(bson.D{{field, bson.M{"$gt": value}}})
}

// WhereGte finds docs where field is greater than or equal to value
func (c *Collection) WhereGte(field string, value interface{}) *Session {
	return c.Where(bson.D{{field, bson.M{"$gte": value}}})
}

// WhereLt finds docs where field is less than value
func (c *Collection) WhereLt(field string, value interface{}) *Session {
	return c.Where(bson.D{{field, bson.M{"$lt": value}}})
}

// WhereLte finds docs where field is less than or equal to value
func (c *Collection) WhereLte(field string, value interface{}) *Session {
	return c.Where(bson.D{{field, bson.M{"$lte": value}}})
}

// WhereRegex finds docs where field matches regex pattern
func (c *Collection) WhereRegex(field string, pattern string) *Session {
	return c.Where(bson.D{{field, bson.M{"$regex": pattern}}})
}

// WhereExists finds docs where field exists
func (c *Collection) WhereExists(field string) *Session {
	return c.Where(bson.D{{field, bson.M{"$exists": true}}})
}

// WhereNotExists finds docs where field does not exist
func (c *Collection) WhereNotExists(field string) *Session {
	return c.Where(bson.D{{field, bson.M{"$exists": false}}})
}

// FindByID finds a single document by id.
func (c *Collection) FindByID(id primitive.ObjectID, result any) error {
	return c.Where(bson.D{{"_id", id}}).Find(result)
}

// FindOne finds a single document by filter
func (c *Collection) FindOne(filter bson.D, result any) error {
	return c.Where(filter).Find(result)
}

// FindByField finds a single document by field name and value
func (c *Collection) FindByField(field string, value interface{}, result any) error {
	return c.WhereField(field, value).Find(result)
}

// Exists checks if a document exists with the given filter
func (c *Collection) Exists(filter bson.D) (bool, error) {
	count := c.Where(filter).Count()
	return count > 0, nil
}

// ExistsID checks if a document exists with the given id
func (c *Collection) ExistsID(id primitive.ObjectID) (bool, error) {
	return c.Exists(bson.D{{"_id", id}})
}

// ExistsField checks if a document exists with the given field and value
func (c *Collection) ExistsField(field string, value interface{}) (bool, error) {
	return c.Exists(bson.D{{field, value}})
}

// InsertWithResult inserts a single document into the collection and returns insert one result.
func (c *Collection) Insert(doc any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return c.collection.InsertOne(context.Background(), doc, opts...)
}

// InsertAllWithResult inserts the provided documents and returns insert many result.
func (c *Collection) InsertAll(docs []any, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return c.collection.InsertMany(context.Background(), docs, opts...)
}

// InsertOrUpdate inserts a document if it doesn't exist, otherwise updates it
func (c *Collection) InsertOrUpdate(filter bson.D, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	opts = append(opts, options.Update().SetUpsert(true))
	return c.collection.UpdateOne(context.Background(), filter, update, opts...)
}

// UpdateByID updates a single document by id
func (c *Collection) UpdateByID(id primitive.ObjectID, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.collection.UpdateOne(context.Background(), bson.D{{"_id", id}}, update, opts...)
}

// UpdateOne updates a single document by filter
func (c *Collection) UpdateOne(filter bson.D, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.collection.UpdateOne(context.Background(), filter, update, opts...)
}

// UpdateMany updates multiple documents by filter
func (c *Collection) UpdateMany(filter bson.D, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.collection.UpdateMany(context.Background(), filter, update, opts...)
}

// ReplaceOne replaces a single document by filter
func (c *Collection) ReplaceOne(filter bson.D, replacement any, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	return c.collection.ReplaceOne(context.Background(), filter, replacement, opts...)
}

// FindOneAndUpdate finds and updates a single document
func (c *Collection) FindOneAndUpdate(filter bson.D, update any, result any, opts ...*options.FindOneAndUpdateOptions) error {
	singleResult := c.collection.FindOneAndUpdate(context.Background(), filter, update, opts...)
	return singleResult.Decode(result)
}

// FindOneAndReplace finds and replaces a single document
func (c *Collection) FindOneAndReplace(filter bson.D, replacement any, result any, opts ...*options.FindOneAndReplaceOptions) error {
	singleResult := c.collection.FindOneAndReplace(context.Background(), filter, replacement, opts...)
	return singleResult.Decode(result)
}

// FindOneAndDelete finds and deletes a single document
func (c *Collection) FindOneAndDelete(filter bson.D, result any, opts ...*options.FindOneAndDeleteOptions) error {
	singleResult := c.collection.FindOneAndDelete(context.Background(), filter, opts...)
	return singleResult.Decode(result)
}

// DeleteByID deletes a single document by id
func (c *Collection) DeleteByID(id primitive.ObjectID, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(context.Background(), bson.D{{"_id", id}}, opts...)
}

// DeleteOne deletes a single document by filter
func (c *Collection) DeleteOne(filter bson.D, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(context.Background(), filter, opts...)
}

// DeleteMany deletes multiple documents by filter
func (c *Collection) DeleteMany(filter bson.D, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteMany(context.Background(), filter, opts...)
}

// Truncate deletes all documents in the collection
func (c *Collection) Truncate() (*mongo.DeleteResult, error) {
	return c.collection.DeleteMany(context.Background(), bson.D{})
}

// Drop drops the collection
func (c *Collection) Drop() error {
	return c.collection.Drop(context.Background())
}

// Count counts documents by filter
func (c *Collection) Count(filter bson.D, opts ...*options.CountOptions) (int64, error) {
	return c.collection.CountDocuments(context.Background(), filter, opts...)
}

// CountAll counts all documents in the collection
func (c *Collection) CountAll(opts ...*options.CountOptions) (int64, error) {
	return c.collection.CountDocuments(context.Background(), bson.D{}, opts...)
}

// Aggregate performs an aggregation pipeline.
func (c *Collection) Aggregate(pipeline any, results any, opts ...*options.AggregateOptions) error {
	// 设置超时时间
	ao := options.MergeAggregateOptions(opts...)
	maxTime := 10 * time.Second
	if ao.MaxTime != nil {
		maxTime = *ao.MaxTime
	}
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	cur, err := c.collection.Aggregate(ctx, pipeline, ao)
	if err != nil {
		return err
	}
	return decode(ctx, cur, results)
}

// BulkWrite performs multiple write operations
func (c *Collection) BulkWrite(operations []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return c.collection.BulkWrite(context.Background(), operations, opts...)
}

// BulkInsert performs bulk insert operations
func (c *Collection) BulkInsert(documents []interface{}, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	var operations []mongo.WriteModel
	for _, doc := range documents {
		operations = append(operations, mongo.NewInsertOneModel().SetDocument(doc))
	}
	return c.BulkWrite(operations, opts...)
}

// BulkUpdate performs bulk update operations
func (c *Collection) BulkUpdate(updates []BulkUpdateModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	var operations []mongo.WriteModel
	for _, update := range updates {
		model := mongo.NewUpdateOneModel().SetFilter(update.Filter).SetUpdate(update.Update)
		if update.Upsert {
			model.SetUpsert(true)
		}
		operations = append(operations, model)
	}
	return c.BulkWrite(operations, opts...)
}

// BulkUpdateMany performs bulk update many operations
func (c *Collection) BulkUpdateMany(updates []BulkUpdateModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	var operations []mongo.WriteModel
	for _, update := range updates {
		model := mongo.NewUpdateManyModel().SetFilter(update.Filter).SetUpdate(update.Update)
		if update.Upsert {
			model.SetUpsert(true)
		}
		operations = append(operations, model)
	}
	return c.BulkWrite(operations, opts...)
}

// BulkDelete performs bulk delete operations
func (c *Collection) BulkDelete(filters []bson.D, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	var operations []mongo.WriteModel
	for _, filter := range filters {
		operations = append(operations, mongo.NewDeleteOneModel().SetFilter(filter))
	}
	return c.BulkWrite(operations, opts...)
}

// BulkDeleteMany performs bulk delete many operations
func (c *Collection) BulkDeleteMany(filters []bson.D, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	var operations []mongo.WriteModel
	for _, filter := range filters {
		operations = append(operations, mongo.NewDeleteManyModel().SetFilter(filter))
	}
	return c.BulkWrite(operations, opts...)
}

// BulkReplace performs bulk replace operations
func (c *Collection) BulkReplace(replacements []BulkReplaceModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	var operations []mongo.WriteModel
	for _, replacement := range replacements {
		model := mongo.NewReplaceOneModel().SetFilter(replacement.Filter).SetReplacement(replacement.Replacement)
		if replacement.Upsert {
			model.SetUpsert(true)
		}
		operations = append(operations, model)
	}
	return c.BulkWrite(operations, opts...)
}

// BulkUpdateModel represents a bulk update operation
type BulkUpdateModel struct {
	Filter interface{}
	Update interface{}
	Upsert bool
}

// BulkReplaceModel represents a bulk replace operation
type BulkReplaceModel struct {
	Filter      interface{}
	Replacement interface{}
	Upsert      bool
}

// BulkBuilder provides a fluent interface for building bulk operations
type BulkBuilder struct {
	collection *Collection
	operations []mongo.WriteModel
}

// NewBulkBuilder creates a new bulk builder
func (c *Collection) NewBulkBuilder() *BulkBuilder {
	return &BulkBuilder{
		collection: c,
		operations: make([]mongo.WriteModel, 0),
	}
}

// Insert adds an insert operation to the bulk
func (bb *BulkBuilder) Insert(document interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewInsertOneModel().SetDocument(document))
	return bb
}

// UpdateOne adds an update one operation to the bulk
func (bb *BulkBuilder) UpdateOne(filter interface{}, update interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	return bb
}

// UpdateMany adds an update many operation to the bulk
func (bb *BulkBuilder) UpdateMany(filter interface{}, update interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewUpdateManyModel().SetFilter(filter).SetUpdate(update))
	return bb
}

// Upsert adds an upsert operation to the bulk
func (bb *BulkBuilder) Upsert(filter interface{}, update interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	return bb
}

// DeleteOne adds a delete one operation to the bulk
func (bb *BulkBuilder) DeleteOne(filter interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewDeleteOneModel().SetFilter(filter))
	return bb
}

// DeleteMany adds a delete many operation to the bulk
func (bb *BulkBuilder) DeleteMany(filter interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewDeleteManyModel().SetFilter(filter))
	return bb
}

// ReplaceOne adds a replace one operation to the bulk
func (bb *BulkBuilder) ReplaceOne(filter interface{}, replacement interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(replacement))
	return bb
}

// ReplaceUpsert adds a replace upsert operation to the bulk
func (bb *BulkBuilder) ReplaceUpsert(filter interface{}, replacement interface{}) *BulkBuilder {
	bb.operations = append(bb.operations, mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(replacement).SetUpsert(true))
	return bb
}

// Execute executes the bulk operations
func (bb *BulkBuilder) Execute(opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	if len(bb.operations) == 0 {
		return nil, errors.New("no operations to execute")
	}
	return bb.collection.BulkWrite(bb.operations, opts...)
}

// Count returns the number of operations in the bulk
func (bb *BulkBuilder) Count() int {
	return len(bb.operations)
}

// Reset resets the bulk builder
func (bb *BulkBuilder) Reset() *BulkBuilder {
	bb.operations = make([]mongo.WriteModel, 0)
	return bb
}

// GetOperations returns the current operations
func (bb *BulkBuilder) GetOperations() []mongo.WriteModel {
	return bb.operations
}

// CreateIndex creates a single index
func (c *Collection) CreateIndex(keys bson.D, opts ...*options.IndexOptions) (string, error) {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.MergeIndexOptions(opts...),
	}
	return c.collection.Indexes().CreateOne(context.Background(), indexModel)
}

// CreateIndexes creates multiple indexes
func (c *Collection) CreateIndexes(indexes []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	return c.collection.Indexes().CreateMany(context.Background(), indexes, opts...)
}

// Distinct finds distinct values for a field
func (c *Collection) Distinct(field string, filter bson.D, opts ...*options.DistinctOptions) ([]any, error) {
	return c.collection.Distinct(context.Background(), field, filter, opts...)
}

// Clone returns a copy of the collection
func (c *Collection) Clone() (*Collection, error) {
	cloned, err := c.collection.Clone()
	if err != nil {
		return nil, err
	}
	return &Collection{collection: cloned}, nil
}

// Database returns the database this collection belongs to
func (c *Collection) Database() *Database {
	return &Database{database: c.collection.Database()}
}

// Name returns the collection name
func (c *Collection) Name() string {
	return c.collection.Name()
}

// EstimatedDocumentCount returns an estimated count of documents in the collection
func (c *Collection) EstimatedDocumentCount(opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	return c.collection.EstimatedDocumentCount(context.Background(), opts...)
}

// Watch returns a change stream for the collection
func (c *Collection) Watch(pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return c.collection.Watch(context.Background(), pipeline, opts...)
}

// GroupBy performs a group by aggregation
func (c *Collection) GroupBy(groupBy bson.M, having bson.M, results any) error {
	pipeline := []bson.M{
		{"$group": groupBy},
	}

	if having != nil {
		pipeline = append(pipeline, bson.M{"$match": having})
	}

	return c.Aggregate(pipeline, results)
}

// Sum calculates the sum of a field
func (c *Collection) Sum(field string, filter bson.D) (float64, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$" + field},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	total, ok := result[0]["total"].(float64)
	if !ok {
		if intTotal, ok := result[0]["total"].(int32); ok {
			return float64(intTotal), nil
		}
		if longTotal, ok := result[0]["total"].(int64); ok {
			return float64(longTotal), nil
		}
		return 0, nil
	}

	return total, nil
}

// Average calculates the average of a field
func (c *Collection) Average(field string, filter bson.D) (float64, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":     nil,
			"average": bson.M{"$avg": "$" + field},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	avg, ok := result[0]["average"].(float64)
	if !ok {
		return 0, nil
	}

	return avg, nil
}

// Min finds the minimum value of a field
func (c *Collection) Min(field string, filter bson.D) (interface{}, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id": nil,
			"min": bson.M{"$min": "$" + field},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0]["min"], nil
}

// Max finds the maximum value of a field
func (c *Collection) Max(field string, filter bson.D) (interface{}, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id": nil,
			"max": bson.M{"$max": "$" + field},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result[0]["max"], nil
}

// CountByField counts documents grouped by a field
func (c *Collection) CountByField(field string, filter bson.D) ([]bson.M, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":   "$" + field,
			"count": bson.M{"$sum": 1},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SumByField calculates sum grouped by a field
func (c *Collection) SumByField(groupField, sumField string, filter bson.D) ([]bson.M, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":   "$" + groupField,
			"total": bson.M{"$sum": "$" + sumField},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// AverageByField calculates average grouped by a field
func (c *Collection) AverageByField(groupField, avgField string, filter bson.D) ([]bson.M, error) {
	pipeline := []bson.M{}

	if filter != nil {
		pipeline = append(pipeline, bson.M{"$match": filter})
	}

	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":     "$" + groupField,
			"average": bson.M{"$avg": "$" + avgField},
		},
	})

	var result []bson.M
	if err := c.Aggregate(pipeline, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// AggregateBuilder provides a fluent interface for building aggregation pipelines
type AggregateBuilder struct {
	collection *Collection
	pipeline   []bson.M
}

// NewAggregateBuilder creates a new aggregate builder
func (c *Collection) NewAggregateBuilder() *AggregateBuilder {
	return &AggregateBuilder{
		collection: c,
		pipeline:   make([]bson.M, 0),
	}
}

// Match adds a $match stage to the pipeline
func (ab *AggregateBuilder) Match(filter bson.M) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$match": filter})
	return ab
}

// Group adds a $group stage to the pipeline
func (ab *AggregateBuilder) Group(group bson.M) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$group": group})
	return ab
}

// Sort adds a $sort stage to the pipeline
func (ab *AggregateBuilder) Sort(sort bson.D) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$sort": sort})
	return ab
}

// Limit adds a $limit stage to the pipeline
func (ab *AggregateBuilder) Limit(limit int64) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$limit": limit})
	return ab
}

// Skip adds a $skip stage to the pipeline
func (ab *AggregateBuilder) Skip(skip int64) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$skip": skip})
	return ab
}

// Project adds a $project stage to the pipeline
func (ab *AggregateBuilder) Project(project bson.M) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$project": project})
	return ab
}

// Lookup adds a $lookup stage to the pipeline
func (ab *AggregateBuilder) Lookup(from, localField, foreignField, as string) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return ab
}

// Unwind adds an $unwind stage to the pipeline
func (ab *AggregateBuilder) Unwind(path string) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.M{"$unwind": "$" + path})
	return ab
}

// AddStage adds a custom stage to the pipeline
func (ab *AggregateBuilder) AddStage(stage bson.M) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, stage)
	return ab
}

// Execute executes the aggregation pipeline
func (ab *AggregateBuilder) Execute(results any, opts ...*options.AggregateOptions) error {
	return ab.collection.Aggregate(ab.pipeline, results, opts...)
}

// GetPipeline returns the current pipeline
func (ab *AggregateBuilder) GetPipeline() []bson.M {
	return ab.pipeline
}

// Reset resets the pipeline
func (ab *AggregateBuilder) Reset() *AggregateBuilder {
	ab.pipeline = make([]bson.M, 0)
	return ab
}
