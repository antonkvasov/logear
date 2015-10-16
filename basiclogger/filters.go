package basiclogger

import (
	"encoding/json"
	"errors"
	"gopkg.in/vmihailenco/msgpack.v2"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Filter struct {
	regexp   *regexp.Regexp //For regexp type of filter
	template string         // Template for parsed data
}

var filters map[string]Filter

func AddFilter(conf map[string]interface{}) {
	name, nameok := conf["name"]
	regexpstr, regexpok := conf["regexp"]
	template, templateok := conf["template"]
	if !nameok || !regexpok || !templateok {
		log.Fatal("[ERROR] Please specify all field of custom filter")
	}
	r, err := regexp.Compile(regexpstr.(string))
	if err != nil {
		log.Fatal("[ERROR] Incorrect regular expression")
	}
	filters[name.(string)] = Filter{regexp: r, template: template.(string)}
	log.Printf("[DEBUG] [filters] \"%s\" Filter added", name.(string))
}

func FilterData(name, data string, m *map[string]interface{}) error {
	switch name {
	case "json":
		err := json.Unmarshal([]byte(data), &m)
		return err
	case "msgpack":
		err := msgpack.Unmarshal([]byte(data), &m)
		return err
	default:
		if f, ok := filters[name]; ok {
			if f.regexp != nil && len(f.template) > 0 {
				matches := f.regexp.FindStringSubmatch(data)
				j := f.template
				for i, match := range matches {
					escaped := url.QueryEscape(match)
					j = strings.Replace(j, "$("+strconv.Itoa(i)+")", escaped, -1)
				}
				err := json.Unmarshal([]byte(j), &m)
				return err
			} else {
				return errors.New("Regexp filter error: " + name)
			}
		} else {
			return errors.New("Unknown filter type \"" + name + "\"")
		}
	}
	return errors.New("Unknown error")
}