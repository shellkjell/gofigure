package main

import (
	"errors"
	"regexp"
	"strconv"
)

func lookupIdentifierInRoot(multiKeyName *string, root map[string]interface{}) (interface{}, error) {
	keyNames := splitIdentifiers(*multiKeyName)

	var currRoot interface{}
	currRoot = root
	for _, keyName := range keyNames {
		if _, exists := currRoot.(map[string]interface{})[keyName]; !exists {
			return nil, errors.New("No key with the name \"" + keyName + "\" exists from query: \"" + *multiKeyName + "\"")
		}

		currRoot = currRoot.(map[string]interface{})[keyName]
	}

	return currRoot, nil
}

func isValidFinalValue(val interface{}) bool {
	if val != nil { // nil value is always valid
		switch val.(type) {
		case *string:
		case *int64:
		case *float64:
		case map[string]interface{}:
			break
		default:
			return false
		}
	}

	return true
}

func checkConfigError(err error, v *Value) {
	if err != nil {
		panic(err.Error() + " in " + v.Pos.Filename + ":" + strconv.FormatInt(int64(v.Pos.Line), 10) + ":" + strconv.FormatInt(int64(v.Pos.Column), 10))
	}
}

func (v *Value) toFinalValue(root map[string]interface{}) (ret interface{}) {
	if v.Identifier != nil { // First look in local scope
		valAtIdent, err := lookupIdentifierInRoot(v.Identifier, root)

		if err != nil && &root != &globalRoot { // If not in local scope then look in global scope
			valAtIdent, err = lookupIdentifierInRoot(v.Identifier, globalRoot)
		}

		checkConfigError(err, v)

		if !isValidFinalValue(valAtIdent) {
			panic("Value at identifier \"" + *v.Identifier + "\" is not a valid value")
		}

		ret = valAtIdent
	} else if v.Map != nil {
		nwMap := map[string]interface{}{}
		// expand roots?
		for _, field := range v.Map {
			if field.Value == nil {
				delete(nwMap, field.Key)
			} else {
				nwMap[field.Key] = field.Value.toFinalValue(nwMap)
			}
		}

		ret = nwMap
	} else if v.List != nil {
		nwList := make([]interface{}, len(v.List), len(v.List))
		for i, val := range v.List {
			nwList[i] = val.toFinalValue(root)
		}

		ret = nwList
	} else if v.Float != nil {
		ret = v.Float
	} else if v.Integer != nil {
		ret = v.Integer
	} else if v.String != nil {
		ret = v.String
	}

	return
}

func mergeMapsOfInterface(dst, src map[string]interface{}) {
	for key, val := range src {
		// Child level select all roots
		if key == "@" {
			for _, root := range dst {
				switch root.(type) {
				case map[string]interface{}:
					mergeMapsOfInterface(root.(map[string]interface{}), val.(map[string]interface{}))
					break
				default:
					// Nothing?
				}
			}

			continue
		}

		if _, exists := dst[key]; !exists {
			dst[key] = val
			continue
		}

		switch src[key].(type) {
		case map[string]interface{}:
			switch dst[key].(type) {
			case map[string]interface{}:
				mergeMapsOfInterface(dst[key].(map[string]interface{}), src[key].(map[string]interface{}))

				break

			default:
				dst[key] = val
			}

			break

		default:
			dst[key] = val
		}
	}
}

// Transform - Takes a freshly parsed config file and transforms it to a map
func (c *CONFIG) Transform() map[string]interface{} {
	c = c.parseIncludesAndAppendToConfig()
	c = c.explodeSectionsToFields()
	c = c.childFieldsToMap()

	return c.toMap()
}

var globalRoot map[string]interface{}

func (c *CONFIG) toMap() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	globalRoot = ret

	for _, entry := range c.Entries {
		field := entry.Field
		value := field.Value

		var finalValue interface{}

		if value != nil {
			finalValue = value.toFinalValue(ret)
		}

		// Top level selection of all roots
		if field.Key == "@" {
			var processRoots []map[string]interface{}

			for _, val := range ret {
				switch val.(type) {
				case map[string]interface{}:
					processRoots = append(processRoots, val.(map[string]interface{}))
					break

				default:
					// do nothing?
				}
			}

			for _, root := range processRoots {
				mergeMapsOfInterface(root, finalValue.(map[string]interface{}))
			}

			continue
		}

		// Otherwise regular value
		if value != nil {
			if _, exists := ret[field.Key]; !exists {
				ret[field.Key] = finalValue
			} else { // Only maps are Reassigns == false, change to better name?
				mergeMapsOfInterface(ret[field.Key].(map[string]interface{}), finalValue.(map[string]interface{}))
			}
		} else {
			ret[field.Key] = nil
		}
	}

	// Remove empty keys
	for key, val := range ret {
		if val == nil {
			delete(ret, key)
		}
	}

	return
}

func (c *CONFIG) childFieldsToMap() (ret *CONFIG) {
	ret = &CONFIG{}
	ret.Entries = make([]*Entry, len(c.Entries))

	for i, entry := range c.Entries {
		if entry.Field == nil {
			ret.Entries[i] = entry
			continue
		}

		currField := entry.Field
		for currField.Child != nil {
			currField.Value = &Value{Map: []*Field{currField.Child}}
			currField.Child = nil
			currField = currField.Value.Map[0]
		}

		ret.Entries[i] = entry
	}

	return
}

type expansionSection struct {
	Prefix []string
	Rest   []string
	Fields []*Field
}

func transformToExpansionSection(sectRoot SectionRoot) (retVal expansionSection) {
	currSect := sectRoot.Child
	retVal.Rest = append(retVal.Rest, *sectRoot.Identifier)

	if sectRoot.Fields != nil {
		retVal.Fields = sectRoot.Fields
		return
	}

	for currSect != nil {
		retVal.Rest = append(retVal.Rest, *currSect.Identifier)

		if currSect.Fields != nil {
			retVal.Fields = currSect.Fields
			break
		}

		currSect = currSect.Child
	}

	return
}

func transformSectionToSectionRoot(expandedSects []expansionSection) (retVal []SectionRoot) {
	for _, expandedSect := range expandedSects {
		nwRootName := expandedSect.Prefix[0]
		newSectRoot := SectionRoot{Identifier: &nwRootName}

		if len(expandedSect.Prefix) > 1 {
			newSectRoot.Child = &SectionChild{}
		} else {
			newSectRoot.Fields = expandedSect.Fields
			retVal = append(retVal, newSectRoot)
			continue
		}

		currSectChild := newSectRoot.Child

		for i := 1; i < len(expandedSect.Prefix); i++ {
			sectChildName := expandedSect.Prefix[i]
			currSectChild.Identifier = &sectChildName

			if i != len(expandedSect.Prefix)-1 {
				currSectChild.Child = &SectionChild{}
				currSectChild = currSectChild.Child
			}
		}

		currSectChild.Fields = expandedSect.Fields
		retVal = append(retVal, newSectRoot)
	}

	return
}

func expand(sectRoot SectionRoot) (retVal []SectionRoot) {
	nwExpSect := transformToExpansionSection(sectRoot)

	expanded := _expandInternal([]expansionSection{nwExpSect}, []expansionSection{})

	return transformSectionToSectionRoot(expanded)
}

func _expandInternal(toProcess, result []expansionSection) (retVal []expansionSection) {
	if len(toProcess) == 0 {
		return result
	}

	first := toProcess[0]
	rest := toProcess[1:]

	expandedItems := expandSection(first)
	for _, expandedItem := range expandedItems {
		if len(expandedItem.Rest) == 0 {
			result = append(result, expandedItem)
		} else {
			rest = append(rest, expandedItem)
		}
	}

	return _expandInternal(rest, result)
}

func expandSection(val expansionSection) (retVal []expansionSection) {
	if len(val.Rest) != 0 {
		element := val.Rest[0]

		if isValidExpansionMacro(element) {
			expansionMacros := splitExpansionMacro(element)

			for _, expansionMacro := range expansionMacros {
				nwPrefix := make([]string, len(val.Prefix))
				copy(nwPrefix, val.Prefix)

				nwPrefix = append(nwPrefix, expansionMacro)
				nwSect := expansionSection{
					Prefix: nwPrefix,
					Fields: val.Fields,
					Rest:   val.Rest[1:],
				}

				retVal = append(retVal, nwSect)
			}
		} else {
			newSect := expansionSection{Prefix: append(val.Prefix, element), Fields: val.Fields}
			newSect.Rest = val.Rest[1:]
			retVal = append(retVal, newSect)
		}
	} else {
		retVal = append(retVal, val)

		return
	}

	return
}

func (c *CONFIG) explodeSectionsToFields() (ret *CONFIG) {
	ret = &CONFIG{}
	ret.Entries = make([]*Entry, len(c.Entries))

	for i, newEntriesIndex := 0, 0; i < len(c.Entries); i++ {
		entry := c.Entries[i]
		if entry.Section == nil {
			ret.Entries[newEntriesIndex] = entry
			newEntriesIndex++
			continue
		}

		section := entry.Section

		newSections := expand(*section)

		newEntries := []*Entry{}

		for _, nwSection := range newSections {
			// expand all sections to values
			childSection := nwSection.Child

			rootField := &Field{Key: *nwSection.Identifier, Pos: entry.Pos}
			currField := rootField

			if childSection == nil {
				if len(nwSection.Fields) != 0 {
					rootField.Value = &Value{Map: nwSection.Fields}
				}

				newEntries = append(newEntries, &Entry{Field: rootField})
				continue
			}

			currField.Value = &Value{Map: []*Field{&Field{Key: *childSection.Identifier, Pos: entry.Pos}}, Pos: entry.Pos}
			currField = currField.Value.Map[0]

			for childSection.Child != nil {
				childSection = childSection.Child

				currField.Value = &Value{Map: []*Field{&Field{Key: *childSection.Identifier, Pos: entry.Pos}}, Pos: entry.Pos}
				currField = currField.Value.Map[0]
			}

			// Make a new value for the field
			newValue := &Value{}
			newValue.Map = childSection.Fields

			currField.Value = newValue

			newEntries = append(newEntries, &Entry{Field: rootField, Pos: entry.Pos})
		}

		firstSlice := ret.Entries[:newEntriesIndex]
		var lastSlice []*Entry

		if newEntriesIndex < len(ret.Entries)-1 {
			lastSlice = ret.Entries[newEntriesIndex+1:]
		}

		ret.Entries = append(firstSlice, newEntries...)
		ret.Entries = append(ret.Entries, lastSlice...)

		newEntriesIndex += len(newEntries)
	}

	return
}

func (c *CONFIG) parseIncludesAndAppendToConfig() (ret *CONFIG) {
	ret = &CONFIG{}
	ret.Entries = make([]*Entry, len(c.Entries))

	for i, newEntriesIndex := 0, 0; i < len(c.Entries); i++ {
		entry := c.Entries[i]

		if entry.Include == nil {
			ret.Entries[newEntriesIndex] = entry
			newEntriesIndex++
			continue
		}

		include := entry.Include
		newConfigList := make([]*CONFIG, len(include.Includes))

		parser := BuildParser()

		// Parse includes
		for j, includeName := range include.Includes {
			newConfig := ParseFile(includeName, parser)

			newConfig = newConfig.parseIncludesAndAppendToConfig()

			newConfigList[j] = newConfig
		}

		// Append new entries in main config and remove the include entry
		for _, newConfig := range newConfigList {
			newEntryList := append(ret.Entries[:newEntriesIndex], newConfig.Entries...)
			newEntryList = append(newEntryList, ret.Entries[newEntriesIndex+1:]...)

			ret.Entries = newEntryList
			newEntriesIndex += len(newConfig.Entries)
		}
	}

	return
}

// This function removes leading/trailing whitespaces, string quotes etc.
func (thisArg *UnprocessedString) transform() (final string) {
	re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_leadclose_quotes := regexp.MustCompile(`^("""|''')|("""|''')$`)
	re_inside_whtsp := regexp.MustCompile(`[\r\f\t \p{Zs}]{2,}`)
	re_backslashes := regexp.MustCompile(`\\(?P<C>[^n])`)
	re_newline_whtsp := regexp.MustCompile(`\n +|\n`)

	final = re_leadclose_whtsp.ReplaceAllString(*thisArg.String, "")
	final = re_leadclose_quotes.ReplaceAllString(final, "")
	final = re_inside_whtsp.ReplaceAllString(final, " ")
	final = re_backslashes.ReplaceAllString(final, `\\$C`)
	final = re_newline_whtsp.ReplaceAllString(final, `\\n`)

	return
}
