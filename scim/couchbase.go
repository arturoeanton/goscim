package scim

import (
	"log"
	"os"
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

	username := os.Getenv("SCIM_ADMIN_USER")
	password := os.Getenv("SCIM_ADMIN_PASSWORD")
	urlCouchbase := os.Getenv("SCIM_COUCHBASE_URL")
	if username == "" {
		username = "Administrator"
	}
	if urlCouchbase == "" {
		urlCouchbase = "localhost"
	}

	Cluster, err = gocb.Connect(
		urlCouchbase,
		gocb.ClusterOptions{
			Username: username,
			Password: password,
		})
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = Cluster.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Println("check var SCIM_COUCHBASE_URL:" + urlCouchbase)
		log.Println("check var SCIM_ADMIN_USER:" + username)
		log.Println("check var SCIM_ADMIN_PASSWORD:" + hidepassword(password))
		log.Println("Error waiting for Couchbase cluster to be ready")
		log.Fatalln(err.Error())
		return
	}
}

func hidepassword(password string) string {
	if len(password) == 0 {
		return "?"
	}
	return password[:1] + "*" + password[len(password)-1:]
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
