package scim

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocbcore/v10/memd"
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

// CouchbaseStore is the production Store, backed by a Couchbase cluster with
// one bucket per resource type.
type CouchbaseStore struct {
	cluster *gocb.Cluster
}

// NewCouchbaseStore wraps an already-connected cluster.
func NewCouchbaseStore(cluster *gocb.Cluster) *CouchbaseStore {
	return &CouchbaseStore{cluster: cluster}
}

// InitDB connects to Couchbase and installs a CouchbaseStore as the active DB.
func InitDB() error {
	username := os.Getenv("SCIM_ADMIN_USER")
	password := os.Getenv("SCIM_ADMIN_PASSWORD")
	endpoint := os.Getenv("SCIM_COUCHBASE_URL")

	if username == "" {
		username = "Administrator"
	}
	if endpoint == "" {
		endpoint = "localhost"
	}

	connectionString, security, err := couchbaseConnection(endpoint)
	if err != nil {
		return err
	}
	if security.TLSSkipVerify {
		log.Println("WARNING: SCIM_COUCHBASE_TLS_SKIP_VERIFY is set - the Couchbase certificate is not verified")
	}

	// Initialize the Connection
	cluster, err := gocb.Connect(connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
		SecurityConfig: security,
	})
	if err != nil {
		return err
	}
	if err := cluster.WaitUntilReady(5*time.Second, nil); err != nil {
		log.Println("check var SCIM_COUCHBASE_URL:" + endpoint)
		log.Println("check var SCIM_ADMIN_USER:" + username)
		log.Println("check var SCIM_ADMIN_PASSWORD:" + hidepassword(password))
		log.Println("Error waiting for Couchbase cluster to be ready")
		return err
	}

	Cluster = cluster
	DB = NewCouchbaseStore(cluster)
	return nil
}

func hidepassword(password string) string {
	if len(password) == 0 {
		return "?"
	}
	return password[:1] + "*" + password[len(password)-1:]
}

// Close releases the cluster connection.
func (s *CouchbaseStore) Close() error {
	return s.cluster.Close(nil)
}

func (s *CouchbaseStore) collection(bucket string) *gocb.Collection {
	return s.cluster.Bucket(bucket).DefaultCollection()
}

// isKeyNotFound reports whether err is Couchbase's "document missing" error.
func isKeyNotFound(err error) bool {
	var kvErr *gocb.KeyValueError
	if errors.As(err, &kvErr) {
		return kvErr.StatusCode == memd.StatusKeyNotFound
	}
	return errors.Is(err, gocb.ErrDocumentNotFound)
}

// Get implements Store.
func (s *CouchbaseStore) Get(bucket, id string) (map[string]interface{}, error) {
	data, err := s.collection(bucket).Get(id, &gocb.GetOptions{})
	if err != nil {
		if isKeyNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var doc map[string]interface{}
	if err := data.Content(&doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Upsert implements Store.
func (s *CouchbaseStore) Upsert(bucket, id string, doc map[string]interface{}) error {
	_, err := s.collection(bucket).Upsert(id, doc, &gocb.UpsertOptions{})
	return err
}

// Replace implements Store.
func (s *CouchbaseStore) Replace(bucket, id string, doc map[string]interface{}) error {
	_, err := s.collection(bucket).Replace(id, doc, &gocb.ReplaceOptions{})
	if err != nil && isKeyNotFound(err) {
		return ErrNotFound
	}
	return err
}

// Remove implements Store.
func (s *CouchbaseStore) Remove(bucket, id string) error {
	_, err := s.collection(bucket).Remove(id, &gocb.RemoveOptions{})
	if err != nil && isKeyNotFound(err) {
		return ErrNotFound
	}
	return err
}

// Search implements Store by translating the SCIM filter to N1QL.
func (s *CouchbaseStore) Search(q SearchQuery) (int, []map[string]interface{}, error) {
	queryPage, queryCount, err := parser.FilterToN1QL(q.Bucket, q.Filter)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %s", ErrInvalidFilter, err)
	}

	sortBy := q.SortBy
	if sortBy == "" {
		sortBy = "id"
	} else {
		cache := make([]string, 0)
		for _, s := range strings.Split(sortBy, ",") {
			cache = append(cache, parser.AddQuote(s))
		}
		sortBy = strings.Join(cache, ",")
	}
	sortBy = strings.Trim(sortBy, " ")
	sortBy = strings.ReplaceAll(sortBy, ";", "")

	sortOrder := "ASC"
	if q.SortDescending {
		sortOrder = "DESC"
	}

	queryPage += "\nORDER BY " + sortBy + " " + sortOrder
	queryPage += "\nOFFSET " + strconv.Itoa(q.Offset)
	queryPage += "\nLIMIT " + strconv.Itoa(q.Limit)

	options := &gocb.QueryOptions{ScanConsistency: queryScanConsistency()}

	rowsCount, err := s.cluster.Query(queryCount, options)
	if err != nil {
		return 0, nil, err
	}
	defer rowsCount.Close()
	var countResult struct {
		Count int
	}
	if err := rowsCount.One(&countResult); err != nil {
		return 0, nil, err
	}

	rows, err := s.cluster.Query(queryPage, options)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	resources := make([]map[string]interface{}, 0)
	for rows.Next() {
		var item Resource
		if err := rows.Row(&item); err != nil {
			return 0, nil, err
		}
		// "SELECT * FROM `Bucket`" nests every row under the bucket name.
		doc, ok := item[q.Bucket].(map[string]interface{})
		if !ok {
			continue
		}
		resources = append(resources, doc)
	}
	return countResult.Count, resources, rows.Err()
}

// EnsureBucket implements Store, creating the bucket and its primary index.
func (s *CouchbaseStore) EnsureBucket(name string) error {
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
		file, err := os.ReadFile(FolderBucketSetting + name + ".json")
		if err != nil {
			return err
		}
		if err := json.Unmarshal(file, &config); err != nil {
			return err
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
			switch *config.BucketType {
			case "memcached":
				defaultConfig.BucketType = gocb.MemcachedBucketType
			case "couchbase":
				defaultConfig.BucketType = gocb.CouchbaseBucketType
			case "ephemeral":
				defaultConfig.BucketType = gocb.EphemeralBucketType
			}
		}
		if config.CompressionMode != nil {
			switch *config.CompressionMode {
			case "off":
				defaultConfig.CompressionMode = gocb.CompressionModeOff
			case "passive":
				defaultConfig.CompressionMode = gocb.CompressionModePassive
			case "active":
				defaultConfig.CompressionMode = gocb.CompressionModeActive
			}
		}
		if config.MaxExpiry != nil {
			defaultConfig.MaxExpiry, err = time.ParseDuration(*config.MaxExpiry)
			if err != nil {
				return err
			}
		}
		if config.EvictionPolicy != nil {
			switch *config.EvictionPolicy {
			case "fullEviction":
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeFull
			case "valueOnly":
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeValueOnly
			case "nruEviction":
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeNotRecentlyUsed
			case "noEviction":
				defaultConfig.EvictionPolicy = gocb.EvictionPolicyTypeNoEviction
			}
		}
		if config.ConflictResolutionType != nil {
			switch *config.ConflictResolutionType {
			case "lww":
				defaultConfig.ConflictResolutionType = gocb.ConflictResolutionTypeTimestamp
			case "seqno":
				defaultConfig.ConflictResolutionType = gocb.ConflictResolutionTypeSequenceNumber
			}
		}
	}

	if _, err := s.cluster.Buckets().GetBucket(name, nil); err != nil {
		if err := s.cluster.Buckets().CreateBucket(defaultConfig, nil); err != nil {
			return fmt.Errorf("creating bucket %q: %w "+
				"(note that compression_mode and some other bucket settings are Enterprise Edition only)", name, err)
		}
		log.Println("Create Bucket -> " + name)
	}
	bucket := s.cluster.Bucket(name)
	if err := bucket.WaitUntilReady(bucketReadyTimeout, nil); err != nil {
		return fmt.Errorf("waiting for bucket %q: %w", name, err)
	}
	log.Println("Ready Bucket -> " + name)
	return s.createPrimaryIndex(name)
}

const (
	bucketReadyTimeout = 20 * time.Second
	indexReadyTimeout  = 60 * time.Second
	indexRetryInterval = time.Second
)

// queryScanConsistency decides how fresh a search has to be.
//
// N1QL defaults to "not bounded": the query runs against whatever the index
// happens to hold, so a resource that was just created is not found. That is
// the wrong default for a provisioning API, where a client creating a user and
// then searching for it is the ordinary flow, not an edge case. Request-plus
// makes the query wait for every mutation accepted before it was issued.
//
// SCIM_QUERY_CONSISTENCY=not_bounded trades that correctness back for latency,
// for deployments that would rather have the speed.
func queryScanConsistency() gocb.QueryScanConsistency {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SCIM_QUERY_CONSISTENCY")), "not_bounded") {
		return gocb.QueryScanConsistencyNotBounded
	}
	return gocb.QueryScanConsistencyRequestPlus
}

// createPrimaryIndex asks for the bucket's primary index, retrying while the
// query service catches up.
//
// A bucket that has just been created is not immediately visible to the query
// service, which answers "service not available" until it is. Asking once meant
// the very first startup against a fresh cluster failed, so this waits instead
// of giving up.
func (s *CouchbaseStore) createPrimaryIndex(name string) error {
	deadline := time.Now().Add(indexReadyTimeout)
	for attempt := 1; ; attempt++ {
		err := s.cluster.QueryIndexes().CreatePrimaryIndex(name,
			&gocb.CreatePrimaryQueryIndexOptions{IgnoreIfExists: true})
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("creating the primary index on %q after %d attempts: %w", name, attempt, err)
		}
		time.Sleep(indexRetryInterval)
	}
}
