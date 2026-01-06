package constructor

func (g *Generator) makeGetSet() {
	var getList []string
	var setList []string
	for _, f := range g.fields {
		if f.isGet {
			getList = append(getList, f.name)
		}
		if f.isSet {
			setList = append(setList, f.name)
		}
	}

	g.data.GetterList = getList
	g.data.SetterList = setList
	g.data.GetSet = len(getList) > 0 || len(setList) > 0
}
