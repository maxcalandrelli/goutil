package gu_json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	JSON_DEBUG            = false
	BUILT_ARRAY_NAME      = "@"
	BUILT_ELEMENTARY_NAME = "$"
)

type JSONError struct {
	description string
	element     string
}

var (
	StopIteration = JSONError{description: "end of iteration"}
)

func (je JSONError) Error() string {
	return fmt.Sprintf("%s accessing '%s'", je.description, je.element)
}

type JSONData map[string]interface{}

func (data JSONData) SetParameter(name, value string) {
	if values, exists := data[name]; !exists {
		data[name] = value
	} else {
		switch values.(type) {
		case string:
			data[name] = []string{values.(string)}
		case []string:
			data[name] = append(data[name].([]string), value)
		}
	}
}

func (data JSONData) SetParameters(block string) {
	for _, p := range strings.Split(block, "&") {
		nv := strings.Split(p, "=")
		if len(nv) == 2 {
			data.SetParameter(nv[0], nv[1])
		}
	}
}

func (data JSONData) ToHTMLForm() url.Values {
	form := url.Values{}
	for n, v := range data {
		switch v.(type) {
		case map[string]interface{}:
			if data, err := json.Marshal(v); err == nil {
				form.Add(n, string(data))
			}
		default:
			form.Add(n, field_string(v))
		}
	}
	return form
}

func (data JSONData) POSTForm(finalUrl string) (*http.Request, error) {
	req, err := http.NewRequest("POST", finalUrl, strings.NewReader(data.ToHTMLForm().Encode()))
	if req != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, err
}

func (data JSONData) GetData(stream io.Reader, arrayName string) (JSONData, error) {
	var (
		ret JSONData = JSONData{}
		m   json.RawMessage
		err error
	)
	err = json.NewDecoder(stream).Decode(&m)
	if err == nil {
		aval := []interface{}{}
		err = json.Unmarshal(m, &aval)
		if err == nil {
			ret[arrayName] = aval
		} else {
			err = json.Unmarshal(m, &ret)
		}
	}
	return ret, err
}

func (data JSONData) DoPOSTExchange(finalUrl string, otherHeaders JSONData) (JSONData, error) {
	req, err := data.POSTForm(finalUrl)
	if err != nil {
		return nil, err
	}
	for name, value := range otherHeaders {
		req.Header.Add(name, field_string(value))
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		err = errors.New(fmt.Sprintf("exchange error %d on %s", resp.StatusCode, req.URL))
		return nil, err
	}
	return data.GetData(resp.Body, "_array")
}

func (data JSONData) _ExchangeJSON(finalUrl, auth string) (ret *JSONData, err error) {
	ret = new(JSONData)
	jdata, err := json.Marshal(data)
	if err != nil {
		return
	}
	form := url.Values{}
	for n, v := range data {
		switch v.(type) {
		case string:
			form.Add(n, v.(string))
		case map[string]interface{}:
			if data, err := json.Marshal(v); err == nil {
				form.Add(n, string(data))
			} else {
				return nil, err
			}
		}
	}
	if len(auth) > 0 {
		form.Add("session", auth)
	}
	req, err := http.NewRequest("POST", finalUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}
	if len(auth) > 0 {
		req.Header.Add("CMDBuild-Authorization", auth)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if JSON_DEBUG {
		encoder := json.NewEncoder(os.Stderr)
		encoder.SetIndent("REQUEST", " ")
		if err := encoder.Encode([]interface{}{req.URL, req.Header, form}); err != nil {
			panic(err)
		}
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode > 299 {
		if JSON_DEBUG {
			fmt.Fprintf(os.Stderr, "jdata: %s\nURL=%s\n", string(jdata), req.URL)
		}
		err = errors.New(fmt.Sprintf("exchange error %d on %s", resp.StatusCode, req.URL))
		return
	}
	err = json.NewDecoder(resp.Body).Decode(ret)
	if JSON_DEBUG {
		encoder := json.NewEncoder(os.Stderr)
		encoder.SetIndent("RESPONSE", " ")
		if err := encoder.Encode([]interface{}{resp.Header, ret}); err != nil {
			panic(err)
		}
	}
	return
}

func (data JSONData) String() string {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.Encode(data)
	return string(buffer.Bytes())
}

func (data JSONData) IndentedString(prefix, indent string) string {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(prefix, indent)
	encoder.Encode(data)
	return string(buffer.Bytes())
}

func matching_array_items(array []interface{}, index, rest string) interface{} {
	elemIndex, err := strconv.ParseInt(index, 10, 64)
	if err == nil {
		if elemIndex < 0 {
			elemIndex = int64(len(array)) - elemIndex
		}
		intIndex := int(elemIndex)
		if intIndex >= len(array) {
			return JSONError{description: "index too large"}
		} else if intIndex < 0 {
			return JSONError{description: "negative index"}
		}
		return json_access(array[intIndex], rest)
	}
	re := regexp.MustCompile("(([^<>=~\\s]+)\\s*(<|>|=|<=|>=|<>|~)\\s*([^<>=\\s]+))|^$]")
	is := re.FindStringSubmatch(index)
	if len(is) < 1 {
		return JSONError{description: "bad filter"}
	}
	if len(is[0]) == 0 {
		return json_access(array, rest)
	}
	ret := make([]interface{}, 0)
	for _, iface := range array {
		include := false
		if _l, _ok := iface.(map[string]interface{}); _ok {
			_l, err := JSONData(_l).Access(is[2])
			if err == nil {
				left := field_string(_l)
				op := is[3]
				right := is[4]
				switch op {
				case "~":
					sre, err := regexp.Compile("(?i)" + right)
					if err != nil {
						return JSONError{description: err.Error()}
					}
					include = sre.MatchString(left)
				case "=":
					include = left == right
				case "<>":
					include = left != right
				case "<":
					include = left < right
				case "<=":
					include = left <= right
				case ">":
					include = left > right
				case ">=":
					include = left >= right
				}
			}
		}
		if include {
			ret = append(ret, iface)
		}
	}
	return json_access(ret, rest)
}

func json_access(data interface{}, field string) interface{} {
	if len(field) == 0 {
		return data
	}
	index := strings.IndexFunc(field, func(c rune) bool { return c == '.' || c == '[' || c == ']' })
	elem := ""
	next := ""
	found := false
	value := interface{}(nil)
	switch {
	case index < 0:
		elem = field
		next = ""
	case field[index] == '.':
		if index == 0 {
			return json_access(data, field[1:])
		}
		elem = field[:index]
		next = field[index+1:]
	case field[index] == ']':
		return JSONError{description: "missing opening brace", element: field[:index]}
	case field[index] == '[':
		elem = field[:index]
		if index == 0 {
			switch data.(type) {
			case []interface{}:
				next = field[index+1:]
				array := data.([]interface{})
				closing := strings.IndexFunc(next, func(c rune) bool { return c == ']' })
				if closing < 0 {
					return JSONError{description: "missing closing brace after index"}
				}
				if closing < 1 {
					return JSONError{description: "missing index"}
				}
				return matching_array_items(array, next[:closing], next[closing+1:])
			default:
				return JSONError{description: "not an array"}
			}
		} else {
			next = field[index:]
		}
	}
	switch data.(type) {
	case map[string]interface{}:
		value, found = data.(map[string]interface{})[elem]
	case JSONData:
		value, found = data.(JSONData)[elem]
	case []interface{}:
		value := make([]interface{}, 0)
		for _, _v := range data.([]interface{}) {
			value = append(value, json_access(_v, elem))
		}
		if len(next) == 0 {
			return value
		}
		retval := make([]interface{}, 0)
		for _, _v := range value {
			retval = append(retval, json_access(_v, next))
		}
		return retval
	default:
		return JSONError{description: fmt.Sprintf("unhandled type (%T)", data), element: elem}
	}
	if !found {
		return JSONError{description: "element not found", element: elem}
	}
	if len(next) == 0 {
		return value
	}
	result := json_access(value, next)
	switch result.(type) {
	case JSONError:
		jerr := result.(JSONError)
		if len(elem) > 0 {
			if len(jerr.element) > 0 {
				jerr.element = elem + "." + jerr.element
			} else {
				jerr.element = elem
			}
		}
		return jerr
	}
	return result
}

func (data JSONData) Access(field string) (interface{}, error) {
	ret := json_access(data, field)
	switch ret.(type) {
	case error:
		return nil, ret.(error)
	case JSONError:
		return nil, error(ret.(JSONError))
	}
	return ret, nil
}

func (data JSONData) Iterate(field string, f func(JSONData) error) error {
	err := error(nil)
	ret := json_access(data, field)
	switch ret.(type) {
	case error:
		return ret.(error)
	case JSONError:
		return error(ret.(JSONError))
	case []interface{}:
		for _, v := range ret.([]interface{}) {
			switch v.(type) {
			case JSONData:
				err = f(v.(JSONData))
			case map[string]interface{}:
				err = f(JSONData(v.(map[string]interface{})))
			case []interface{}:
				return JSONError{description: "nested array", element: field}
			default:
				return JSONError{description: fmt.Sprintf("unhandle iteration on type %T", v), element: field}
			}
			if err != nil {
				if err == StopIteration {
					return nil
				}
				return err
			}
		}
		return nil
	}
	return JSONError{description: fmt.Sprintf("not an array (%T)", ret), element: field}
}

func field_string(data interface{}) string {
	switch data.(type) {
	case string:
		return data.(string)
	case float64:
		f := data.(float64)
		if f == float64(int64(f)) {
			return fmt.Sprintf("%.0f", f)
		}
	}
	return fmt.Sprintf("%#v", data)
}

func DisplayString(data interface{}) string {
	return field_string(data)
}

func (data JSONData) GetString(field string) string {
	ret, err := data.Access(field)
	if err != nil {
		return err.Error()
	}
	return field_string(ret)
}

func Build(data interface{}, e error) (JSONData, error) {
	ret := JSONData{}
	if e != nil {
		return nil, e
	}
	switch data.(type) {
	case JSONData:
		ret = data.(JSONData)
	case map[string]interface{}:
		ret = JSONData(data.(map[string]interface{}))
	case []interface{}:
		ret[BUILT_ARRAY_NAME] = data
	}
	ret[BUILT_ELEMENTARY_NAME] = fmt.Sprintf("%#v", data)
	return ret, nil
}
