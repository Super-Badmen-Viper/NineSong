package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ============== 接口定义 ==============

type BulkWriteResult interface {
	InsertedCount() int64
	MatchedCount() int64
	ModifiedCount() int64
	DeletedCount() int64
	UpsertedCount() int64
	UpsertedIDs() map[int64]interface{}
}

type BulkWrite interface {
	AddModel(models ...BulkModel)
	Execute(ctx context.Context) (BulkWriteResult, error)
}

type BulkModel interface{}

type Database interface {
	Collection(string) Collection
	Client() Client
	ListCollectionNames(ctx context.Context, filter interface{}) ([]string, error)
	CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error
	Raw() interface{}
}

type Collection interface {
	FindOne(context.Context, interface{}) SingleResult
	InsertOne(context.Context, interface{}) (interface{}, error)
	InsertMany(context.Context, []interface{}) ([]interface{}, error)
	DeleteOne(context.Context, interface{}) (int64, error)
	DeleteMany(context.Context, interface{}) (int64, error)
	Find(context.Context, interface{}, ...*options.FindOptions) (Cursor, error)
	CountDocuments(context.Context, interface{}, ...*options.CountOptions) (int64, error)
	Aggregate(context.Context, interface{}) (Cursor, error)
	UpdateOne(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateByID(ctx context.Context, id interface{}, update interface{}) (*mongo.UpdateResult, error)
	Indexes() IndexView
	BulkWrite() BulkWrite
	ListIndexSpecifications(ctx context.Context) ([]*mongo.IndexSpecification, error)
}

type SingleResult interface {
	Decode(interface{}) error
}

type Cursor interface {
	Close(context.Context) error
	Next(context.Context) bool
	Decode(interface{}) error
	All(context.Context, interface{}) error
}

type Client interface {
	Database(string) Database
	Connect(context.Context) error
	Disconnect(context.Context) error
	StartSession() (mongo.Session, error)
	UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error
	Ping(context.Context) error
}

type IndexView interface {
	CreateOne(ctx context.Context, model mongo.IndexModel) (string, error)
	DropAll(ctx context.Context) (bson.Raw, error)
	ListSpecifications(ctx context.Context) ([]*mongo.IndexSpecification, error)
}

// ============== 批量操作实现 ==============
type mongoBulkWrite struct {
	models []mongo.WriteModel
	coll   *mongo.Collection
}

func (mb *mongoBulkWrite) AddModel(models ...BulkModel) {
	for _, model := range models {
		mb.models = append(mb.models, model.(mongo.WriteModel))
	}
}

func (mb *mongoBulkWrite) Execute(ctx context.Context) (BulkWriteResult, error) {
	if len(mb.models) == 0 {
		return nil, errors.New("no operations to execute")
	}
	result, err := mb.coll.BulkWrite(ctx, mb.models)
	if err != nil {
		return nil, err
	}
	return &mongoBulkWriteResult{res: result}, nil
}

type mongoBulkWriteResult struct {
	res *mongo.BulkWriteResult
}

func (m *mongoBulkWriteResult) InsertedCount() int64 { return m.res.InsertedCount }
func (m *mongoBulkWriteResult) MatchedCount() int64  { return m.res.MatchedCount }
func (m *mongoBulkWriteResult) ModifiedCount() int64 { return m.res.ModifiedCount }
func (m *mongoBulkWriteResult) DeletedCount() int64  { return m.res.DeletedCount }
func (m *mongoBulkWriteResult) UpsertedCount() int64 { return m.res.UpsertedCount }
func (m *mongoBulkWriteResult) UpsertedIDs() map[int64]interface{} {
	ids := make(map[int64]interface{})
	for idx, id := range m.res.UpsertedIDs {
		ids[int64(idx)] = id
	}
	return ids
}

// ============== 核心实现 ==============
type mongoClient struct{ cl *mongo.Client }
type mongoDatabase struct{ db *mongo.Database }
type mongoCollection struct{ coll *mongo.Collection }
type mongoSingleResult struct{ sr *mongo.SingleResult }
type mongoCursor struct{ mc *mongo.Cursor }
type mongoSession struct{ mongo.Session }
type mongoIndexView struct{ iv *mongo.IndexView }

func (mc *mongoCollection) BulkWrite() BulkWrite {
	return &mongoBulkWrite{
		coll:   mc.coll,
		models: make([]mongo.WriteModel, 0),
	}
}

func (mc *mongoClient) Ping(ctx context.Context) error {
	return mc.cl.Ping(ctx, readpref.Primary())
}

func (mc *mongoClient) Database(dbName string) Database {
	db := mc.cl.Database(dbName)
	return &mongoDatabase{db: db}
}

func (mc *mongoClient) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return mc.cl.UseSession(ctx, fn)
}

func (mc *mongoClient) StartSession() (mongo.Session, error) {
	session, err := mc.cl.StartSession()
	return &mongoSession{session}, err
}

func (mc *mongoClient) Connect(ctx context.Context) error {
	return mc.cl.Connect(ctx)
}

func (mc *mongoClient) Disconnect(ctx context.Context) error {
	return mc.cl.Disconnect(ctx)
}

func (md *mongoDatabase) Collection(colName string) Collection {
	collection := md.db.Collection(colName)
	return &mongoCollection{coll: collection}
}

func (md *mongoDatabase) Client() Client {
	client := md.db.Client()
	return &mongoClient{cl: client}
}

func (md *mongoDatabase) ListCollectionNames(ctx context.Context, filter interface{}) ([]string, error) {
	return md.db.ListCollectionNames(ctx, filter)
}

func (md *mongoDatabase) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	return md.db.CreateCollection(ctx, name, opts...)
}

func (md *mongoDatabase) Raw() interface{} {
	return md.db
}

func (mc *mongoCollection) FindOne(ctx context.Context, filter interface{}) SingleResult {
	singleResult := mc.coll.FindOne(ctx, filter)
	return &mongoSingleResult{sr: singleResult}
}

func (mc *mongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mc.coll.UpdateOne(ctx, filter, update, opts[:]...)
}

func (mc *mongoCollection) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	id, err := mc.coll.InsertOne(ctx, document)
	return id.InsertedID, err
}

func (mc *mongoCollection) InsertMany(ctx context.Context, document []interface{}) ([]interface{}, error) {
	res, err := mc.coll.InsertMany(ctx, document)
	return res.InsertedIDs, err
}

func (mc *mongoCollection) DeleteOne(ctx context.Context, filter interface{}) (int64, error) {
	count, err := mc.coll.DeleteOne(ctx, filter)
	return count.DeletedCount, err
}

func (mc *mongoCollection) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	count, err := mc.coll.DeleteMany(ctx, filter)
	return count.DeletedCount, err
}

func (mc *mongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	findResult, err := mc.coll.Find(ctx, filter, opts...)
	return &mongoCursor{mc: findResult}, err
}

func (mc *mongoCollection) Aggregate(ctx context.Context, pipeline interface{}) (Cursor, error) {
	aggregateResult, err := mc.coll.Aggregate(ctx, pipeline)
	return &mongoCursor{mc: aggregateResult}, err
}

func (mc *mongoCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mc.coll.UpdateMany(ctx, filter, update, opts[:]...)
}

func (mc *mongoCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return mc.coll.UpdateByID(ctx, id, update)
}

func (mc *mongoCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return mc.coll.CountDocuments(ctx, filter, opts...)
}

func (mc *mongoCollection) Indexes() IndexView {
	indexView := mc.coll.Indexes()
	return &mongoIndexView{iv: &indexView}
}

func (mc *mongoCollection) ListIndexSpecifications(ctx context.Context) ([]*mongo.IndexSpecification, error) {
	return mc.coll.Indexes().ListSpecifications(ctx)
}

func (sr *mongoSingleResult) Decode(v interface{}) error {
	return sr.sr.Decode(v)
}

func (mr *mongoCursor) Close(ctx context.Context) error {
	return mr.mc.Close(ctx)
}

func (mr *mongoCursor) Next(ctx context.Context) bool {
	return mr.mc.Next(ctx)
}

func (mr *mongoCursor) Decode(v interface{}) error {
	return mr.mc.Decode(v)
}

func (mr *mongoCursor) All(ctx context.Context, result interface{}) error {
	return mr.mc.All(ctx, result)
}

func (miv *mongoIndexView) CreateOne(ctx context.Context, model mongo.IndexModel) (string, error) {
	return miv.iv.CreateOne(ctx, model)
}

func (miv *mongoIndexView) DropAll(ctx context.Context) (bson.Raw, error) {
	return miv.iv.DropAll(ctx)
}

func (miv *mongoIndexView) ListSpecifications(ctx context.Context) ([]*mongo.IndexSpecification, error) {
	return (*miv.iv).ListSpecifications(ctx)
}

// ============== 客户端初始化 ==============

func NewClient(connection string) (Client, error) {
	time.Local = time.UTC
	c, err := mongo.NewClient(options.Client().ApplyURI(connection))
	return &mongoClient{cl: c}, err
}
