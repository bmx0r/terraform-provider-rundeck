package rundeck

import (
	"encoding/xml"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
)

func validateValueFunc(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(string)
		valid := false
		for _, role := range values {
			if value == role {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%s is an invalid value for argument %s", value, k))
		}
		return
	}
}

func marshalJobMapToXML(c *map[string]string, e *xml.Encoder, start xml.StartElement, entryName string, keyName string, valueName string) error {
	if len(*c) == 0 {
		return nil
	}
	e.EncodeToken(start)

	// Sort the keys so we'll have a deterministic result.
	keys := []string{}
	for k, _ := range *c {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := (*c)[k]
		e.EncodeToken(xml.StartElement{
			Name: xml.Name{Local: entryName},
			Attr: []xml.Attr{
				xml.Attr{
					Name:  xml.Name{Local: keyName},
					Value: k,
				},
				xml.Attr{
					Name:  xml.Name{Local: valueName},
					Value: v,
				},
			},
		})
		e.EncodeToken(xml.EndElement{xml.Name{Local: entryName}})
	}
	e.EncodeToken(xml.EndElement{start.Name})
	return nil
}

func unmarshalJobMapFromXML(c *map[string]string, d *xml.Decoder, start xml.StartElement, entryName string, keyName string, valueName string) error {
	result := map[string]string{}
	for {
		token, err := d.Token()
		if token == nil {
			err = fmt.Errorf("EOF while decoding job command plugin config")
		}
		if err != nil {
			return err
		}

		switch t := token.(type) {
		default:
			continue
		case xml.StartElement:
			if t.Name.Local != entryName {
				return fmt.Errorf("unexpected element %s while looking for config entries", t.Name.Local)
			}
			var k string
			var v string
			for _, attr := range t.Attr {
				if attr.Name.Local == keyName {
					k = attr.Value
				} else if attr.Name.Local == valueName {
					v = attr.Value
				}
			}
			if k == "" {
				return fmt.Errorf("found config entry with empty key")
			}
			result[k] = v
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				*c = result
				return nil
			}
		}
	}
}

func marshalLogMapToXML(c *map[string]string, e *xml.Encoder, start xml.StartElement) error {
	if len(*c) == 0 {
		return nil
	}
	e.EncodeToken(start)

	// Sort the keys so we'll have a deterministic result.
	keys := []string{}
	for k, _ := range *c {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := (*c)[k]
		e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
		e.EncodeToken(xml.CharData([]byte(v)))
		e.EncodeToken(xml.EndElement{xml.Name{Local: k}})
	}
	e.EncodeToken(xml.EndElement{start.Name})
	return nil
}

func unmarshalLogMapFromXML(c *map[string]string, d *xml.Decoder, start xml.StartElement) error {
	result := map[string]string{}
	for {
		token, err := d.Token()
		if token == nil {
			err = fmt.Errorf("EOF while decoding plugin config")
		}
		if err != nil {
			return err
		}

		var k string
		var v string
		switch t := token.(type) {
		default:
			continue
		case xml.StartElement:
			k = t.Name.Local
		case xml.CharData:
			v = string(xml.CharData(t))
			result[k] = v
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				*c = result
				return nil
			}
		}
	}
}
