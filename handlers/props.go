package handlers

import "encoding/xml"

const MaxRequestedProps = 200

func uniqueProps(props []xml.Name) (unique []xml.Name, ok bool) {
	seen := make(map[xml.Name]bool, len(props))
	unique = make([]xml.Name, 0, len(props))
	for _, prop := range props {
		if seen[prop] {
			continue
		}
		if len(unique) == MaxRequestedProps {
			return nil, false
		}
		seen[prop] = true
		unique = append(unique, prop)
	}
	return unique, true
}
