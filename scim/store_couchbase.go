package scim

import (
	"context"
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

const (
	uniqueScopeName      = "_default"
	uniqueCollectionName = "scim_unique"
	orphanSentinelGrace  = 30 * time.Second
	maxCASRetries        = 5
)

var (
	//Cluster ...
	Cluster             *gocb.Cluster
	FolderBucketSetting = "config/bucketSettings/"
)

type sentinelDoc struct {
	Owner     string `json:"owner"`
	Attribute string `json:"attribute"`
	Key       string `json:"key"`
	CreatedAt int64  `json:"createdAt"`
}

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
		return "(empty)"
	}
	return "[REDACTED]"
}

// Ping implements Store, checking that the query service and a configured
// bucket's key-value service answer within ctx.
func (s *CouchbaseStore) Ping(ctx context.Context) error {
	if s == nil || s.cluster == nil {
		return errors.New("couchbase cluster is not initialised")
	}
	res, err := s.cluster.Ping(&gocb.PingOptions{
		ServiceTypes: []gocb.ServiceType{gocb.ServiceTypeQuery},
		Timeout:      2 * time.Second,
		Context:      ctx,
	})
	if err != nil {
		return err
	}
	if err := checkPingReport(res, gocb.ServiceTypeQuery); err != nil {
		return err
	}
	for _, rt := range Resources {
		if rt.Name == "" {
			continue
		}
		kvRes, err := s.cluster.Bucket(rt.Name).Ping(&gocb.PingOptions{
			ServiceTypes: []gocb.ServiceType{gocb.ServiceTypeKeyValue},
			Timeout:      2 * time.Second,
			Context:      ctx,
		})
		if err != nil {
			return err
		}
		return checkPingReport(kvRes, gocb.ServiceTypeKeyValue)
	}
	return nil
}

func checkPingReport(res *gocb.PingResult, svc gocb.ServiceType) error {
	if res == nil {
		return errors.New("empty ping report")
	}
	endpoints := res.Services[svc]
	if len(endpoints) == 0 {
		return fmt.Errorf("no %v endpoints responded to ping", svc)
	}
	for _, ep := range endpoints {
		if ep.State != gocb.PingStateOk {
			if ep.Error != "" {
				return fmt.Errorf("%v ping failed: %s", svc, ep.Error)
			}
			return fmt.Errorf("%v ping returned non-ok state", svc)
		}
	}
	return nil
}

// Close releases the cluster connection.
func (s *CouchbaseStore) Close() error {
	return s.cluster.Close(nil)
}

func (s *CouchbaseStore) collection(bucket string) *gocb.Collection {
	return s.cluster.Bucket(bucket).DefaultCollection()
}

func (s *CouchbaseStore) uniqueCollection(bucket string) *gocb.Collection {
	return s.cluster.Bucket(bucket).Scope(uniqueScopeName).Collection(uniqueCollectionName)
}

// isKeyNotFound reports whether err is Couchbase's "document missing" error.
func isKeyNotFound(err error) bool {
	var kvErr *gocb.KeyValueError
	if errors.As(err, &kvErr) {
		return kvErr.StatusCode == memd.StatusKeyNotFound
	}
	return errors.Is(err, gocb.ErrDocumentNotFound)
}

// isKeyExists reports whether err is Couchbase's "document already exists" error.
func isKeyExists(err error) bool {
	var kvErr *gocb.KeyValueError
	if errors.As(err, &kvErr) {
		return kvErr.StatusCode == memd.StatusKeyExists
	}
	return errors.Is(err, gocb.ErrDocumentExists)
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

// CreateUnique implements Store using atomic KV Insert operations in the
// dedicated scim_unique collection before writing the resource document.
func (s *CouchbaseStore) CreateUnique(bucket, id string, doc map[string]interface{}, claims []UniqueClaim) error {
	acquired := make([]string, 0, len(claims))
	rollback := func() {
		for _, lockID := range acquired {
			s.releaseSentinel(bucket, id, lockID)
		}
	}

	for _, claim := range claims {
		isNew, err := s.acquireSentinel(bucket, id, claim)
		if err != nil {
			rollback()
			return err
		}
		if isNew {
			acquired = append(acquired, claim.LockID)
		}
		if err := s.checkLegacyUniqueOwner(bucket, id, claim); err != nil {
			rollback()
			return err
		}
	}

	if _, err := s.collection(bucket).Insert(id, doc, &gocb.InsertOptions{}); err != nil {
		rollback()
		if isKeyExists(err) {
			return &UniqueConflictError{Attribute: "id"}
		}
		return err
	}
	return nil
}

// ReplaceUnique implements Store using CAS on the resource document combined
// with atomic KV sentinels in scim_unique.
func (s *CouchbaseStore) ReplaceUnique(bucket, id string, doc map[string]interface{}, newClaims []UniqueClaim, extractClaims func(map[string]interface{}) []UniqueClaim) error {
	newSet := make(map[string]struct{}, len(newClaims))
	for _, claim := range newClaims {
		newSet[claim.LockID] = struct{}{}
	}

	for attempt := 0; attempt < maxCASRetries; attempt++ {
		getRes, err := s.collection(bucket).Get(id, &gocb.GetOptions{})
		if err != nil {
			if isKeyNotFound(err) {
				return ErrNotFound
			}
			return err
		}
		var currentDoc map[string]interface{}
		if err := getRes.Content(&currentDoc); err != nil {
			return err
		}
		var oldClaims []UniqueClaim
		if extractClaims != nil {
			oldClaims = extractClaims(currentDoc)
		}

		newlyAcquired := make([]string, 0, len(newClaims))
		rollbackNew := func() {
			for _, lockID := range newlyAcquired {
				s.releaseSentinel(bucket, id, lockID)
			}
		}

		var claimErr error
		for _, claim := range newClaims {
			isNew, err := s.acquireSentinel(bucket, id, claim)
			if err != nil {
				claimErr = err
				break
			}
			if isNew {
				newlyAcquired = append(newlyAcquired, claim.LockID)
			}
			if err := s.checkLegacyUniqueOwner(bucket, id, claim); err != nil {
				claimErr = err
				break
			}
		}
		if claimErr != nil {
			rollbackNew()
			return claimErr
		}

		_, err = s.collection(bucket).Replace(id, doc, &gocb.ReplaceOptions{Cas: getRes.Cas()})
		if err != nil {
			rollbackNew()
			if isKeyNotFound(err) {
				return ErrNotFound
			}
			if errors.Is(err, gocb.ErrCasMismatch) {
				continue
			}
			return err
		}

		for _, old := range oldClaims {
			if _, stillHeld := newSet[old.LockID]; !stillHeld {
				s.releaseSentinel(bucket, id, old.LockID)
			}
		}
		return nil
	}
	return errors.New("scim: concurrent modification retry limit exceeded")
}

// RemoveUnique implements Store.
func (s *CouchbaseStore) RemoveUnique(bucket, id string, claims []UniqueClaim) error {
	if err := s.Remove(bucket, id); err != nil {
		return err
	}
	for _, claim := range claims {
		s.releaseSentinel(bucket, id, claim.LockID)
	}
	return nil
}

func (s *CouchbaseStore) acquireSentinel(bucket, ownerID string, claim UniqueClaim) (bool, error) {
	col := s.uniqueCollection(bucket)
	nowMs := time.Now().UnixMilli()
	candidate := sentinelDoc{
		Owner:     ownerID,
		Attribute: claim.Attribute,
		Key:       claim.Key,
		CreatedAt: nowMs,
	}

	_, err := col.Insert(claim.LockID, candidate, &gocb.InsertOptions{})
	if err == nil {
		return true, nil
	}
	if !isKeyExists(err) {
		return false, err
	}

	getRes, getErr := col.Get(claim.LockID, &gocb.GetOptions{})
	if getErr != nil {
		if isKeyNotFound(getErr) {
			if _, retryErr := col.Insert(claim.LockID, candidate, &gocb.InsertOptions{}); retryErr == nil {
				return true, nil
			}
		}
		return false, &UniqueConflictError{Attribute: claim.Attribute}
	}

	var existing sentinelDoc
	if err := getRes.Content(&existing); err != nil {
		return false, &UniqueConflictError{Attribute: claim.Attribute}
	}
	if existing.Owner == ownerID {
		return false, nil
	}

	// Self-heal sentinels left behind by a process crash: only consider entries
	// older than orphanSentinelGrace so an in-flight writer that has inserted
	// its sentinel milliseconds before inserting its resource cannot have its
	// claim stolen.
	if existing.CreatedAt > 0 && time.Since(time.UnixMilli(existing.CreatedAt)) > orphanSentinelGrace {
		_, ownerErr := s.collection(bucket).Get(existing.Owner, &gocb.GetOptions{})
		if isKeyNotFound(ownerErr) {
			if _, takeoverErr := col.Replace(claim.LockID, candidate, &gocb.ReplaceOptions{Cas: getRes.Cas()}); takeoverErr == nil {
				return true, nil
			}
		}
	}

	return false, &UniqueConflictError{Attribute: claim.Attribute}
}

func (s *CouchbaseStore) releaseSentinel(bucket, ownerID, lockID string) {
	col := s.uniqueCollection(bucket)
	getRes, err := col.Get(lockID, &gocb.GetOptions{})
	if err != nil {
		return
	}
	var existing sentinelDoc
	if err := getRes.Content(&existing); err != nil || existing.Owner != ownerID {
		return
	}
	_, _ = col.Remove(lockID, &gocb.RemoveOptions{Cas: getRes.Cas()})
}

// checkLegacyUniqueOwner guards upgrades where resources created prior to the
// scim_unique collection do not yet have sentinel documents.
func (s *CouchbaseStore) checkLegacyUniqueOwner(bucket, selfID string, claim UniqueClaim) error {
	quotedAttr := parser.AddQuote(claim.Attribute)
	var query string
	if !claim.CaseExact {
		if _, isStr := claim.RawValue.(string); isStr {
			query = "SELECT id FROM `" + bucket + "` WHERE LOWER(" + quotedAttr + ") = LOWER($1) AND `id` <> $2 LIMIT 1"
		}
	}
	if query == "" {
		query = "SELECT id FROM `" + bucket + "` WHERE " + quotedAttr + " = $1 AND `id` <> $2 LIMIT 1"
	}

	rows, err := s.cluster.Query(query, &gocb.QueryOptions{
		ScanConsistency:      queryScanConsistency(),
		PositionalParameters: []interface{}{claim.RawValue, selfID},
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var row struct {
		ID string `json:"id"`
	}
	if err := rows.One(&row); err != nil {
		if errors.Is(err, gocb.ErrNoResult) {
			return nil
		}
		return err
	}
	if row.ID != "" {
		return &UniqueConflictError{Attribute: claim.Attribute}
	}
	return rows.Err()
}

// Search implements Store by translating the SCIM filter to N1QL.
func (s *CouchbaseStore) Search(q SearchQuery) (int, []map[string]interface{}, error) {
	query, err := parser.FilterToN1QLWithOptions(q.Bucket, q.Filter, parser.FilterOptions{
		Resolve:       q.ResolveName,
		IsCaseExact:   q.IsCaseExact,
		IsMultiValued: q.IsMultiValued,
	})
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

	queryPage := query.Page +
		"\nORDER BY " + sortBy + " " + sortOrder +
		"\nOFFSET " + strconv.Itoa(q.Offset) +
		"\nLIMIT " + strconv.Itoa(q.Limit)

	options := &gocb.QueryOptions{
		ScanConsistency:      queryScanConsistency(),
		PositionalParameters: query.Params,
	}

	rowsCount, err := s.cluster.Query(query.Count, options)
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

// FindIDByAttribute implements Store.
//
// The value is passed as a bound parameter rather than concatenated: it comes
// from a client payload, and this is a query built outside the filter
// translation, so it must not grow its own injection hole.
func (s *CouchbaseStore) FindIDByAttribute(bucket, attributePath string, value interface{}) (string, error) {
	query := "SELECT id FROM `" + bucket + "` WHERE " + parser.AddQuote(attributePath) + " = $1 LIMIT 1"

	rows, err := s.cluster.Query(query, &gocb.QueryOptions{
		ScanConsistency:      queryScanConsistency(),
		PositionalParameters: []interface{}{value},
	})
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var row struct {
		ID string `json:"id"`
	}
	if err := rows.One(&row); err != nil {
		if errors.Is(err, gocb.ErrNoResult) {
			return "", nil
		}
		return "", err
	}
	return row.ID, rows.Err()
}

// EnsureBucket implements Store, creating the bucket, its scim_unique sentinel
// collection, and its primary N1QL index.
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
	if defaultConfig.BucketType != gocb.MemcachedBucketType {
		if err := s.ensureUniqueCollection(name); err != nil {
			return err
		}
	}
	return s.createPrimaryIndex(name)
}

func (s *CouchbaseStore) ensureUniqueCollection(bucketName string) error {
	bucket := s.cluster.Bucket(bucketName)
	deadline := time.Now().Add(bucketReadyTimeout)
	for {
		err := bucket.CollectionsV2().CreateCollection(uniqueScopeName, uniqueCollectionName, nil, nil)
		if err == nil || errors.Is(err, gocb.ErrCollectionExists) || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("creating unique collection on %q: %w", bucketName, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Wait until the KV engine's collection manifest recognises scim_unique so
	// an immediate write right after startup cannot fail with CollectionNotFound.
	col := s.uniqueCollection(bucketName)
	for {
		_, err := col.Exists("_probe", &gocb.ExistsOptions{})
		if err == nil || isKeyNotFound(err) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("waiting for unique collection on %q: %w", bucketName, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
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
