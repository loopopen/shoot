package constructor

func (g *Generator) makeJson() {
	if !g.flags.json {
		return
	}

	tagMap := make(map[string]string)
	for _, f := range g.fields {
		if f.isShadowed || f.isEmbeded {
			continue
		}
		if f.jsonTag != "" {
			tagMap[f.name] = f.jsonTag
		}
	}
	g.data.JSONTagMap = tagMap
	g.data.JSON = true
}
