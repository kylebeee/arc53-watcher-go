package uuid

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/rs/xid"
	"gopkg.in/yaml.v3"
)

const pkg errors.Pkg = "uuid"

// Prefix is a type for object prefix constants to use on uuid's
type Prefix int

// uuid object constants
const (
	Collection Prefix = iota + 1
	Property
)

func (p Prefix) String() string {
	return prefixToString[p]
}

var prefixToString = map[Prefix]string{
	Collection: "col",
	Property:   "prp",
}

var prefixToID = map[string]Prefix{
	"col": Collection,
	"prp": Property,
}

// UnmarshalYAML checks to see if its type is valid
func (p *Prefix) UnmarshalYAML(v *yaml.Node) error {
	const op errors.Op = "Prefix.UnmarshalYAML"

	var str string
	err := v.Decode(&str)
	if err != nil {
		return errors.E(pkg, op, err)
	}

	if strings.TrimSpace(str) == "" {
		*p = 0
		return nil
	}

	prefixID, ok := prefixToID[str]
	if !ok {
		return errors.E(pkg, op, fmt.Errorf("Prefix: %v is invalid", str))
	}
	*p = prefixID
	return nil
}

// MarshalYAML stringifies our enum
func (p *Prefix) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

// UnmarshalJSON properly decodes our type
func (p *Prefix) UnmarshalJSON(b []byte) error {
	const op errors.Op = "Prefix.UnmarshalJSON"

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return errors.E(pkg, op, err)
	}

	if strings.TrimSpace(str) == "" {
		*p = 0
		return nil
	}

	prefixID, ok := prefixToID[str]
	if !ok {
		return errors.E(pkg, op, fmt.Errorf("Prefix: %v is invalid", str))
	}
	*p = prefixID
	return nil
}

// MarshalJSON encodes our enum in json
func (p *Prefix) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// New returns a prefix uuid
func New(p Prefix) string {
	return p.String() + "_" + xid.New().String()
}

// NewWithTime returns a prefixed uuid based on the time provided
func NewWithTime(p Prefix, t time.Time) string {
	return p.String() + "_" + xid.NewWithTime(t).String()
}

// GetPrefix parses a uuid and returns a prefix
func GetPrefix(uuid string) (Prefix, error) {
	const op errors.Op = "GetPrefix"
	exploded := strings.Split(uuid, "_")
	exploded = exploded[:len(exploded)-1]
	recombined := strings.Join(exploded, "_")
	p, ok := prefixToID[recombined]
	if !ok {
		return p, errors.E(pkg, op, fmt.Errorf("Prefix: %v not found", recombined))
	}
	return p, nil
}

// HasPrefix returns whether the provided uuid has the provided prefix or not
func HasPrefix(uuid string, p Prefix) (bool, error) {
	const op errors.Op = "HasPrefix"
	providedPrefix, err := GetPrefix(uuid)
	if err != nil {
		return false, errors.E(pkg, op, err)
	}
	if p == providedPrefix {
		return true, nil
	}
	return false, nil
}
