package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// https://github.com/upper/db/blob/master/postgresql/custom_types.go#L43
type Metadata map[string]interface{}

// Scan satisfies the sql.Scanner interface.
func (m *Metadata) Scan(src interface{}) error {
	source, ok := src.(string)
	if !ok {
		return errors.New("Type assertion .(string) failed.")
	}

	var i interface{}
	err := json.Unmarshal([]byte(source), &i)
	if err != nil {
		return err
	}

	*m, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("Type assertion .(map[string]interface{}) failed.")
	}

	return nil
}

// Value satisfies the driver.Valuer interface.
func (m Metadata) Value() (driver.Value, error) {
	j, err := json.Marshal(m)
	return string(j), err
}

func toMetadata(m *Metadata) map[string]interface{} {
	md := make(map[string]interface{})
	for k, v := range *m {
		md[k] = v
	}
	return md
}
