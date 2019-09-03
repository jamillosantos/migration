package migration

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// MongoDBTarget implements the migration.Target of the MongoDB.
//
// In order to get access to the MongoDB, migration.MongoDBTarget uses the MGo
// library (http://github.com/globalsign/mgo).
type MongoDBTarget struct {
	db             *mgo.Database
	collectionName string
}

// mongoDBMigrationVersion represents the version stored on the MongoDB.
type mongoDBMigrationVersion struct {
	ID time.Time `bson:"_id"`
}

// NewMongoDB returns a new instance of the migration.MongoDBTarget
func NewMongoDB(db *mgo.Database) *MongoDBTarget {
	return &MongoDBTarget{
		collectionName: DefaultMigrationTable,
		db:             db,
	}
}

func (t *MongoDBTarget) runWithDB(cb func(db *mgo.Database) error) error {
	return cb(t.db)
}

func (t *MongoDBTarget) collection(db *mgo.Database) *mgo.Collection {
	return db.C(t.collectionName)
}

// Version implements the migration.Target.Version by fetching the current
// version of the database from the collection defined by
// migration.MongoDBTarget.SetCollectionName.
//
// It returns the current version of the database.
//
// Any error returned by the MGo, will be passed up to the caller.
func (t *MongoDBTarget) Version() (migrationID time.Time, err error) {
	err = t.runWithDB(func(db *mgo.Database) error {
		c := t.collection(db)
		var version mongoDBMigrationVersion
		q := c.Find(nil).Sort("-_id").Limit(1) // Most recent
		if err = q.One(&version); err == nil {
			migrationID = version.ID.UTC()
			return nil
		} else if err == mgo.ErrNotFound {
			migrationID = NoVersion
			return nil
		}
		migrationID = NoVersion
		return err
	})
	return
}

// SetVersion implements the migration.Target.SetVersion by storing the passed
// version on the database.
//
// It returns eny error returned by the MGo.
func (t *MongoDBTarget) AddMigration(summary *Summary) error {
	return t.runWithDB(func(db *mgo.Database) error {
		c := t.collection(db)

		if _, err := c.Upsert(
			bson.M{"_id": summary.Migration.GetID()},
			&mongoDBMigrationVersion{
				ID: summary.Migration.GetID(),
			}); err != nil {
			return err
		}
		return nil
	})
}

// RemoveMigration find and removes a migrations from the collection.
func (t *MongoDBTarget) RemoveMigration(summary *Summary) error {
	return t.runWithDB(func(db *mgo.Database) error {
		c := t.collection(db)
		return c.Remove(map[string]interface{}{"_id": summary.Migration.GetID()})
	})
}

func (t *MongoDBTarget) MigrationsExecuted() ([]time.Time, error) {
	migrations := make([]mongoDBMigrationVersion, 0)
	err := t.runWithDB(func(db *mgo.Database) error {
		c := t.collection(db)
		err := c.Find(nil).Sort("_id").All(&migrations)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	r := make([]time.Time, len(migrations))
	for i, migration := range migrations {
		r[i] = migration.ID
	}
	return r, nil
}

// SetCollectionName sets the name of the collection used to store the current
// version of the database.
func (t *MongoDBTarget) SetCollectionName(collection string) *MongoDBTarget {
	t.collectionName = collection
	return t
}

// Database returns the `*mgo.Database` reference of this target.
func (t *MongoDBTarget) Database() *mgo.Database {
	return t.db
}

// BeforeRun implements the BeforeRun interface.
func (t *MongoDBTarget) BeforeRun(executionContext interface{}) {
	if db, ok := executionContext.(*mgo.Database); ok {
		db.Session.ResetIndexCache()
	}
	if session, ok := executionContext.(*mgo.Session); ok {
		session.ResetIndexCache()
	}
}
