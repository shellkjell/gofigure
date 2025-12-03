package fig

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

// Constants for special selectors and configuration.
const (
	// RootSelector is the special key used to select all root-level maps.
	RootSelector = "@"
	// RangeOperator is used in section macros to define integer ranges.
	RangeOperator = "..."
)

// Pre-compiled regex patterns for multiline string processing.
var (
	reLeadCloseWhitespace = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	reLeadCloseQuotes     = regexp.MustCompile(`^("""|''')|("""|''')$`)
	reInsideWhitespace    = regexp.MustCompile(`[\r\f\t \p{Zs}]{2,}`)
	reBackslashes         = regexp.MustCompile(`\\(?P<C>[^n])`)
	reNewlineWhitespace   = regexp.MustCompile(`\n +|\n`)
)

// Transformer handles the transformation of parsed Config to final output.
type Transformer struct {
	parser *Parser
}

// NewTransformer creates a new Transformer with the given parser.
func NewTransformer(parser *Parser) *Transformer {
	return &Transformer{
		parser: parser,
	}
}

// Transform takes a parsed Config and transforms it to a map.
// The transformation pipeline processes the config through several stages:
// 1. Process includes - inline %include directives
// 2. Expand sections - convert [section] syntax to nested fields
// 3. Convert child fields - flatten dot-notation to nested maps
// 4. Resolve identifiers - replace references with actual values
// 5. Process arrays - handle array syntax and sequential keys
func (t *Transformer) Transform(c *Config) (map[string]any, error) {
	pipeline := []func(*Config) *Config{
		t.processIncludes,
		t.expandSections,
		t.convertChildFields,
		t.resolveIdentifiers,
		t.processArrayFields,
		t.mergeArrayFields,
		t.convertSequentialKeysToArrays,
	}

	for _, stage := range pipeline {
		c = stage(c)
	}

	return t.buildFinalMap(c), nil
}

// Positioned is an interface for types that have a lexer position.
type Positioned interface {
	Position() lexer.Position
}

// identifier wraps a reference name for deferred resolution.
type identifier struct {
	name *string
}

// --- Identifier Resolution ---

// findInFields searches for a key in a list of fields.
func findInFields(key string, fields []*Field) (*Value, bool) {
	for _, field := range fields {
		if field.Key == key {
			return field.Value, true
		}
	}
	return nil, false
}

// findIdentifierInConfig traverses the config to find a value by dot-separated path.
func findIdentifierInConfig(path string, root *Config) (*Value, bool) {
	parts := strings.Split(path, ".")

	for _, entry := range root.Entries {
		if entry.Field == nil {
			continue
		}
		field := entry.Field

		// Single-part path: direct key match
		if len(parts) == 1 {
			if field.Key == parts[0] {
				return field.Value, true
			}
			continue
		}

		// Multi-part path: traverse nested maps
		if field.Key != parts[0] || field.Value == nil || field.Value.Map == nil {
			continue
		}

		// Traverse remaining path segments
		value := field.Value
		found := true
		for i := 1; i < len(parts) && found; i++ {
			if value.Map == nil {
				found = false
				break
			}
			value, found = findInFields(parts[i], value.Map)
		}

		if found {
			return value, true
		}
	}
	return nil, false
}

// resolveIdentifiersInValues resolves identifier references in a slice of values.
func (t *Transformer) resolveIdentifiersInValues(values []*Value, root *Config) {
	for i, value := range values {
		if value == nil {
			continue
		}
		switch {
		case value.Map != nil:
			t.resolveIdentifiersInFields(value.Map, root)
		case value.ParsedArray != nil:
			t.resolveIdentifiersInValues(value.ParsedArray, root)
		case value.Identifier != nil:
			if resolved, ok := findIdentifierInConfig(*value.Identifier, root); ok {
				values[i] = resolved
			}
		}
	}
}

// resolveIdentifiersInFields resolves identifier references in a slice of fields.
func (t *Transformer) resolveIdentifiersInFields(fields []*Field, root *Config) {
	for _, field := range fields {
		if field.Value == nil {
			continue
		}
		switch {
		case field.Value.Map != nil:
			t.resolveIdentifiersInFields(field.Value.Map, root)
		case field.Value.ParsedArray != nil:
			t.resolveIdentifiersInValues(field.Value.ParsedArray, root)
		case field.Value.Identifier != nil:
			if resolved, ok := findIdentifierInConfig(*field.Value.Identifier, root); ok {
				field.Value = resolved
			}
		}
	}
}

// resolveIdentifiers resolves all identifier references in the config.
func (t *Transformer) resolveIdentifiers(c *Config) *Config {
	for i, entry := range c.Entries {
		if entry.Field == nil || entry.Field.Value == nil {
			continue
		}

		// Only look at entries up until this point when searching for identifiers
		partialConfig := &Config{Entries: c.Entries[:i+1]}
		value := entry.Field.Value

		switch {
		case value.Map != nil:
			t.resolveIdentifiersInFields(value.Map, partialConfig)
		case value.ParsedArray != nil:
			t.resolveIdentifiersInValues(value.ParsedArray, partialConfig)
		case value.Identifier != nil:
			if resolved, ok := findIdentifierInConfig(*value.Identifier, partialConfig); ok {
				entry.Field.Value = resolved
			}
		}
	}
	return c
}

// toFinalValue converts a Value to its final any representation.
func (v *Value) toFinalValue() any {
	switch {
	case v.Identifier != nil:
		return &identifier{v.Identifier}
	case v.Map != nil:
		return v.mapToInterface()
	case v.Float != nil:
		return v.Float
	case v.Integer != nil:
		return v.Integer
	case v.Boolean != nil:
		return v.Boolean
	case v.String != nil:
		return v.String
	case v.MultilineString != nil:
		return v.MultilineString.transform()
	case v.FinalArray != nil:
		return v.arrayToInterface(v.FinalArray)
	case v.ParsedArray != nil:
		return v.arrayToInterface(v.ParsedArray)
	default:
		return map[string]any{}
	}
}

// mapToInterface converts map fields to a Go map.
func (v *Value) mapToInterface() map[string]any {
	result := make(map[string]any, len(v.Map))
	for _, field := range v.Map {
		if field.Value == nil {
			result[field.Key] = nil
		} else {
			result[field.Key] = field.Value.toFinalValue()
		}
	}
	return result
}

// arrayToInterface converts array values to a Go slice.
func (v *Value) arrayToInterface(values []*Value) []any {
	result := make([]any, len(values))
	for i, val := range values {
		if val != nil {
			result[i] = val.toFinalValue()
		}
	}
	return result
}

// --- Map Merging ---

// mergeMaps recursively merges src into dst.
func mergeMaps(dst, src map[string]any) {
	for key, val := range src {
		// Handle @ selector for child level selection
		if key == RootSelector {
			if srcMap, ok := val.(map[string]any); ok {
				for _, root := range dst {
					if rootMap, ok := root.(map[string]any); ok {
						mergeMaps(rootMap, srcMap)
					}
				}
			}
			continue
		}

		existing, exists := dst[key]
		if !exists {
			dst[key] = val
			continue
		}

		// If both are maps, merge recursively
		srcMap, srcIsMap := val.(map[string]any)
		dstMap, dstIsMap := existing.(map[string]any)
		if srcIsMap && dstIsMap {
			mergeMaps(dstMap, srcMap)
			continue
		}

		// Otherwise, overwrite
		dst[key] = val
	}
}

// mergeValues merges two Values, preferring fields from dom.
func mergeValues(dom, sub *Value) *Value {
	if dom.Map == nil || sub.Map == nil {
		return dom
	}

	newValue := &Value{Pos: dom.Pos}
	fieldMap := make(map[string]*Field)

	// Start with sub fields
	for _, f := range sub.Map {
		fieldMap[f.Key] = f
	}

	// Overwrite with dom fields
	for _, f := range dom.Map {
		fieldMap[f.Key] = f
	}

	// Convert back to slice maintaining order (sub first, then dom additions)
	newMap := make([]*Field, 0, len(fieldMap))
	seen := make(map[string]bool)

	for _, f := range sub.Map {
		newMap = append(newMap, fieldMap[f.Key])
		seen[f.Key] = true
	}
	for _, f := range dom.Map {
		if !seen[f.Key] {
			newMap = append(newMap, f)
		}
	}

	newValue.Map = newMap
	return newValue
}

// --- Array Processing ---

// isSequentialKeys checks if all field keys are sequential integers starting from 0.
func isSequentialKeys(fields []*Field) bool {
	if len(fields) == 0 {
		return false
	}
	var lastIndex int
	for _, f := range fields {
		idx, err := strconv.Atoi(f.Key)
		if err != nil {
			return false
		}
		if idx < lastIndex || (idx != 0 && lastIndex == 0 && idx != lastIndex) {
			return false
		}
		lastIndex = idx
	}
	return true
}

// maxKeyValue returns the largest integer key in the fields.
func maxKeyValue(fields []*Field) int {
	var max int
	for _, f := range fields {
		if idx, err := strconv.Atoi(f.Key); err == nil && idx > max {
			max = idx
		}
	}
	return max
}

// processArrayFields processes fields for array conversion.
func (t *Transformer) processArrayFields(c *Config) *Config {
	ret := &Config{Entries: make([]*Entry, len(c.Entries))}
	for i, entry := range c.Entries {
		if entry.Field == nil {
			ret.Entries[i] = entry
			continue
		}
		ret.Entries[i] = &Entry{Pos: entry.Pos, Field: entry.Field.processForArrays()}
	}
	return ret
}

func (f *Field) processForArrays() *Field {
	if f.Value == nil {
		return f
	}
	return &Field{
		Pos:        f.Pos,
		Key:        f.Key,
		ArrayIndex: f.ArrayIndex,
		Value:      f.Value.processForArrays(),
	}
}

func (v *Value) processForArrays() *Value {
	if v.Map == nil {
		return v
	}
	ret := &Value{Pos: v.Pos}
	ret.Map = make([]*Field, len(v.Map))
	for i, f := range v.Map {
		ret.Map[i] = f.processForArrays()
	}
	return ret
}

// mergeArrayFields handles array field merging.
func (t *Transformer) mergeArrayFields(c *Config) *Config {
	ret := &Config{Entries: make([]*Entry, 0, len(c.Entries))}
	for _, entry := range c.Entries {
		if newEntry := entry.mergeWithConfig(ret); newEntry != nil {
			ret.Entries = append(ret.Entries, newEntry)
		}
	}
	return ret
}

func (e *Entry) mergeWithConfig(config *Config) *Entry {
	if e.Field == nil {
		return e
	}
	if newField := e.Field.mergeWithConfig("", config); newField != nil {
		return &Entry{Pos: e.Pos, Field: newField}
	}
	return nil
}

func (f *Field) mergeWithConfig(prefix string, config *Config) *Field {
	if f.ArrayIndex != nil {
		val, ok := findIdentifierInConfig(prefix, config)
		if !ok || val.Map == nil {
			return f
		}

		indexStr := strconv.Itoa(int(*f.ArrayIndex))
		for _, mapVal := range val.Map {
			if mapVal.Key == indexStr {
				mapVal.Value = mergeValues(f.Value, mapVal.Value)
				return nil
			}
		}
		return f
	}

	if f.Value == nil {
		return f
	}

	newPrefix := f.Key
	if prefix != "" {
		newPrefix = prefix + "." + f.Key
	}

	if f.Value.Map == nil {
		return &Field{Pos: f.Pos, Value: f.Value, Key: f.Key}
	}

	newMap := make([]*Field, 0, len(f.Value.Map))
	for _, mapVal := range f.Value.Map {
		if newField := mapVal.mergeWithConfig(newPrefix, config); newField != nil {
			newMap = append(newMap, newField)
		}
	}

	if len(newMap) == 0 {
		return nil
	}

	return &Field{
		Pos:   f.Pos,
		Key:   f.Key,
		Value: &Value{Pos: f.Value.Pos, Map: newMap},
	}
}

// convertSequentialKeysToArrays converts maps with sequential integer keys to arrays.
func (t *Transformer) convertSequentialKeysToArrays(c *Config) *Config {
	ret := &Config{Entries: make([]*Entry, len(c.Entries)), Pos: c.Pos}
	for i, entry := range c.Entries {
		if entry.Field == nil {
			ret.Entries[i] = entry
			continue
		}
		ret.Entries[i] = &Entry{Pos: entry.Pos, Field: entry.Field.convertSequentialKeys()}
	}
	return ret
}

func (f *Field) convertSequentialKeys() *Field {
	if f.Value == nil || f.Value.Map == nil {
		return &Field{Pos: f.Pos, Key: f.Key, Value: f.Value}
	}

	newField := &Field{Pos: f.Pos, Key: f.Key}
	newValue := &Value{Pos: f.Value.Pos}

	// First, recursively process children
	processedMap := make([]*Field, len(f.Value.Map))
	for i, mapVal := range f.Value.Map {
		processedMap[i] = mapVal.convertSequentialKeys()
	}

	// Then check if this level should become an array
	if isSequentialKeys(processedMap) {
		arraySize := maxKeyValue(processedMap) + 1
		newValue.FinalArray = make([]*Value, arraySize)
		for _, mapVal := range processedMap {
			if idx, err := strconv.Atoi(mapVal.Key); err == nil {
				newValue.FinalArray[idx] = mapVal.Value
			}
		}
	} else {
		newValue.Map = processedMap
	}

	newField.Value = newValue
	return newField
}

// --- Child Field Conversion ---

// convertChildFields flattens chained child fields into nested maps.
func (t *Transformer) convertChildFields(c *Config) *Config {
	ret := &Config{Entries: make([]*Entry, len(c.Entries))}
	for i, entry := range c.Entries {
		if entry.Field == nil {
			ret.Entries[i] = entry
			continue
		}

		field := entry.Field
		for field.Child != nil {
			field.Value = &Value{
				Map: []*Field{{
					ArrayIndex: field.Child.ArrayIndex,
					Child:      field.Child.Child,
					Key:        field.Child.Key,
					Value:      field.Child.Value,
					Pos:        field.Child.Pos,
				}},
			}
			field.Child = nil
			field = field.Value.Map[0]
		}
		ret.Entries[i] = entry
	}
	return ret
}

// --- Section Expansion ---

// expandSections expands section definitions into fields.
func (t *Transformer) expandSections(c *Config) *Config {
	var result []*Entry

	for _, entry := range c.Entries {
		if entry.Section == nil {
			result = append(result, entry)
			continue
		}

		for _, field := range entry.Section.expandToFields() {
			result = append(result, &Entry{Field: field})
		}
	}

	return &Config{Entries: result}
}

func (s *Section) expandToFields() []*Field {
	var result []*Field
	for _, root := range s.Roots {
		result = append(result, root.expandToFields(s.Fields)...)
	}
	return result
}

func (s *SectionRoot) expandToFields(setTo []*Field) []*Field {
	childFields := setTo
	if s.Child != nil {
		childFields = s.Child.expandToFields(setTo)
	}

	result := make([]*Field, 0, len(s.Identifier))
	for _, name := range s.Identifier {
		result = append(result, &Field{
			Key:   name,
			Value: &Value{Map: childFields},
		})
	}
	return result
}

func (s *SectionChild) expandToFields(setTo []*Field) []*Field {
	childFields := setTo
	if s.Child != nil {
		childFields = s.Child.expandToFields(setTo)
	}

	// Handle integer range macro (e.g., %{1...5})
	s.expandRangeMacro()

	result := make([]*Field, 0, len(s.Identifier))
	for _, name := range s.Identifier {
		result = append(result, &Field{
			Key:   name,
			Value: &Value{Map: childFields},
		})
	}
	return result
}

// expandRangeMacro expands range macros like %{1...5} into [1,2,3,4,5].
func (s *SectionChild) expandRangeMacro() {
	if len(s.Identifier) != 3 || s.Identifier[1] != RangeOperator {
		return
	}

	start, err1 := strconv.Atoi(s.Identifier[0])
	end, err2 := strconv.Atoi(s.Identifier[2])
	if err1 != nil || err2 != nil {
		return
	}

	length := (end - start) + 1
	s.Identifier = make([]string, length)
	for i := range s.Identifier {
		s.Identifier[i] = strconv.Itoa(start + i)
	}
}

// --- Include Processing ---

// processIncludes handles %include directives by parsing and merging included files.
func (t *Transformer) processIncludes(c *Config) *Config {
	var result []*Entry

	for _, entry := range c.Entries {
		if entry.Include == nil {
			result = append(result, entry)
			continue
		}

		for _, filename := range entry.Include.Includes {
			included, err := t.parser.ParseFile(filename)
			if err != nil {
				// TODO: Consider adding error collection instead of silent skip
				continue
			}
			included = t.processIncludes(included)
			result = append(result, included.Entries...)
		}
	}

	return &Config{Pos: c.Pos, Entries: result}
}

// --- Final Map Building ---

// buildFinalMap converts the processed Config to the final map representation.
func (t *Transformer) buildFinalMap(c *Config) map[string]any {
	result := make(map[string]any)

	for _, entry := range c.Entries {
		if entry.Field == nil {
			continue
		}
		field := entry.Field

		var finalValue any
		if field.Value != nil {
			finalValue = field.Value.toFinalValue()
		}

		// Handle @ selector for all roots
		if field.Key == RootSelector {
			applyToAllRoots(result, finalValue)
			continue
		}

		// Regular field handling
		addFieldToResult(result, field.Key, field.Value, finalValue)
	}

	return result
}

// applyToAllRoots applies a value to all root-level maps.
func applyToAllRoots(result map[string]any, value any) {
	valueMap, ok := value.(map[string]any)
	if !ok {
		return
	}
	for _, root := range result {
		if rootMap, ok := root.(map[string]any); ok {
			mergeMaps(rootMap, valueMap)
		}
	}
}

// addFieldToResult adds a field to the result map, merging if necessary.
func addFieldToResult(result map[string]any, key string, value *Value, finalValue any) {
	if value == nil {
		result[key] = nil
		return
	}

	existing, exists := result[key]
	if !exists {
		result[key] = finalValue
		return
	}

	// Merge maps, otherwise keep existing (first wins for non-maps)
	valueMap, valueIsMap := finalValue.(map[string]any)
	existingMap, existingIsMap := existing.(map[string]any)
	if valueIsMap && existingIsMap {
		mergeMaps(existingMap, valueMap)
	}
}

// --- String Processing ---

// transform processes a multiline string, normalizing whitespace and escapes.
func (s *UnprocessedString) transform() string {
	final := reLeadCloseWhitespace.ReplaceAllString(*s.String, "")
	final = reLeadCloseQuotes.ReplaceAllString(final, "")
	final = reInsideWhitespace.ReplaceAllString(final, " ")
	final = reBackslashes.ReplaceAllString(final, `\\$C`)
	final = reNewlineWhitespace.ReplaceAllString(final, `\\n`)
	return final
}
