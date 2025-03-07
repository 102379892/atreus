package main

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"strconv"
)

// http://swagger.io/specification/#infoObject
type swaggerInfoObject struct {
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	TermsOfService string `json:"termsOfService,omitempty"`
	Version        string `json:"version"`

	Contact *swaggerContactObject `json:"contact,omitempty"`
	License *swaggerLicenseObject `json:"license,omitempty"`
}

// http://swagger.io/specification/#contactObject
type swaggerContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// http://swagger.io/specification/#licenseObject
type swaggerLicenseObject struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// http://swagger.io/specification/#externalDocumentationObject
type swaggerExternalDocumentationObject struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// http://swagger.io/specification/#swaggerObject
type swaggerObject struct {
	Swagger             string                              `json:"swagger"`
	Info                swaggerInfoObject                   `json:"info"`
	Host                string                              `json:"host,omitempty"`
	BasePath            string                              `json:"basePath,omitempty"`
	Schemes             []string                            `json:"schemes"`
	Consumes            []string                            `json:"consumes"`
	Produces            []string                            `json:"produces"`
	Paths               swaggerPathsObject                  `json:"paths"`
	Definitions         swaggerDefinitionsObject            `json:"definitions"`
	StreamDefinitions   swaggerDefinitionsObject            `json:"x-stream-definitions,omitempty"`
	SecurityDefinitions swaggerSecurityDefinitionsObject    `json:"securityDefinitions,omitempty"`
	Security            []swaggerSecurityRequirementObject  `json:"security,omitempty"`
	ExternalDocs        *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`
}

// http://swagger.io/specification/#securityDefinitionsObject
type swaggerSecurityDefinitionsObject map[string]swaggerSecuritySchemeObject

// http://swagger.io/specification/#securitySchemeObject
type swaggerSecuritySchemeObject struct {
	Type             string              `json:"type"`
	Description      string              `json:"description,omitempty"`
	Name             string              `json:"name,omitempty"`
	In               string              `json:"in,omitempty"`
	Flow             string              `json:"flow,omitempty"`
	AuthorizationURL string              `json:"authorizationUrl,omitempty"`
	TokenURL         string              `json:"tokenUrl,omitempty"`
	Scopes           swaggerScopesObject `json:"scopes,omitempty"`
}

// http://swagger.io/specification/#scopesObject
type swaggerScopesObject map[string]string

// http://swagger.io/specification/#securityRequirementObject
type swaggerSecurityRequirementObject map[string][]string

// http://swagger.io/specification/#pathsObject
type swaggerPathsObject map[string]swaggerPathItemObject

type swaggerPathSummary struct {
	Path    string
	Summary string
}

func (po swaggerPathsObject) MarshalJSON() ([]byte, error) {
	psv := []swaggerPathSummary{}

	for k, v := range po {
		if len(k) == 0 {
			continue
		}
		summary := k
		if v.Get != nil && len(v.Get.Summary) > 0 {
			summary = v.Get.Summary
		} else if v.Post != nil && len(v.Post.Summary) > 0 {
			summary = v.Post.Summary
		} else if v.Delete != nil && len(v.Delete.Summary) > 0 {
			summary = v.Delete.Summary
		} else if v.Put != nil && len(v.Put.Summary) > 0 {
			summary = v.Put.Summary
		} else if v.Patch != nil && len(v.Patch.Summary) > 0 {
			summary = v.Patch.Summary
		} else {
			summary = k
		}

		psv = append(psv, swaggerPathSummary{
			Path:    k,
			Summary: summary,
		})
	}

	sort.Slice(psv, func(i, j int) bool {
		rsi := strings.FieldsFunc(psv[i].Summary, func(r rune) bool {
			return strings.ContainsRune(".,。:", r)
		})

		if len(rsi) == 0 {
			return false
		}

		rsj := strings.FieldsFunc(psv[j].Summary, func(r rune) bool {
			return strings.ContainsRune(".,。:", r)
		})

		if len(rsj) == 0 {
			return true
		}

		indexi, erri := strconv.Atoi(rsi[0])
		indexj, errj := strconv.Atoi(rsj[0])

		if erri == nil && errj == nil {
			return indexi < indexj
		} else if erri != nil {
			return false
		} else if errj != nil {
			return true
		} else {
			return psv[i].Summary < psv[j].Summary
		}
	})

	var buf bytes.Buffer
	buf.WriteString("{")
	for i, ps := range psv {
		if len(ps.Path) == 0 {
			continue
		}

		if i != 0 {
			buf.WriteString(",")
		}

		path, err := json.Marshal(ps.Path)
		if err != nil {
			return nil, err
		}

		buf.Write(path)
		buf.WriteString(":")
		po, _ := po[ps.Path]
		val, err := json.Marshal(po)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}

// http://swagger.io/specification/#pathItemObject
type swaggerPathItemObject struct {
	Get    *swaggerOperationObject `json:"get,omitempty"`
	Delete *swaggerOperationObject `json:"delete,omitempty"`
	Post   *swaggerOperationObject `json:"post,omitempty"`
	Put    *swaggerOperationObject `json:"put,omitempty"`
	Patch  *swaggerOperationObject `json:"patch,omitempty"`
}

// http://swagger.io/specification/#operationObject
type swaggerOperationObject struct {
	Summary     string                  `json:"summary,omitempty"`
	Description string                  `json:"description,omitempty"`
	OperationID string                  `json:"operationId,omitempty"`
	Responses   swaggerResponsesObject  `json:"responses"`
	Parameters  swaggerParametersObject `json:"parameters,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	Deprecated  bool                    `json:"deprecated,omitempty"`

	Security     *[]swaggerSecurityRequirementObject `json:"security,omitempty"`
	ExternalDocs *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type swaggerParametersObject []swaggerParameterObject

// http://swagger.io/specification/#parameterObject
type swaggerParameterObject struct {
	Name             string              `json:"name"`
	Description      string              `json:"description,omitempty"`
	In               string              `json:"in,omitempty"`
	Required         bool                `json:"required"`
	Type             string              `json:"type,omitempty"`
	Format           string              `json:"format,omitempty"`
	Items            *swaggerItemsObject `json:"items,omitempty"`
	Enum             []string            `json:"enum,omitempty"`
	CollectionFormat string              `json:"collectionFormat,omitempty"`
	Default          string              `json:"default,omitempty"`
	MinItems         *int                `json:"minItems,omitempty"`

	// Or you can explicitly refer to another type. If this is defined all
	// other fields should be empty
	Schema *swaggerSchemaObject `json:"schema,omitempty"`
}

// core part of schema, which is common to itemsObject and schemaObject.
// http://swagger.io/specification/#itemsObject
type schemaCore struct {
	Type    string          `json:"type,omitempty"`
	Format  string          `json:"format,omitempty"`
	Ref     string          `json:"$ref,omitempty"`
	Example json.RawMessage `json:"example,omitempty"`

	Items *swaggerItemsObject `json:"items,omitempty"`

	// If the item is an enumeration include a list of all the *NAMES* of the
	// enum values.  I'm not sure how well this will work but assuming all enums
	// start from 0 index it will be great. I don't think that is a good assumption.
	Enum    []string `json:"enum,omitempty"`
	Default string   `json:"default,omitempty"`
}

type swaggerItemsObject schemaCore

func (o *swaggerItemsObject) getType() string {
	if o == nil {
		return ""
	}
	return o.Type
}

// http://swagger.io/specification/#responsesObject
type swaggerResponsesObject map[string]swaggerResponseObject

// http://swagger.io/specification/#responseObject
type swaggerResponseObject struct {
	Description string              `json:"description"`
	Schema      swaggerSchemaObject `json:"schema"`
}

type keyVal struct {
	Key   string
	Value interface{}
}

type swaggerSchemaObjectProperties []keyVal

func (op swaggerSchemaObjectProperties) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range op {
		if i != 0 {
			buf.WriteString(",")
		}
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

// http://swagger.io/specification/#schemaObject
type swaggerSchemaObject struct {
	schemaCore
	// Properties can be recursively defined
	Properties           *swaggerSchemaObjectProperties `json:"properties,omitempty"`
	AdditionalProperties *swaggerSchemaObject           `json:"additionalProperties,omitempty"`

	Description string `json:"description,omitempty"`
	Title       string `json:"title,omitempty"`

	ExternalDocs *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`

	MultipleOf       float64  `json:"multipleOf,omitempty"`
	Maximum          float64  `json:"maximum,omitempty"`
	ExclusiveMaximum bool     `json:"exclusiveMaximum,omitempty"`
	Minimum          float64  `json:"minimum,omitempty"`
	ExclusiveMinimum bool     `json:"exclusiveMinimum,omitempty"`
	MaxLength        uint64   `json:"maxLength,omitempty"`
	MinLength        uint64   `json:"minLength,omitempty"`
	Pattern          string   `json:"pattern,omitempty"`
	MaxItems         uint64   `json:"maxItems,omitempty"`
	MinItems         uint64   `json:"minItems,omitempty"`
	UniqueItems      bool     `json:"uniqueItems,omitempty"`
	MaxProperties    uint64   `json:"maxProperties,omitempty"`
	MinProperties    uint64   `json:"minProperties,omitempty"`
	Required         []string `json:"required,omitempty"`
}

// http://swagger.io/specification/#definitionsObject
type swaggerDefinitionsObject map[string]swaggerSchemaObject
