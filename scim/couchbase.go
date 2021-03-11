package scim

import (
	"log"
	"time"

	"github.com/couchbase/gocb/v2"
)

var (
	//Cluster ...
	Cluster *gocb.Cluster
)

// InitDB ..
func InitDB() {
	var err error
	Cluster, err = gocb.Connect(
		"localhost",
		gocb.ClusterOptions{
			Username: "Administrator",
			Password: "admin!!",
		})
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}

// CreateBucket ...
func CreateBucket(name string) {
	_, err := Cluster.Buckets().GetBucket(name, nil)
	if err != nil {
		bucketSettings := gocb.CreateBucketSettings{
			BucketSettings: gocb.BucketSettings{
				Name:                 name,
				FlushEnabled:         false,
				ReplicaIndexDisabled: true,
				RAMQuotaMB:           200,
				NumReplicas:          1,
				BucketType:           gocb.CouchbaseBucketType,
			},
			ConflictResolutionType: gocb.ConflictResolutionTypeSequenceNumber,
		}
		err = Cluster.Buckets().CreateBucket(bucketSettings, nil)
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Println("Create Bucket -> " + name)
	}
	bucket := Cluster.Bucket(name)
	err = bucket.WaitUntilReady(20*time.Second, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Ready Bucket -> " + name)
	Cluster.QueryIndexes().CreatePrimaryIndex(name, &gocb.CreatePrimaryQueryIndexOptions{IgnoreIfExists: true})
}

/*

couchbase-cli enable-developer-preview -c localhost:8091 -u Administrator -p 'admin!!'  --enable

couchbase-cli collection-manage \
--cluster http://localhost:8091 \
--username Administrator \
--password 'admin!!' \
--bucket bucket_scim \
--create-scope SCIM

cbstats -u Administrator -p 'admin!!'  localhost:11210 -b bucket_scim all | grep collections

couchbase-cli collection-manage \
-c localhost \
--username Administrator \
--password 'admin!!' \
--bucket bucket_scim \
--list-scopes

/opt/couchbase/bin/couchbase-cli collection-manage -c localhost \
--username Administrator \
--password 'admin!!' \
--bucket bucket_scim \
--create-collection SCIM.Element
*/
