package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Database mongo-driver database
type Database struct {
	database *mongo.Database
}

// get database
func (d *Database) Get() *mongo.Database {
	return d.database
}

// CollectionNames returns the collection names present in database.
func (d *Database) CollectionNames() ([]string, error) {
	return d.database.ListCollectionNames(context.Background(), bson.D{})
}

// ListCollections returns the collection specifications present in database.
func (d *Database) ListCollections() ([]CollectionSpecification, error) {
	cursor, err := d.database.ListCollections(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var collections []CollectionSpecification
	for cursor.Next(context.Background()) {
		var collection CollectionSpecification
		if err := cursor.Decode(&collection); err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}
	return collections, cursor.Err()
}

// CreateCollection creates a new collection in the database.
func (d *Database) CreateCollection(name string, opts ...*options.CreateCollectionOptions) error {
	return d.database.CreateCollection(context.Background(), name, opts...)
}

// DropCollection drops a collection from the database.
func (d *Database) DropCollection(name string) error {
	return d.database.Collection(name).Drop(context.Background())
}

// HasCollection checks if a collection exists in the database.
func (d *Database) HasCollection(name string) (bool, error) {
	names, err := d.CollectionNames()
	if err != nil {
		return false, err
	}

	for _, n := range names {
		if n == name {
			return true, nil
		}
	}
	return false, nil
}

// Stats returns statistics about the database.
func (d *Database) Stats() (bson.M, error) {
	var result bson.M
	err := d.database.RunCommand(context.Background(), bson.D{{"dbStats", 1}}).Decode(&result)
	return result, err
}

// Drop drops the database.
func (d *Database) Drop() error {
	return d.database.Drop(context.Background())
}

// Name returns the database name.
func (d *Database) Name() string {
	return d.database.Name()
}

// C returns collection.
func (d *Database) C(collection string) *Collection {
	return d.Collection(collection)
}

// Collection returns collection.
func (d *Database) Collection(collection string) *Collection {
	return &Collection{collection: d.database.Collection(collection)}
}

// RunCommand runs a command against the database.
func (d *Database) RunCommand(command interface{}, opts ...*options.RunCmdOptions) *mongo.SingleResult {
	return d.database.RunCommand(context.Background(), command, opts...)
}

// Aggregate runs an aggregation pipeline against the database.
func (d *Database) Aggregate(pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return d.database.Aggregate(context.Background(), pipeline, opts...)
}

// Watch returns a change stream cursor used to receive notifications of changes to the database.
func (d *Database) Watch(pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return d.database.Watch(context.Background(), pipeline, opts...)
}

// Client returns the client used to create this database.
func (d *Database) Client() *mongo.Client {
	return d.database.Client()
}

// ReadConcern returns the read concern used to create this database.
func (d *Database) ReadConcern() *readconcern.ReadConcern {
	return d.database.ReadConcern()
}

// ReadPreference returns the read preference used to create this database.
func (d *Database) ReadPreference() *readpref.ReadPref {
	return d.database.ReadPreference()
}

// WriteConcern returns the write concern used to create this database.
func (d *Database) WriteConcern() *writeconcern.WriteConcern {
	return d.database.WriteConcern()
}
