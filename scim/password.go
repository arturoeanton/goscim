package scim

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// PasswordMode controls how writeOnly password attributes are handled before
// persistence.
type PasswordMode string

const (
	// PasswordModeArgon2id hashes incoming passwords with Argon2id (PHC string
	// format). This is the default when SCIM_PASSWORD_MODE is unset.
	PasswordModeArgon2id PasswordMode = "argon2id"
	// PasswordModeBcrypt hashes incoming passwords with bcrypt (cost 10).
	PasswordModeBcrypt PasswordMode = "bcrypt"
	// PasswordModeReject refuses any write carrying a password attribute with
	// 400 invalidValue.
	PasswordModeReject PasswordMode = "reject"
	// PasswordModePlain stores passwords as sent. Intended only for legacy or
	// local testing; startup logs a warning when selected.
	PasswordModePlain PasswordMode = "plain"
)

const (
	maxPasswordBytes = 1024

	// OWASP minimum recommended Argon2id parameters (19 MiB, t=2, p=1).
	argon2Memory  uint32 = 19 * 1024
	argon2Time    uint32 = 2
	argon2Threads uint8  = 1
	argon2KeyLen  uint32 = 32
	argon2SaltLen        = 16
)

// errPasswordClient distinguishes client-input failures (400 invalidValue)
// from internal failures such as CSPRNG errors (500).
type errPasswordClient struct {
	message string
}

func (e *errPasswordClient) Error() string { return e.message }

// PasswordModeFromEnv reads and validates SCIM_PASSWORD_MODE, defaulting to
// argon2id when unset.
func PasswordModeFromEnv() (PasswordMode, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_PASSWORD_MODE")))
	switch PasswordMode(raw) {
	case "", PasswordModeArgon2id:
		return PasswordModeArgon2id, nil
	case PasswordModeBcrypt:
		return PasswordModeBcrypt, nil
	case PasswordModeReject:
		return PasswordModeReject, nil
	case PasswordModePlain:
		return PasswordModePlain, nil
	default:
		return "", fmt.Errorf("unknown SCIM_PASSWORD_MODE=%q (expected argon2id, bcrypt, reject, or plain)", raw)
	}
}

// passwordSupported reports whether the active password mode accepts password
// writes, driving ServiceProviderConfig.changePassword.supported.
func passwordSupported() bool {
	mode, err := PasswordModeFromEnv()
	if err != nil {
		return false
	}
	return mode != PasswordModeReject
}

// HashPassword transforms a cleartext password according to mode.
func HashPassword(mode PasswordMode, cleartext string) (string, error) {
	if mode == PasswordModeReject {
		return "", &errPasswordClient{message: "password provisioning is disabled"}
	}
	if cleartext == "" {
		return "", &errPasswordClient{message: "password must not be empty"}
	}
	if len(cleartext) > maxPasswordBytes {
		return "", &errPasswordClient{message: "password exceeds maximum length"}
	}

	switch mode {
	case "", PasswordModeArgon2id:
		salt := make([]byte, argon2SaltLen)
		if _, err := rand.Read(salt); err != nil {
			return "", fmt.Errorf("generating password salt: %w", err)
		}
		key := argon2.IDKey([]byte(cleartext), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
		return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
			argon2.Version,
			argon2Memory,
			argon2Time,
			argon2Threads,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(key),
		), nil

	case PasswordModeBcrypt:
		hashed, err := bcrypt.GenerateFromPassword([]byte(cleartext), bcrypt.DefaultCost)
		if err != nil {
			if errors.Is(err, bcrypt.ErrPasswordTooLong) {
				return "", &errPasswordClient{message: "password exceeds bcrypt's 72-byte limit"}
			}
			return "", fmt.Errorf("hashing password with bcrypt: %w", err)
		}
		return string(hashed), nil

	case PasswordModePlain:
		return cleartext, nil

	default:
		return "", fmt.Errorf("unsupported password mode %q", mode)
	}
}

// VerifyPasswordHash checks candidate against an Argon2id PHC string or a
// bcrypt hash in constant time.
func VerifyPasswordHash(encoded, candidate string) (bool, error) {
	if strings.HasPrefix(encoded, "$argon2id$") {
		return verifyArgon2id(encoded, candidate)
	}
	if strings.HasPrefix(encoded, "$2a$") || strings.HasPrefix(encoded, "$2b$") || strings.HasPrefix(encoded, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(candidate))
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, errors.New("unrecognised password hash format")
}

func verifyArgon2id(encoded, candidate string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("malformed argon2id hash")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parsing argon2id version: %w", err)
	}
	var memory, iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
		return false, fmt.Errorf("parsing argon2id parameters: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decoding argon2id salt: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decoding argon2id digest: %w", err)
	}
	if len(expected) == 0 {
		return false, errors.New("empty argon2id digest")
	}
	computed := argon2.IDKey([]byte(candidate), salt, iterations, memory, threads, uint32(len(expected)))
	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}

// ApplyPasswordPolicy processes writeOnly attributes after schema validation:
//
//   - On POST (existing == nil), any present password attribute is hashed (or
//     rejected under SCIM_PASSWORD_MODE=reject).
//   - On PUT (existing != nil && patchedPaths == nil), omitted writeOnly
//     attributes are preserved from existing per RFC 7644 3.5.1 without
//     re-hashing, while newly supplied password attributes are hashed/rejected.
//   - On PATCH (existing != nil && patchedPaths != nil), only password
//     attributes explicitly set by a patch operation are hashed/rejected;
//     untouched ones keep their stored value.
func ApplyPasswordPolicy(c *gin.Context, resourceType ResourceType, incoming, existing map[string]interface{}, patchedPaths map[string]bool) bool {
	mode, err := PasswordModeFromEnv()
	if err != nil {
		internalError(c, err)
		return false
	}

	coreSchema, ok := Schemas[resourceType.Schema]
	if ok {
		if !applyObjectPasswordPolicy(c, mode, coreSchema.Attributes, "", incoming, existing, patchedPaths) {
			return false
		}
	}

	for _, ext := range resourceType.SchemaExtensions {
		extSchema, ok := Schemas[ext.Schema]
		if !ok {
			continue
		}
		var prevExt map[string]interface{}
		if existing != nil {
			prevExt, _ = existing[ext.Schema].(map[string]interface{})
		}
		nextExt, hasNext := incoming[ext.Schema].(map[string]interface{})
		if !hasNext {
			if prevExt != nil && patchedPaths == nil {
				restored := restoreWriteOnlyAttributes(extSchema.Attributes, prevExt)
				if len(restored) > 0 {
					incoming[ext.Schema] = restored
					ensureSchemaListed(incoming, ext.Schema)
				}
			}
			continue
		}
		if !applyObjectPasswordPolicy(c, mode, extSchema.Attributes, ext.Schema+".", nextExt, prevExt, patchedPaths) {
			return false
		}
	}
	return true
}

func applyObjectPasswordPolicy(
	c *gin.Context,
	mode PasswordMode,
	attributes []Attribute,
	prefix string,
	incoming, existing map[string]interface{},
	patchedPaths map[string]bool,
) bool {
	for _, attr := range attributes {
		path := prefix + attr.Name
		rawVal, present := incoming[attr.Name]

		if strings.EqualFold(attr.Mutability, "writeOnly") {
			if !present {
				// RFC 7644 3.5.1: omitting a writeOnly attribute on PUT keeps
				// the stored value rather than wiping it.
				if existing != nil && patchedPaths == nil {
					if prev, hadPrev := existing[attr.Name]; hadPrev {
						incoming[attr.Name] = prev
					}
				}
				continue
			}

			// On PATCH, incoming started from existing, so a writeOnly
			// attribute is still present even when no operation touched it.
			if patchedPaths != nil && !patchedPaths[strings.ToLower(path)] {
				continue
			}

			if isPasswordAttribute(attr) {
				cleartext, ok := rawVal.(string)
				if !ok {
					MakeTypedError(c, http.StatusBadRequest, "invalidValue", attr.Name+" must be a string")
					return false
				}
				hashed, err := HashPassword(mode, cleartext)
				if err != nil {
					var clientErr *errPasswordClient
					if errors.As(err, &clientErr) {
						MakeTypedError(c, http.StatusBadRequest, "invalidValue", clientErr.Error())
						return false
					}
					internalError(c, err)
					return false
				}
				incoming[attr.Name] = hashed
			}
			continue
		}

		if !present || len(attr.SubAttributes) == 0 {
			continue
		}
		var prevMap map[string]interface{}
		if existing != nil {
			prevMap, _ = existing[attr.Name].(map[string]interface{})
		}
		if nextMap, ok := rawVal.(map[string]interface{}); ok {
			if !applyObjectPasswordPolicy(c, mode, attr.SubAttributes, path+".", nextMap, prevMap, patchedPaths) {
				return false
			}
		}
	}
	return true
}

func restoreWriteOnlyAttributes(attributes []Attribute, existing map[string]interface{}) map[string]interface{} {
	var out map[string]interface{}
	for _, attr := range attributes {
		if strings.EqualFold(attr.Mutability, "writeOnly") {
			if prev, ok := existing[attr.Name]; ok {
				if out == nil {
					out = make(map[string]interface{})
				}
				out[attr.Name] = prev
			}
		}
	}
	return out
}

func isPasswordAttribute(attr Attribute) bool {
	return strings.EqualFold(attr.Name, "password") &&
		strings.EqualFold(attr.Type, "string") &&
		strings.EqualFold(attr.Mutability, "writeOnly")
}

// recordPatchedWriteOnlyPaths records canonical lowercase paths of attributes
// written by op so ApplyPasswordPolicy only hashes newly supplied passwords.
func recordPatchedWriteOnlyPaths(resourceType ResourceType, op Operation, touched map[string]bool) {
	operation := strings.ToLower(strings.TrimSpace(op.Op))
	if operation != "add" && operation != "replace" {
		return
	}
	if op.Path != "" {
		extURN, attrPath := splitPatchURN(resourceType, op.Path)
		if extURN != "" && attrPath == "" {
			if extObj, ok := op.Value.(map[string]interface{}); ok {
				for subKey := range extObj {
					canonical := CanonicalAttributePath(resourceType, extURN+"."+subKey)
					touched[strings.ToLower(canonical)] = true
				}
			}
			return
		}
		if segments, err := parsePatchSegments(attrPath); err == nil {
			bare := make([]string, 0, len(segments))
			for _, seg := range segments {
				bare = append(bare, seg.Attr)
			}
			joined := strings.Join(bare, ".")
			if extURN != "" {
				joined = extURN + "." + joined
			}
			canonical := CanonicalAttributePath(resourceType, joined)
			touched[strings.ToLower(canonical)] = true
			return
		}
		canonical := CanonicalAttributePath(resourceType, op.Path)
		touched[strings.ToLower(canonical)] = true
		return
	}
	values, ok := op.Value.(map[string]interface{})
	if !ok {
		return
	}
	for key, val := range values {
		extURN := matchingExtensionURN(resourceType, key)
		if extURN != "" {
			if extObj, ok := val.(map[string]interface{}); ok {
				for subKey := range extObj {
					canonical := CanonicalAttributePath(resourceType, extURN+"."+subKey)
					touched[strings.ToLower(canonical)] = true
				}
			}
			continue
		}
		canonical := CanonicalAttributePath(resourceType, key)
		touched[strings.ToLower(canonical)] = true
	}
}

func matchingExtensionURN(resourceType ResourceType, candidate string) string {
	for _, ext := range resourceType.SchemaExtensions {
		if strings.EqualFold(ext.Schema, candidate) {
			return ext.Schema
		}
	}
	return ""
}
