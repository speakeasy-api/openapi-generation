package extensions

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func (d *cliManifestDecoder) catalogDefaultsForPresets(presets []CLICommandPreset, variant *cliResolvedSchema) ([]CLICatalogDefault, error) {
	var defaults []CLICatalogDefault
	for _, preset := range presets {
		schema, found := variant.properties[cliPointerPropertyName(preset.Bind.Pointer)]
		if !found {
			continue
		}
		facts, err := d.propertyFacts(schema, nil)
		if err != nil {
			return nil, err
		}
		values, err := d.catalogDefaultsForValue(preset.Value, facts)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if !slices.Contains(defaults, value) {
				defaults = append(defaults, value)
			}
		}
	}
	return defaults, nil
}

func (d *cliManifestDecoder) catalogDefaultsForValue(value any, facts *cliPropertyFacts) ([]CLICatalogDefault, error) {
	if len(facts.unionArms) > 0 {
		var defaults []CLICatalogDefault
		matched := false
		for _, arm := range facts.unionArms {
			armFacts, err := d.propertyFacts(arm, nil)
			if err != nil {
				return nil, err
			}
			if d.checkPresetValueStrict("", "", value, armFacts) != nil {
				continue
			}
			candidate, err := d.catalogDefaultsForValue(value, armFacts)
			if err != nil {
				return nil, err
			}
			if len(candidate) == 0 {
				continue
			}
			if matched && !reflect.DeepEqual(defaults, candidate) {
				return nil, nil
			}
			defaults = candidate
			matched = true
		}
		return defaults, nil
	}

	if items, ok := value.([]any); ok && facts.items != nil {
		var defaults []CLICatalogDefault
		for _, item := range items {
			values, err := d.catalogDefaultsForValue(item, facts.items)
			if err != nil {
				return nil, err
			}
			for _, value := range values {
				if !slices.Contains(defaults, value) {
					defaults = append(defaults, value)
				}
			}
		}
		return defaults, nil
	}

	for _, member := range facts.enum {
		if member == nil || !cliValuesEqual(value, member) {
			continue
		}
		var defaults []CLICatalogDefault
		for _, command := range facts.catalogCommands {
			defaults = append(defaults, CLICatalogDefault{CatalogCommand: command, Value: fmt.Sprint(member)})
		}
		return defaults, nil
	}
	return nil, nil
}

func cliLabelCatalogDefaults(cmd *CLICommand, routes [][]CLICatalogDefault) []CLICatalogDefault {
	var defaults []CLICatalogDefault
	command := strings.Join(cmd.Path, " ")
	for i, values := range routes {
		for _, value := range values {
			shared := true
			for _, other := range routes {
				if !slices.Contains(other, value) {
					shared = false
					break
				}
			}
			value.Label = command
			if !shared {
				value.Label += " --" + cmd.Source.Routes[i].Selector
			}
			if !slices.Contains(defaults, value) {
				defaults = append(defaults, value)
			}
		}
	}
	return defaults
}
