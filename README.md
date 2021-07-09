# goscim
SCIM server written in Go

This is one lite implementation of SCIM 2.0 (http://www.simplecloud.info/)

## Overview 
The System for Cross-domain Identity Management (SCIM) specification is designed to make managing user identities in cloud-based applications and services easier. The specification suite seeks to build upon experience with existing schemas and deployments, placing specific emphasis on simplicity of development and integration, while applying existing authentication, authorization, and privacy models. Its intent is to reduce the cost and complexity of user management operations by providing a common user schema and extension model, as well as binding documents to provide patterns for exchanging this schema using standard protocols. In essence: make it fast, cheap, and easy to move users in to, out of, and around the cloud.
(cut and past :P )

## Run

For run you need one sever couchbase (https://www.couchbase.com/)

### Run docker couchbase

```
docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase
```

 - Go to http://localhost:8091/ui/index.html 
 - If you need config only click in "Setup New Cluster" 
 - Complete "Cluster Name", "Create Admin Username" and "Create Password" and click "Next: Accept Terms"
  - Example 
    - "Cluster Name" -> cluster_scim
    - "Create Admin Username" -> Administrator
    - "Create Password" -> admin!
 - Check "I accept the terms & conditions" and click "Finish With Defaults"

### Set env 

```
export SCIM_ADMIN_USER=Administrator
export SCIM_ADMIN_PASSWORD=admin!
```

You can set SCIM_COUCHBASE_URL but the default value is localhost.


### Run Scim Server

```
go run .
```

## Test http 

I used REST Client https://github.com/Huachao/vscode-restclient

```http
@host=http://127.0.0.1:8080/scim/v2


###

POST {{host}}/Elements/ HTTP/1.1
content-type: application/json

{
    "schemas":[
        "urn:ietf:params:scim:schemas:core:2.0:Element",
        "urn:ietf:params:scim:schemas:extension:one:2.0:Element"
    ],
    "name": "Element1",
    "description":"This is Element 1",
    "$ref": "/Element2",
    "urn:ietf:params:scim:schemas:extension:one:2.0:Element":{
        "required": 1
    }
}

###

POST {{host}}/Elements/ HTTP/1.1
content-type: application/json

{
    "schemas":[
        "urn:ietf:params:scim:schemas:core:2.0:Element",
        "urn:ietf:params:scim:schemas:extension:one:2.0:Element"
    ],
    "name": "Element2",
    "description":"This is Element 2",
    "$ref": "/Element3",
    "urn:ietf:params:scim:schemas:extension:one:2.0:Element":{
        "required": 2
    }
}

###

POST {{host}}/Elements/ HTTP/1.1
content-type: application/json

{
    "schemas":[
        "urn:ietf:params:scim:schemas:core:2.0:Element",
        "urn:ietf:params:scim:schemas:extension:one:2.0:Element"
    ],
    "name": "Element3",
    "description":"This is Element 3",
    "$ref": "/Element4",
    "urn:ietf:params:scim:schemas:extension:one:2.0:Element":{
        "required": 3
    }
}

###

GET {{host}}/Elements HTTP/1.1
content-type: application/json

###

GET {{host}}/Elements
    ?filter=urn:ietf:params:scim:schemas:extension:one:2.0:Element.required ge 0
    &startIndex=1
    &count=2
    &sortBy=urn:ietf:params:scim:schemas:extension:one:2.0:Element.required
    &sortOrder=descending
```

## Schemas

The schemas is location in **config/schemas** and you can modify them and add new schemas. 
Remember that one new schema need one resourceType. The ResourceType is location in **config/resourceType**


### Example User (urn:ietf:params:scim:schemas:extension:enterprise:2.0:User)

#### Resource Type User.json
```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
  "id": "User",
  "name": "User",
  "endpoint": "/Users",
  "description": "User Account",
  "schema": "urn:ietf:params:scim:schemas:core:2.0:User",
  "schemaExtensions": [
    {
      "schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
      "required": false
    }
  ],
  "meta": {
    "location": "/v2/ResourceTypes/User",
    "resourceType": "ResourceType"
  }
}
```

#### Schema urn+ietf+params+scim+schemas+core+2.0+User.json
```json
{
  "id" : "urn:ietf:params:scim:schemas:core:2.0:User",
  "name" : "User",
  "description" : "User Account",
  "attributes" : [
    {
      "name" : "userName",
      "type" : "string",
      "multiValued" : false,
      "description" : "Unique identifier for the User, typically used by the user to directly authenticate to the service provider. Each User MUST include a non-empty userName value.  This identifier MUST be unique across the service provider's entire set of Users. REQUIRED.",
      "required" : true,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "server"
    }, {
      "name" : "name",
      "type" : "complex",
      "multiValued" : false,
      "description" : "The components of the user's real name. Providers MAY return just the full name as a single string in the formatted sub-attribute, or they MAY return just the individual component attributes using the other sub-attributes, or they MAY return both.  If both variants are returned, they SHOULD be describing the same name, with the formatted name indicating how the component attributes should be combined.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "formatted",
          "type" : "string",
          "multiValued" : false,
          "description" : "The full name, including all middle names, titles, and suffixes as appropriate, formatted for display (e.g., 'Ms. Barbara J Jensen, III').",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "familyName",
          "type" : "string",
          "multiValued" : false,
          "description" : "The family name of the User, or last name in most Western languages (e.g., 'Jensen' given the full name 'Ms. Barbara J Jensen, III').",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },{
          "name" : "givenName",
          "type" : "string",
          "multiValued" : false,
          "description" : "The given name of the User, or first name in most Western languages (e.g., 'Barbara' given the full name 'Ms. Barbara J Jensen, III').",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "middleName",
          "type" : "string",
          "multiValued" : false,
          "description" : "The middle name(s) of the User (e.g., 'Jane' given the full name 'Ms. Barbara J Jensen, III').",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "honorificPrefix",
          "type" : "string",
          "multiValued" : false,
          "description" : "The honorific prefix(es) of the User, or title in most Western languages (e.g., 'Ms.' given the full name 'Ms. Barbara J Jensen, III').",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }, {
          "name" : "honorificSuffix",
          "type" : "string",
          "multiValued" : false,
          "description" : "The honorific suffix(es) of the User, or suffix in most Western languages (e.g., 'III' given the full name 'Ms. Barbara J Jensen, III').",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "displayName",
      "type" : "string",
      "multiValued" : false,
      "description" : "The name of the User, suitable for display to end-users.  The name SHOULD be the full name of the User being described, if known.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "nickName",
      "type" : "string",
      "multiValued" : false,
      "description" : "The casual way to address the user in real  life, e.g., 'Bob' or 'Bobby' instead of 'Robert'.  This attribute SHOULD NOT be used to represent a User's username (e.g., 'bjensen' or 'mpepperidge').",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    }, {
      "name" : "profileUrl",
      "type" : "reference",
      "referenceTypes" : ["external"],
      "multiValued" : false,
      "description" : "A fully qualified URL pointing to a page representing the User's online profile.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "title",
      "type" : "string",
      "multiValued" : false,
      "description" : "The user's title, such as \"Vice President.\"",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "userType",
      "type" : "string",
      "multiValued" : false,
      "description" : "Used to identify the relationship between the organization and the user.  Typical values used might be 'Contractor', 'Employee', 'Intern', 'Temp', 'External', and 'Unknown', but any value may be used.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    }, {
      "name" : "preferredLanguage",
      "type" : "string",
      "multiValued" : false,
      "description" : "Indicates the User's preferred written or spoken language.  Generally used for selecting a localized user interface; e.g., 'en_US' specifies the language English and country US.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "locale",
      "type" : "string",
      "multiValued" : false,
      "description" : "Used to indicate the User's default location for purposes of localizing items such as currency, date time format, or numerical representations.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "timezone",
      "type" : "string",
      "multiValued" : false,
      "description" : "The User's time zone in the 'Olson' time zone database format, e.g., 'America/Los_Angeles'.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    }, {
      "name" : "active",
      "type" : "boolean",
      "multiValued" : false,
      "description" : "A Boolean value indicating the User's administrative status.",
      "required" : false,
      "mutability" : "readWrite",
      "returned" : "default"
    },
    {
      "name" : "password",
      "type" : "string",
      "multiValued" : false,
      "description" : "The User's cleartext password.  This attribute is intended to be used as a means to specify an initial password when creating a new User or to reset an existing User's password.",
      "required" : false,
      "caseExact" : false,
      "mutability" : "writeOnly",
      "returned" : "never",
      "uniqueness" : "none"
    },
    {
      "name" : "emails",
      "type" : "complex",
      "multiValued" : true,
      "description" : "Email addresses for the user.  The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "string",
          "multiValued" : false,
          "description" : "Email addresses for the user.  The value SHOULD be canonicalized by the service provider, e.g.,  'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }, {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function, e.g., 'work' or 'home'.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [
            "work",
            "home",
            "other"
          ],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred mailing address or primary email address.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    }, {
      "name" : "phoneNumbers",
      "type" : "complex",
      "multiValued" : true,
      "description" : "Phone numbers for the User.  The value SHOULD be canonicalized by the service provider according to the format specified in RFC 3966, e.g., 'tel:+1-201-555-0123'. Canonical type values of 'work', 'home', 'mobile', 'fax', 'pager',  and 'other'.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "string",
          "multiValued" : false,
          "description" : "Phone number of the User.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function, e.g., 'work', 'home', 'mobile'.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [
            "work",
            "home",
            "mobile",
            "fax",
            "pager",
            "other"
          ],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred phone number or primary phone number.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default"
    },
    {
      "name" : "ims",
      "type" : "complex",
      "multiValued" : true,
      "description" : "Instant messaging addresses for the User.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "string",
          "multiValued" : false,
          "description" : "Instant messaging address for the User.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }, {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function, e.g., 'aim', 'gtalk', 'xmpp'.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [
            "aim",
            "gtalk",
            "icq",
            "xmpp",
            "msn",
            "skype",
            "qq",
            "yahoo"
          ],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred messenger or primary messenger.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default"
    }, {
      "name" : "photos",
      "type" : "complex",
      "multiValued" : true,
      "description" : "URLs of photos of the User.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "reference",
          "referenceTypes" : ["external"],
          "multiValued" : false,
          "description" : "URL of a photo of the User.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }, {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function, i.e., 'photo' or 'thumbnail'.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [
            "photo",
            "thumbnail"
          ],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred photo or thumbnail.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default"
    }, {
      "name" : "addresses",
      "type" : "complex",
      "multiValued" : true,
      "description" : "A physical mailing address for this User. Canonical type values of 'work', 'home', and 'other'.  This attribute is a complex type with the following sub-attributes.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "formatted",
          "type" : "string",
          "multiValued" : false,
          "description" : "The full mailing address, formatted for display or use with a mailing label.  This attribute MAY contain newlines.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "streetAddress",
          "type" : "string",
          "multiValued" : false,
          "description" : "The full street address component, which may include house number, street name, P.O. box, and multi-line extended street address information.  This attribute MAY contain newlines.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "locality",
          "type" : "string",
          "multiValued" : false,
          "description" : "The city or locality component.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "region",
          "type" : "string",
          "multiValued" : false,
          "description" : "The state or region component.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "postalCode",
          "type" : "string",
          "multiValued" : false,
          "description" : "The zip code or postal code component.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "country",
          "type" : "string",
          "multiValued" : false,
          "description" : "The country name component.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },{
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function, e.g., 'work' or 'home'.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [
            "work",
            "home",
            "other"
          ],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default",
      "uniqueness" : "none"
    },
    {
      "name" : "groups",
      "type" : "complex",
      "multiValued" : true,
      "description" : "A list of groups to which the user belongs, either through direct membership, through nested groups, or dynamically calculated.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "string",
          "multiValued" : false,
          "description" : "The identifier of the User's group.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readOnly",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "$ref",
          "type" : "reference",
          "referenceTypes" : [
            "User",
            "Group"
          ],
          "multiValued" : false,
          "description" : "The URI of the corresponding 'Group' resource to which the user belongs.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readOnly",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readOnly",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function, e.g., 'direct' or 'indirect'.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [
            "direct",
            "indirect"
          ],
          "mutability" : "readOnly",
          "returned" : "default",
          "uniqueness" : "none"
        }
      ],
      "mutability" : "readOnly",
      "returned" : "default"
    }, {
      "name" : "entitlements",
      "type" : "complex",
      "multiValued" : true,
      "description" : "A list of entitlements for the User that represent a thing the User has.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "string",
          "multiValued" : false,
          "description" : "The value of an entitlement.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        }, {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default"
    },
    {
      "name" : "roles",
      "type" : "complex",
      "multiValued" : true,
      "description" : "A list of roles for the User that collectively represent who the User is, e.g., 'Student', 'Faculty'.",
      "required" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "string",
          "multiValued" : false,
          "description" : "The value of a role.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default"
    },
    {
      "name" : "x509Certificates",
      "type" : "complex",
      "multiValued" : true,
      "description" : "A list of certificates issued to the User.",
      "required" : false,
      "caseExact" : false,
      "subAttributes" : [
        {
          "name" : "value",
          "type" : "binary",
          "multiValued" : false,
          "description" : "The value of an X.509 certificate.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "display",
          "type" : "string",
          "multiValued" : false,
          "description" : "A human-readable name, primarily used for display purposes.  READ-ONLY.",
          "required" : false,
          "caseExact" : false,
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "type",
          "type" : "string",
          "multiValued" : false,
          "description" : "A label indicating the attribute's function.",
          "required" : false,
          "caseExact" : false,
          "canonicalValues" : [],
          "mutability" : "readWrite",
          "returned" : "default",
          "uniqueness" : "none"
        },
        {
          "name" : "primary",
          "type" : "boolean",
          "multiValued" : false,
          "description" : "A Boolean value indicating the 'primary' or preferred attribute value for this attribute.  The primary attribute value 'true' MUST appear no more than once.",
          "required" : false,
          "mutability" : "readWrite",
          "returned" : "default"
        }
      ],
      "mutability" : "readWrite",
      "returned" : "default"
    }
  ],
  "meta" : {
    "resourceType" : "Schema",
    "location" : "/v2/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
  }
}

```




## Chanege parser SCIM filter.

wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4





