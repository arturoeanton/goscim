package scim

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
)

var (
	//Cluster ...
	Cluster             *gocb.Cluster
	FolderBucketSetting = "config/bucketSettings/"
)

type ConfigBucket struct {
	FlushEnabled           *bool   `json:"flush_enabled"`
	ReplicaIndexDisabled   *bool   `json:"replica_index_disabled"`
	RAMQuotaMb             *uint64 `json:"ram_quota_mb"`
	NumReplicas            *uint32 `json:"num_replicas"`
	BucketType             *string `json:"bucket_type"`
	CompressionMode        *string `json:"compression_mode"`
	MaxExpiry              *string `json:"MaxExpiry"`
	EvictionPolicy         *string `json:"EvictionPolicy"`
	ConflictResolutionType *string `json:"conflict_resolution_type"`
}

// InitDB ..
func InitDB() {
	var err error

	username := os.Getenv("SCIM_ADMIN_USER")
	password := os.Getenv("SCIM_ADMIN_PASSWORD")
	endpoint := os.Getenv("SCIM_COUCHBASE_URL")

	if username == "" {
		username = "Administrator"
	}
	if endpoint == "" {
		endpoint = "localhost"
	}
	// Initialize the Connection
	Cluster, err = gocb.Connect("couchbases://"+endpoint, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
		SecurityConfig: gocb.SecurityConfig{
			TLSSkipVerify: true,
		},
	})
	if err != nil {
		log.Fatalln(">>>", err.Error())
		return
	}
}

// CreateBucket ...
func CreateBucket(name string) {
	err := Cluster.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatalln(">>", err.Error())
		return
	}

	defaultConfig := gocb.CreateBucketSettings{
		BucketSettings: gocb.BucketSettings{
			Name:                 name,
			FlushEnabled:         true,
			ReplicaIndexDisabled: true,
			RAMQuotaMB:           200,
			NumReplicas:          0,
			BucketType:           gocb.CouchbaseBucketType,
		},
		ConflictResolutionType: gocb.ConflictResolutionTypeSequenceNumber,
	}

	if _, err := os.Stat(FolderBucketSetting + name + ".json"); err == nil {
		var config ConfigBucket
		file, err := ioutil.ReadFile(FolderBucketSetting + name + ".json")
		if err != nil {
			log.Fatalln(">>", err.Error())
			return
		}
		err = json.Unmarshal(file, &config)
		if err != nil {
			log.Fatalln(">>", err.Error())
			return
		}
		if config.FlushEnabled != nil {
			defaultConfig.FlushEnabled = *config.FlushEnabled
		}
		if config.ReplicaIndexDisabled != nil {
			defaultConfig.ReplicaIndexDisabled = *config.ReplicaIndexDisabled
		}
		if config.RAMQuotaMb != nil {
			defaultConfig.RAMQuotaMB = *config.RAMQuotaMb
		}
		if config.NumReplicas != nil {
			defaultConfig.NumReplicas = *config.NumReplicas
		}
		if config.BucketType != nil {
			if *config.BucketType == "memcached" {
				defaultConfig.BucketType = gocb.MemcachedBucketType
			}
			if *config.BucketType == "couchbase" {
				defaultConfig.BucketType = gocb.CouchbaseBucketType
			}
			if *config.BucketType == "ephemeral" {
				defaultConfig.BucketType = gocb.EphemeralBucketType
			}
		}
		if config.CompressionMode != nil {
			if *config.CompressionMode == "off" {
				defaultConfig.CompressionMode = gocb.CompressionModeOff
			}
			if *config.CompressionMode == "passive" {
				defaultConfig.CompressionMode = gocb.CompressionModePassive
			}
			if *config.CompressionMode == "active" {
				defaultConfig.CompressionMode = gocb.CompressionModeActive
			}
		}
		if config.MaxExpiry != nil {
			defaultConfig.MaxExpiry, err = time.ParseDuration(*config.MaxExpiry)
			if err != nil {
				log.Fatalln(">>", err.Error())
				return
			}
		}
		if config.EvictionPolicy != nil {
			if *config.EvictionPolicy == "fullEviction" {
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeFull
			}
			if *config.EvictionPolicy == "valueOnly" {
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeValueOnly
			}
			if *config.EvictionPolicy == "nruEviction" {
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeNotRecentlyUsed
			}
			if *config.EvictionPolicy == "noEviction" {
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeNoEviction
			}

		}
		if config.ConflictResolutionType != nil {
			if *config.ConflictResolutionType == "lww" {
				defaultConfig.ConflictResolutionType = gocb.ConflictResolutionTypeTimestamp
			}
			if *config.ConflictResolutionType == "seqno" {
				defaultConfig.ConflictResolutionType = gocb.ConflictResolutionTypeSequenceNumber
			}
		}

	}

	_, err = Cluster.Buckets().GetBucket(name, nil)
	if err != nil {
		bucketSettings := defaultConfig
		gocb.SetLogger(gocb.VerboseStdioLogger())
		err = Cluster.Buckets().CreateBucket(bucketSettings, nil)
		if err != nil {
			log.Fatalln(">", err.Error())
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
