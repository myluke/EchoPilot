package mongo

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Table = "persons"

type Person struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Age       int                `bson:"age" json:"age"`
	Email     string             `bson:"email" json:"email"`
	Active    bool               `bson:"active" json:"active"`
	Tags      []string           `bson:"tags" json:"tags"`
	Score     int                `bson:"score" json:"score"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

var session *Session

func init() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	session = New(`mongodb://root:1234@test.localhost:27017/easygram?authsource=admin&connectTimeoutMS=2000&replicaSet=rs0&readPreference=primaryPreferred&maxStalenessSeconds=120`)

	_, err := session.C(Table).Index(
		bson.M{
			"unique": true,
			"keys": bson.D{
				{"name", 1},
			},
		},
		bson.M{
			"keys": bson.D{
				{"created_at", 1},
			},
		},
	)
	if err != nil {
		log.Println(err)
	}
}

func TestPagination(t *testing.T) {
	var err error
	var total int64
	var results []Person
	if total, err = session.C(Table).Where(bson.D{}).SetOpts(options.Find().SetSort(bson.D{{"_id", -1}})).Pagination(2, 2, &results); err != nil {
		log.Println(err)
	}

	log.Printf("total: %v", total)
	for _, r := range results {
		log.Println(r.Name)
	}
}

func TestRun(t *testing.T) {
	var results []Person
	session.C(Table).Where(bson.D{}).SetOpts(options.Find().SetSort(bson.D{{"_id", -1}})).Run(100, func(c *mongo.Cursor) {
		var r Person
		if err := c.Decode(&r); err != nil {
			log.Println(err)
		}
		results = append(results, r)
	})
	for _, r := range results {
		log.Println(r.Name)
	}
}

func TestMain(t *testing.T) {

	defer session.Close()

	// Find find all
	var result []Person
	if err := session.C(Table).Where(bson.D{}).FetchAll(&result); err != nil {
		log.Println(err)
	}

	for _, r := range result {
		log.Println(r.Name)
	}

	// Update one
	if _, err := session.C(Table).Where(bson.D{{"name", "name1"}}).UpdateOne(bson.M{"$set": bson.M{"name": "name01"}}); err != nil {
		log.Println(err)
	}

	// Update update all
	info, err := session.C(Table).Where(bson.D{{"name", "name01"}}).Update(bson.M{"$set": bson.M{"name": "name"}})
	if err != nil {
		log.Println(err)
	}
	log.Printf("%+v", info)

	// Remove one
	if err := session.C(Table).Where(bson.D{{"name", "name"}}).RemoveOne(); err != nil {
		log.Println(err)
	}

	// RemoveAll
	if err := session.C(Table).Where(bson.D{{"name", "name"}}).Remove(); err != nil {
		log.Println(err)
	}

	// Insert
	if _, err := session.C(Table).Insert(bson.M{"name": "name"}); err != nil {
		log.Println(err)
	}

	// InsertAll
	var docs []any
	for index := 0; index < 10; index++ {
		docs = append(docs, bson.M{
			"name":       fmt.Sprintf("name%d", index),
			"created_at": time.Now(),
		})
	}

	if _, err := session.C(Table).InsertAll(docs); err != nil {
		log.Println(err)
	}

	// Count
	count := session.C(Table).Where(bson.D{{"name", "name"}}).Count()
	log.Println(count)
}

// TestNewQueryBuilder 测试新的查询构建器方法
func TestNewQueryBuilder(t *testing.T) {
	collection := session.C(Table)

	// 测试便利的查询方法
	t.Run("WhereField", func(t *testing.T) {
		var results []Person
		err := collection.WhereField("active", true).FetchAll(&results)
		if err != nil {
			t.Errorf("WhereField error: %v", err)
		}
		log.Printf("Found %d active users", len(results))
	})

	t.Run("WhereGt", func(t *testing.T) {
		var results []Person
		err := collection.WhereGt("age", 25).FetchAll(&results)
		if err != nil {
			t.Errorf("WhereGt error: %v", err)
		}
		log.Printf("Found %d users older than 25", len(results))
	})

	t.Run("WhereIn", func(t *testing.T) {
		var results []Person
		err := collection.WhereIn("name", []interface{}{"name1", "name2", "name3"}).FetchAll(&results)
		if err != nil {
			t.Errorf("WhereIn error: %v", err)
		}
		log.Printf("Found %d users with specific names", len(results))
	})

	t.Run("ChainedQueries", func(t *testing.T) {
		var results []Person
		err := collection.WhereField("active", true).
			WhereGte("age", 18).
			WhereLte("age", 65).
			SortDesc("created_at").
			Limit(10).
			FetchAll(&results)
		if err != nil {
			t.Errorf("Chained queries error: %v", err)
		}
		log.Printf("Found %d active adult users", len(results))
	})
}

// TestConvenienceHelpers 测试便利函数
func TestConvenienceHelpers(t *testing.T) {
	collection := session.C(Table)

	t.Run("HelperFunctions", func(t *testing.T) {
		// 使用 D 函数创建过滤器
		filter := D("active", true, "age", Gt(18))

		// 使用 Set 函数创建更新
		update := Set(M("last_login", time.Now()))

		result, err := collection.UpdateMany(filter, update)
		if err != nil {
			t.Errorf("Helper functions error: %v", err)
		}
		log.Printf("Updated %d documents using helper functions", result.ModifiedCount)
	})

	t.Run("ObjectIDHelpers", func(t *testing.T) {
		// 生成新的 ObjectID
		newID := NewObjectID()
		log.Printf("Generated new ObjectID: %v", newID)

		// 检查 ObjectID 是否有效
		if IsValidObjectID(newID.Hex()) {
			log.Println("ObjectID is valid")
		}

		// 从字符串创建 ObjectID
		if objID, err := ObjectID(newID.Hex()); err == nil {
			log.Printf("Created ObjectID from string: %v", objID)
		}
	})
}

// TestBulkOperations 测试批量操作
func TestBulkOperations(t *testing.T) {
	collection := session.C(Table)

	t.Run("BulkBuilder", func(t *testing.T) {
		bulk := collection.NewBulkBuilder()

		// 添加多个操作
		bulk.Insert(M("name", "Alice", "age", 25, "active", true)).
			Insert(M("name", "Bob", "age", 35, "active", true)).
			UpdateOne(D("name", "Alice"), Set(M("status", "active"))).
			DeleteOne(D("name", "Bob"))

		result, err := bulk.Execute()
		if err != nil {
			t.Errorf("Bulk operation error: %v", err)
		}
		log.Printf("Bulk operation: %d inserted, %d modified, %d deleted",
			result.InsertedCount, result.ModifiedCount, result.DeletedCount)
	})

	t.Run("BulkInsert", func(t *testing.T) {
		docs := []interface{}{
			M("name", "User1", "age", 20, "active", true),
			M("name", "User2", "age", 30, "active", false),
			M("name", "User3", "age", 40, "active", true),
		}

		result, err := collection.BulkInsert(docs)
		if err != nil {
			t.Errorf("BulkInsert error: %v", err)
		}
		log.Printf("Bulk inserted %d documents", result.InsertedCount)
	})
}

// TestAggregationOperations 测试聚合操作
func TestAggregationOperations(t *testing.T) {
	collection := session.C(Table)

	t.Run("SimpleAggregation", func(t *testing.T) {
		// 按年龄分组计数
		ageGroups, err := collection.CountByField("age", D("active", true))
		if err != nil {
			t.Errorf("CountByField error: %v", err)
		}
		log.Printf("Age groups: %v", ageGroups)

		// 计算平均年龄
		avgAge, err := collection.Average("age", D("active", true))
		if err != nil {
			t.Errorf("Average error: %v", err)
		}
		log.Printf("Average age: %.2f", avgAge)
	})

	t.Run("AggregationBuilder", func(t *testing.T) {
		pipeline := collection.NewAggregateBuilder().
			Match(M("active", true)).
			Group(M("_id", "$age", "count", M("$sum", 1))).
			Sort(SortBy(Desc("count"))).
			Limit(5)

		var results []bson.M
		err := pipeline.Execute(&results)
		if err != nil {
			t.Errorf("Aggregation builder error: %v", err)
		}
		log.Printf("Top ages: %v", results)
	})
}

// TestIndexManagement 测试索引管理
func TestIndexManagement(t *testing.T) {
	collection := session.C(Table)

	t.Run("IndexOperations", func(t *testing.T) {
		// 确保索引存在
		err := collection.EnsureIndex(D("email", 1), true)
		if err != nil {
			t.Errorf("EnsureIndex error: %v", err)
		}
		log.Println("Email index ensured")

		// 列出所有索引
		indexes, err := collection.ListIndexes()
		if err != nil {
			t.Errorf("ListIndexes error: %v", err)
		}
		log.Printf("Found %d indexes", len(indexes))

		// 创建复合索引
		_, err = collection.CreateIndex(D("name", 1, "age", -1))
		if err != nil {
			log.Printf("Create compound index error: %v", err) // 可能已存在
		}
	})
}

// TestArrayOperations 测试数组操作
func TestArrayOperations(t *testing.T) {
	collection := session.C(Table)

	t.Run("ArrayUpdates", func(t *testing.T) {
		// 先插入一个带数组的文档
		_, err := collection.Insert(M("name", "TestUser", "tags", A("test"), "score", 100))
		if err != nil {
			t.Errorf("Insert error: %v", err)
		}

		// Push 到数组
		_, err = collection.WhereField("name", "TestUser").
			Push("tags", "developer", "golang")
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
		log.Println("Tags pushed successfully")

		// Pull 从数组
		_, err = collection.WhereField("name", "TestUser").
			Pull("tags", "test")
		if err != nil {
			t.Errorf("Pull error: %v", err)
		}
		log.Println("Tag pulled successfully")

		// AddToSet (唯一数组)
		_, err = collection.WhereField("name", "TestUser").
			AddToSet("tags", "golang", "mongodb", "docker")
		if err != nil {
			t.Errorf("AddToSet error: %v", err)
		}
		log.Println("Tags added to set successfully")
	})

	t.Run("IncrementDecrement", func(t *testing.T) {
		// 增加分数
		_, err := collection.WhereField("name", "TestUser").
			Increment("score", 10)
		if err != nil {
			t.Errorf("Increment error: %v", err)
		}
		log.Println("Score incremented successfully")

		// 减少分数
		_, err = collection.WhereField("name", "TestUser").
			Decrement("score", 5)
		if err != nil {
			t.Errorf("Decrement error: %v", err)
		}
		log.Println("Score decremented successfully")
	})
}

// TestUtilityMethods 测试实用方法
func TestUtilityMethods(t *testing.T) {
	collection := session.C(Table)

	t.Run("ExistenceChecks", func(t *testing.T) {
		// 检查文档是否存在
		exists, err := collection.Exists(D("name", "TestUser"))
		if err != nil {
			t.Errorf("Exists error: %v", err)
		}
		log.Printf("TestUser exists: %v", exists)

		// 检查是否不存在
		query := collection.WhereField("name", "NonExistentUser")
		if query.DoesntExist() {
			log.Println("NonExistentUser doesn't exist")
		}
	})

	t.Run("DistinctValues", func(t *testing.T) {
		distinctAges, err := collection.Distinct("age", D("active", true))
		if err != nil {
			t.Errorf("Distinct error: %v", err)
		}
		log.Printf("Distinct ages: %v", distinctAges)
	})

	t.Run("CountOperations", func(t *testing.T) {
		count, err := collection.Count(D("active", true))
		if err != nil {
			t.Errorf("Count error: %v", err)
		}
		log.Printf("Active users count: %d", count)

		totalCount, err := collection.CountAll()
		if err != nil {
			t.Errorf("CountAll error: %v", err)
		}
		log.Printf("Total users count: %d", totalCount)
	})
}

// TestTransactionSupport 测试事务支持
func TestTransactionSupport(t *testing.T) {
	collection := session.C(Table)

	t.Run("WithTransaction", func(t *testing.T) {
		ctx := context.Background()

		result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// 在事务中执行多个操作
			_, err := collection.Insert(M("name", "TransactionUser", "age", 25))
			if err != nil {
				return nil, err
			}

			_, err = collection.UpdateOne(D("name", "TransactionUser"), Set(M("verified", true)))
			if err != nil {
				return nil, err
			}

			return "Transaction completed", nil
		})

		if err != nil {
			t.Errorf("Transaction error: %v", err)
		} else {
			log.Printf("Transaction result: %v", result)
		}
	})
}

// TestPaginationAndChunking 测试分页和分块
func TestPaginationAndChunking(t *testing.T) {
	collection := session.C(Table)

	t.Run("PaginationWithBuilder", func(t *testing.T) {
		var results []Person
		total, err := collection.WhereField("active", true).
			SortDesc("created_at").
			Paginate(1, 5, &results)
		if err != nil {
			t.Errorf("Pagination error: %v", err)
		}
		log.Printf("Page 1 results: %d/%d", len(results), total)
	})

	t.Run("ChunkProcessing", func(t *testing.T) {
		err := collection.WhereField("active", true).
			Chunk(3, func(chunk []bson.M) error {
				log.Printf("Processing chunk of %d documents", len(chunk))
				return nil
			})
		if err != nil {
			t.Errorf("Chunk processing error: %v", err)
		}
	})

	t.Run("EachDocument", func(t *testing.T) {
		count := 0
		err := collection.WhereField("active", true).
			Limit(5).
			Each(func(doc bson.M) error {
				count++
				log.Printf("Processing document %d: %v", count, doc["name"])
				return nil
			})
		if err != nil {
			t.Errorf("Each processing error: %v", err)
		}
		log.Printf("Processed %d documents", count)
	})
}

// TestDatabaseOperations 测试数据库操作
func TestDatabaseOperations(t *testing.T) {
	db := session.DB("testdb")

	t.Run("DatabaseInfo", func(t *testing.T) {
		// 获取数据库统计信息
		stats, err := db.Stats()
		if err != nil {
			t.Errorf("Database stats error: %v", err)
		}
		log.Printf("Database stats: %v", stats["db"])

		// 列出集合
		collections, err := db.CollectionNames()
		if err != nil {
			t.Errorf("CollectionNames error: %v", err)
		}
		log.Printf("Collections: %v", collections)

		// 检查集合是否存在
		hasTable, err := db.HasCollection(Table)
		if err != nil {
			t.Errorf("HasCollection error: %v", err)
		}
		log.Printf("Has %s collection: %v", Table, hasTable)
	})
}

// TestAdvancedQueries 测试高级查询
func TestAdvancedQueries(t *testing.T) {
	collection := session.C(Table)

	t.Run("ComplexFilters", func(t *testing.T) {
		// 复杂的过滤条件
		var results []Person
		err := collection.WhereField("active", true).
			And(D("age", Between(18, 65))).
			Or(D("name", Regex("^Test"))).
			WhereExists("email").
			FetchAll(&results)
		if err != nil {
			t.Errorf("Complex filter error: %v", err)
		}
		log.Printf("Complex query results: %d", len(results))
	})

	t.Run("FieldProjection", func(t *testing.T) {
		var results []bson.M
		err := collection.WhereField("active", true).
			SelectFields("name", "email", "age").
			FetchAll(&results)
		if err != nil {
			t.Errorf("Field projection error: %v", err)
		}
		log.Printf("Projected results: %d", len(results))
	})

	t.Run("PluckValues", func(t *testing.T) {
		var names []string
		err := collection.WhereField("active", true).
			Pluck("name", &names)
		if err != nil {
			t.Errorf("Pluck error: %v", err)
		}
		log.Printf("Plucked names: %v", names)
	})
}

// TestSessionManagement 测试会话管理
func TestSessionManagement(t *testing.T) {
	t.Run("SessionInfo", func(t *testing.T) {
		// 检查连接状态
		if IsConnected(session) {
			log.Println("Session is connected")
		}

		// 获取会话数量
		count := GetSessionCount()
		log.Printf("Active sessions: %d", count)

		// 获取所有会话
		sessions := GetAllSessions()
		log.Printf("Session URIs: %v", func() []string {
			var uris []string
			for uri := range sessions {
				uris = append(uris, uri)
			}
			return uris
		}())
	})
}

// TestCleanup 清理测试数据
func TestCleanup(t *testing.T) {
	collection := session.C(Table)

	t.Run("CleanupTestData", func(t *testing.T) {
		// 删除测试数据
		_, err := collection.DeleteMany(D("name", Regex("^Test")))
		if err != nil {
			t.Errorf("Cleanup error: %v", err)
		}

		_, err = collection.DeleteMany(D("name", Regex("^User")))
		if err != nil {
			t.Errorf("Cleanup error: %v", err)
		}

		_, err = collection.DeleteMany(D("name", In("Alice", "Bob", "TransactionUser")))
		if err != nil {
			t.Errorf("Cleanup error: %v", err)
		}

		log.Println("Test data cleaned up")
	})
}
